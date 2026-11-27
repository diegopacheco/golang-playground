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

	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/cluster"
	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/config"
	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/httpapi"
	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/node"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "raftkv:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		return err
	}
	logger := logx.New(os.Stderr, cfg.LogLevel).With("node", cfg.NodeID)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	n, err := node.Open(node.DiskConfig{
		ID:        cfg.NodeID,
		HTTPAddr:  cfg.HTTPAddr,
		RaftAddr:  cfg.RaftAddr,
		DataDir:   cfg.DataDir,
		Bootstrap: cfg.Bootstrap,
		Logger:    logx.Raft(os.Stderr, cfg.NodeID, cfg.LogLevel),
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := n.Shutdown(); err != nil {
			logger.Error("node shutdown", "err", err)
		}
	}()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           logx.Middleware(logger, httpapi.New(n).Handler()),
		ReadHeaderTimeout: 5 * time.Second,
	}
	serveErr := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			serveErr <- fmt.Errorf("http server: %w", err)
		}
		close(serveErr)
	}()
	logger.Info("node started", "http", cfg.HTTPAddr, "raft", cfg.RaftAddr, "data", cfg.DataDir, "fresh", n.Fresh())

	runErr := joinIfNeeded(ctx, cfg, n)
	if runErr == nil {
		select {
		case <-ctx.Done():
			logger.Info("shutdown signal received")
		case runErr = <-serveErr:
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return errors.Join(runErr, fmt.Errorf("http shutdown: %w", err))
	}
	return runErr
}

func joinIfNeeded(ctx context.Context, cfg config.Config, n *node.Node) error {
	if cfg.Join == "" || !n.Fresh() {
		return nil
	}
	ctx = logx.WithCorrelationID(ctx, slog.Default(), logx.NewID())
	req := node.JoinRequest{ID: cfg.NodeID, RaftAddr: cfg.RaftAddr, HTTPAddr: cfg.HTTPAddr}
	if err := cluster.Join(ctx, cfg.Join, req, 30); err != nil {
		return err
	}
	slog.Info("joined cluster", "via", cfg.Join)
	return nil
}
