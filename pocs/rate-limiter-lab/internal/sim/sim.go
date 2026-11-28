package sim

import (
	"fmt"
	"slices"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/clock"
	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/limiter"
	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/load"
)

var epoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

type Algorithm struct {
	Name  string
	State string
	New   func(clock.Clock) (limiter.Limiter, error)
}

type Spec struct {
	Pattern load.Pattern
	Tick    time.Duration
	Window  time.Duration
}

type Result struct {
	Name       string
	State      string
	PerTick    []float64
	Allowed    int
	Rejected   int
	PeakTick   int
	PeakWindow int
	MaxDelay   time.Duration
}

func Offered(spec Spec) []float64 {
	var out []float64
	for _, n := range spec.Pattern.PerTick() {
		out = append(out, float64(n))
	}
	return out
}

func Run(spec Spec, algos []Algorithm) ([]Result, error) {
	arrivals := spec.Pattern.Arrivals(spec.Tick)
	results := make([]Result, 0, len(algos))
	for _, a := range algos {
		r, err := runOne(spec, a, arrivals)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

func runOne(spec Spec, a Algorithm, arrivals []time.Duration) (Result, error) {
	c := clock.NewManual(epoch)
	l, err := a.New(c)
	if err != nil {
		return Result{}, fmt.Errorf("build %s: %w", a.Name, err)
	}
	r := Result{Name: a.Name, State: a.State, PerTick: make([]float64, spec.Pattern.Ticks())}
	var executed []time.Duration
	for _, at := range arrivals {
		c.Set(epoch.Add(at))
		d := l.Allow()
		if !d.Allowed {
			r.Rejected++
			continue
		}
		r.Allowed++
		r.MaxDelay = max(r.MaxDelay, d.Delay)
		exec := at + d.Delay
		executed = append(executed, exec)
		if i := int(exec / spec.Tick); i < len(r.PerTick) {
			r.PerTick[i]++
		}
	}
	for _, v := range r.PerTick {
		r.PeakTick = max(r.PeakTick, int(v))
	}
	r.PeakWindow = PeakInWindow(executed, spec.Window)
	return r, nil
}

func PeakInWindow(times []time.Duration, window time.Duration) int {
	sorted := slices.Sorted(slices.Values(times))
	peak, lo := 0, 0
	for hi := range sorted {
		for sorted[hi]-sorted[lo] >= window {
			lo++
		}
		peak = max(peak, hi-lo+1)
	}
	return peak
}
