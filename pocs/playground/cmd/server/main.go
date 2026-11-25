package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/playground/internal/api"
	"github.com/diegopacheco/golang-playground/pocs/playground/internal/complete"
	"github.com/diegopacheco/golang-playground/pocs/playground/internal/config"
	"github.com/diegopacheco/golang-playground/pocs/playground/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/playground/internal/runner"
	"github.com/diegopacheco/golang-playground/pocs/playground/web"
)

func main() {
	logger := logx.New("go-playground")
	ctx := context.Background()
	if err := run(ctx, logger); err != nil {
		logger.ErrorContext(ctx, "server stopped", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	idx, err := complete.Build(cfg.GoRoot)
	if err != nil {
		return err
	}
	version, err := exec.Command(cfg.GoBin, "env", "GOVERSION").Output()
	if err != nil {
		return err
	}
	static, err := fs.Sub(web.Static, "static")
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	api.New(runner.Runner{GoBin: cfg.GoBin, Timeout: cfg.RunTimeout, MaxOutput: cfg.MaxOutput}, idx, logger, strings.TrimSpace(string(version))).Routes(mux)
	mux.Handle("GET /", http.FileServerFS(static))
	srv := &http.Server{Addr: cfg.Addr, Handler: logx.Middleware(mux), ReadHeaderTimeout: 5 * time.Second}
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() {
		logger.InfoContext(ctx, "listening", "addr", "http://"+cfg.Addr, "symbols", idx.Size())
		errCh <- srv.ListenAndServe()
	}()
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-sigCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
