package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/diegopacheco/wails-fun/internal/config"
	"github.com/diegopacheco/wails-fun/internal/db"
	"github.com/diegopacheco/wails-fun/internal/employee"
	"github.com/diegopacheco/wails-fun/internal/httpapi"
	"github.com/diegopacheco/wails-fun/internal/invoice"
	"github.com/diegopacheco/wails-fun/internal/logx"
	"github.com/diegopacheco/wails-fun/internal/receipt"
	"github.com/diegopacheco/wails-fun/internal/sale"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logger := logx.New("wails-fun")
	if err := run(logger); err != nil {
		logger.ErrorContext(context.Background(), "app failed", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	conn, err := db.Open(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	dist, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		return fmt.Errorf("assets: %w", err)
	}
	router := httpapi.NewRouter(logger, services(conn), dist)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.Port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	server := &http.Server{Handler: router, ReadHeaderTimeout: 5 * time.Second}
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()
	logger.InfoContext(ctx, "rest api listening", "port", cfg.Port, "db", cfg.DBPath, "headless", cfg.Headless)

	if cfg.Headless {
		select {
		case <-ctx.Done():
		case err := <-serveErr:
			return fmt.Errorf("serve: %w", err)
		}
	} else if err := wails.Run(&options.App{
		Title:       "Small Biz Manager",
		Width:       1240,
		Height:      820,
		AssetServer: &assetserver.Options{Assets: dist, Handler: router},
	}); err != nil {
		return fmt.Errorf("wails: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

func services(conn *sql.DB) httpapi.Services {
	now := time.Now
	employees := employee.NewService(employee.NewSQLStore(conn), now)
	sales := sale.NewService(sale.NewSQLStore(conn), employees, now)
	return httpapi.Services{
		Employees: employees,
		Sales:     sales,
		Receipts:  receipt.NewService(receipt.NewSQLStore(conn), sales, now),
		Invoices:  invoice.NewService(invoice.NewSQLStore(conn), sales, now),
	}
}
