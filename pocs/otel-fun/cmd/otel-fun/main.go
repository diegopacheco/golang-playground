package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/checkout"
	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/config"
	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/httpapi"
	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/telemetry"
)

func main() {
	logger := logx.New(os.Stderr)
	if err := run(logger); err != nil {
		logger.Error("service failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	prov, err := telemetry.Setup(ctx, cfg, os.Stdout)
	if err != nil {
		return fmt.Errorf("setup telemetry: %w", err)
	}

	svc, err := checkout.New(prov.Tracer, prov.Meter)
	if err != nil {
		return fmt.Errorf("create checkout service: %w", err)
	}

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           httpapi.NewRouter(svc, logger, prov),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", cfg.Addr(), "otlp_endpoint", cfg.OTLPEndpoint, "service", cfg.ServiceName)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve: %w", err)
		}
	case <-ctx.Done():
		logger.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	return errors.Join(srv.Shutdown(shutdownCtx), prov.Shutdown(shutdownCtx))
}
