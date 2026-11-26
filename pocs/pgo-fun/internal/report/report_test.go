package report

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

func TestLineRoundTrips(t *testing.T) {
	in := Run{Label: "pgo", RPS: 1234.5, P50: 1500 * time.Microsecond, P99: 9 * time.Millisecond, Requests: 100, Errors: 1}
	runs, err := Parse(strings.NewReader(in.Line() + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0] != in {
		t.Fatalf("what loadgen writes must be what the report reads, got %+v", runs)
	}
}

func TestParseRejectsBrokenLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "a truncated line from a killed run must not be averaged in", in: "off\t100\t5\n"},
		{name: "a non numeric throughput must not be read as zero", in: "off\tfast\t1\t2\t3\t0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse(strings.NewReader(tt.in)); !errors.Is(err, ErrFormat) {
				t.Fatalf("want ErrFormat, got %v", err)
			}
		})
	}
}

func TestSummarizeUsesMedianToResistOutliers(t *testing.T) {
	runs := []Run{
		{Label: "off", RPS: 100, P50: 1 * time.Millisecond, P99: 4 * time.Millisecond},
		{Label: "off", RPS: 10, P50: 9 * time.Millisecond, P99: 40 * time.Millisecond},
		{Label: "off", RPS: 110, P50: 2 * time.Millisecond, P99: 5 * time.Millisecond, Errors: 2},
		{Label: "pgo", RPS: 500},
	}
	s, err := Summarize(runs, "off")
	if err != nil {
		t.Fatal(err)
	}
	if s.Rounds != 3 || s.RPS != 100 || s.P50 != 2*time.Millisecond || s.P99 != 5*time.Millisecond || s.Errors != 2 {
		t.Fatalf("one noisy round must not swing the result, got %+v", s)
	}
	if _, err := Summarize(runs, "missing"); err == nil {
		t.Fatalf("a missing build must fail loudly instead of comparing against zero")
	}
}

func TestDelta(t *testing.T) {
	tests := []struct {
		name       string
		base, cand float64
		want       float64
	}{
		{name: "more throughput is a positive gain", base: 100, cand: 105, want: 5},
		{name: "lower latency is a negative delta", base: 200, cand: 190, want: -5},
		{name: "a zero baseline reports no change instead of infinity", base: 0, cand: 10, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Delta(tt.base, tt.cand); math.Abs(got-tt.want) > 1e-9 {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestWriteShowsEveryRoundAndTheMedianDelta(t *testing.T) {
	runs := []Run{{Label: "off", RPS: 100, P50: time.Millisecond}, {Label: "pgo", RPS: 110, P50: time.Millisecond}}
	base, _ := Summarize(runs, "off")
	cand, _ := Summarize(runs, "pgo")
	var b strings.Builder
	if err := Write(&b, runs, base, cand); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"round", "off", "pgo", "+10.0%", "p99 latency"} {
		if !strings.Contains(out, want) {
			t.Fatalf("table missing %q:\n%s", want, out)
		}
	}
}
