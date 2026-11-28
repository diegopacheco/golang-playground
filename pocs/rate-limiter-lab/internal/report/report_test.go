package report

import (
	"strings"
	"testing"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/load"
	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/sim"
)

func build(t *testing.T, color bool) Report {
	t.Helper()
	spec := sim.Spec{Pattern: load.Bursty(), Tick: 250 * time.Millisecond, Window: time.Second}
	limits := sim.DefaultLimits()
	results, err := sim.Run(spec, sim.Algorithms(limits))
	if err != nil {
		t.Fatal(err)
	}
	return Report{Spec: spec, Limits: limits, Offered: sim.Offered(spec), Results: results, Color: color}
}

func TestPlainReportHasNoEscapeCodesSoItCanBePipedToFiles(t *testing.T) {
	if out := build(t, false).String(); strings.Contains(out, "\x1b[") {
		t.Fatal("plain report contains ANSI escape codes")
	}
}

func TestColorReportIsColored(t *testing.T) {
	if out := build(t, true).String(); !strings.Contains(out, "\x1b[") {
		t.Fatal("color report has no ANSI escape codes")
	}
}

func TestReportNamesEveryAlgorithmInPlotsAndSummary(t *testing.T) {
	r := build(t, false)
	plots, summary := r.Plots(), r.Summary()
	for _, a := range sim.Algorithms(sim.DefaultLimits()) {
		if !strings.Contains(plots, a.Name+" · allowed per tick") {
			t.Errorf("plots miss %s", a.Name)
		}
		if !strings.Contains(summary, a.Name) {
			t.Errorf("summary misses %s", a.Name)
		}
	}
}

func TestSummaryRowsShowTheBoundaryEffect(t *testing.T) {
	summary := build(t, false).Summary()
	for line := range strings.Lines(summary) {
		if strings.HasPrefix(line, "fixed window") && !strings.Contains(line, " 20 ") {
			t.Fatalf("fixed window row must show a peak of 20 in one window:\n%s", summary)
		}
	}
}

func TestAxisKeepsOneRowPerRequestOrAWholeStepSoLabelsAreExactCounts(t *testing.T) {
	tests := []struct {
		name      string
		peak      float64
		wantUpper float64
		wantRows  int
	}{
		{"small peaks get one row per request", 12, 12, 12},
		{"a peak of zero still draws one row", 0.4, 1, 1},
		{"the offered burst of 30 still gets one row per request", 30, 30, 30},
		{"larger peaks use a whole number step per row", 75, 90, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upper, rows := axis(tt.peak)
			if upper != tt.wantUpper || rows != tt.wantRows {
				t.Fatalf("axis(%v) = %v, %d, want %v, %d", tt.peak, upper, rows, tt.wantUpper, tt.wantRows)
			}
		})
	}
}
