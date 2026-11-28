package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/ansihtml"
	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/load"
	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/report"
	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/sim"
)

type options struct {
	tick    time.Duration
	plain   bool
	capture string
}

func main() {
	var o options
	flag.DurationVar(&o.tick, "tick", 250*time.Millisecond, "simulated time per plotted tick")
	flag.BoolVar(&o.plain, "plain", false, "disable ANSI colors")
	flag.StringVar(&o.capture, "capture", "", "write each report section as HTML into this folder and exit")
	flag.Parse()

	if err := run(o); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(o options) error {
	if o.tick <= 0 {
		return fmt.Errorf("tick must be positive, got %v", o.tick)
	}
	limits := sim.DefaultLimits()
	spec := sim.Spec{Pattern: load.Bursty(), Tick: o.tick, Window: limits.Window}
	results, err := sim.Run(spec, sim.Algorithms(limits))
	if err != nil {
		return fmt.Errorf("simulate: %w", err)
	}
	r := report.Report{Spec: spec, Limits: limits, Offered: sim.Offered(spec), Results: results, Color: !o.plain}
	if o.capture != "" {
		return capture(o.capture, r)
	}
	fmt.Print(r.String())
	return nil
}

func capture(dir string, r report.Report) error {
	pages := []struct {
		name, title, content string
	}{
		{"01-offered-load", "Offered load", r.Header() + "\n\n" + r.OfferedPlot()},
		{"02-allowed-per-tick", "Allowed per tick", r.Plots()},
		{"03-summary", "Summary", r.Summary()},
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create capture dir: %w", err)
	}
	for _, p := range pages {
		page := ansihtml.Page(p.title, p.content, ansihtml.Dark)
		if err := os.WriteFile(filepath.Join(dir, p.name+".html"), []byte(page), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", p.name, err)
		}
	}
	return nil
}
