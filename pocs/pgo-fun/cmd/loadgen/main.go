package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/loadgen"
	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/profile"
	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/report"
)

type options struct {
	url            string
	concurrency    int
	warmup         time.Duration
	duration       time.Duration
	label          string
	out            string
	profilePath    string
	profileSeconds int
}

func main() {
	var o options
	flag.StringVar(&o.url, "url", "http://localhost:8082", "service base url")
	flag.IntVar(&o.concurrency, "c", 32, "concurrent workers")
	flag.DurationVar(&o.warmup, "warmup", 3*time.Second, "warmup before measuring")
	flag.DurationVar(&o.duration, "duration", 15*time.Second, "measured duration")
	flag.StringVar(&o.label, "label", "run", "label written to the results file")
	flag.StringVar(&o.out, "out", "", "append a tab separated result line to this file")
	flag.StringVar(&o.profilePath, "profile", "", "capture a cpu profile from the service into this file while measuring")
	flag.IntVar(&o.profileSeconds, "profile-seconds", 10, "cpu profile length in seconds")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, o); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, o options) error {
	client := &http.Client{
		Timeout:   time.Duration(o.profileSeconds+30) * time.Second,
		Transport: &http.Transport{MaxIdleConns: o.concurrency * 2, MaxIdleConnsPerHost: o.concurrency * 2},
	}
	w, err := loadgen.NewWorkload(42, 256, 16, 100)
	if err != nil {
		return err
	}
	if err := loadgen.Seed(ctx, client, o.url, w); err != nil {
		return err
	}
	profErr := make(chan error, 1)
	onMeasure := func() {
		if o.profilePath == "" {
			profErr <- nil
			return
		}
		go func() {
			profErr <- profile.Capture(ctx, client, o.url, o.profileSeconds, o.profilePath)
		}()
	}
	cfg := loadgen.Config{BaseURL: o.url, Concurrency: o.concurrency, Warmup: o.warmup, Duration: o.duration}
	res, err := loadgen.Run(ctx, client, cfg, w, onMeasure)
	if err != nil {
		return err
	}
	if err := <-profErr; err != nil {
		return err
	}
	r := report.Run{Label: o.label, RPS: res.RPS(), P50: res.P50, P99: res.P99, Requests: res.Requests, Errors: res.Errors}
	fmt.Printf("%-6s %10.1f req/s  p50 %-10s p99 %-10s requests %d errors %d\n", r.Label, r.RPS, r.P50, r.P99, r.Requests, r.Errors)
	if o.out == "" {
		return nil
	}
	f, err := os.OpenFile(o.out, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open results: %w", err)
	}
	if _, err := fmt.Fprintln(f, r.Line()); err != nil {
		f.Close()
		return fmt.Errorf("write results: %w", err)
	}
	return f.Close()
}
