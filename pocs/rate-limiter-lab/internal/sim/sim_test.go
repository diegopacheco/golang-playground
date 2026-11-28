package sim

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/clock"
	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/limiter"
	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/load"
)

func spec() Spec {
	return Spec{Pattern: load.Bursty(), Tick: 250 * time.Millisecond, Window: time.Second}
}

func results(t *testing.T) map[string]Result {
	t.Helper()
	rs, err := Run(spec(), Algorithms(DefaultLimits()))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]Result{}
	for _, r := range rs {
		out[r.Name] = r
	}
	return out
}

func TestSimulationShowsEachAlgorithmsSignature(t *testing.T) {
	rs := results(t)
	limit := DefaultLimits().Limit
	tests := []struct {
		name  string
		check func() bool
	}{
		{"fixed window lets double the limit through in one second around the window edge", func() bool { return rs["fixed window"].PeakWindow == 2*limit }},
		{"sliding log never exceeds the limit in any one second span", func() bool { return rs["sliding log"].PeakWindow <= limit }},
		{"token bucket spends its saved burst in one tick", func() bool { return rs["token bucket"].PeakTick >= limit }},
		{"leaky bucket smooths the same burst to the leak rate per tick", func() bool {
			return rs["leaky bucket"].PeakTick <= 3 && rs["leaky bucket"].MaxDelay > 0
		}},
		{"gcra admits exactly what the token bucket admits", func() bool {
			return slices.Equal(rs["gcra"].PerTick, rs["token bucket"].PerTick)
		}},
		{"sliding counter is stricter than the fixed window at the edge", func() bool {
			return rs["sliding counter"].PeakWindow < rs["fixed window"].PeakWindow
		}},
		{"only the leaky bucket delays requests, the others answer now or never", func() bool {
			for name, r := range rs {
				if name != "leaky bucket" && r.MaxDelay != 0 {
					return false
				}
			}
			return true
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Fatalf("results: %+v", rs)
			}
		})
	}
}

func TestEveryRequestIsEitherAllowedOrRejected(t *testing.T) {
	total := len(spec().Pattern.Arrivals(spec().Tick))
	for name, r := range results(t) {
		if r.Allowed+r.Rejected != total {
			t.Fatalf("%s: allowed %d + rejected %d != offered %d", name, r.Allowed, r.Rejected, total)
		}
	}
}

func TestRunIsDeterministicSoPlotsAreReproducible(t *testing.T) {
	a, b := results(t), results(t)
	for name := range a {
		if !slices.Equal(a[name].PerTick, b[name].PerTick) {
			t.Fatalf("%s differs between runs", name)
		}
	}
}

func TestRunFailsLoudlyOnABrokenAlgorithm(t *testing.T) {
	bad := []Algorithm{{Name: "bad", New: func(c clock.Clock) (limiter.Limiter, error) { return limiter.NewGCRA(c, 0, 1) }}}
	if _, err := Run(spec(), bad); !errors.Is(err, limiter.ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestPeakInWindow(t *testing.T) {
	ms := func(n int) time.Duration { return time.Duration(n) * time.Millisecond }
	tests := []struct {
		name  string
		times []time.Duration
		want  int
	}{
		{"no executions means no peak", nil, 0},
		{"events exactly one window apart do not share a window", []time.Duration{0, ms(1000)}, 1},
		{"unsorted input is handled", []time.Duration{ms(900), ms(10), ms(950), ms(1500)}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PeakInWindow(tt.times, time.Second); got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}
