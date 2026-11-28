package limiter

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/clock"
)

var start = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

type step struct {
	at    time.Duration
	n     int
	want  int
	delay []time.Duration
}

func run(t *testing.T, c *clock.Manual, l Limiter, steps []step) {
	t.Helper()
	for i, s := range steps {
		c.Set(start.Add(s.at))
		got := 0
		var delays []time.Duration
		for range s.n {
			d := l.Allow()
			if d.Allowed {
				got++
				delays = append(delays, d.Delay)
			}
		}
		if got != s.want {
			t.Fatalf("step %d at %v: allowed %d of %d, want %d", i, s.at, got, s.n, s.want)
		}
		if s.delay != nil {
			for j, want := range s.delay {
				if delays[j] != want {
					t.Fatalf("step %d: delay[%d] = %v, want %v", i, j, delays[j], want)
				}
			}
		}
	}
}

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

func TestTokenBucket(t *testing.T) {
	tests := []struct {
		name  string
		steps []step
	}{
		{"a full bucket absorbs a burst up to its capacity and then rejects", []step{{at: 0, n: 8, want: 5}}},
		{"tokens refill at the configured rate so one interval buys one request", []step{{at: 0, n: 5, want: 5}, {at: ms(100), n: 3, want: 1}, {at: ms(350), n: 3, want: 2}}},
		{"a long idle period refills only up to capacity, never more", []step{{at: 0, n: 5, want: 5}, {at: time.Hour, n: 20, want: 5}}},
		{"partial credit is kept so sub interval gaps add up", []step{{at: 0, n: 5, want: 5}, {at: ms(50), n: 1, want: 0}, {at: ms(100), n: 1, want: 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := clock.NewManual(start)
			l, err := NewTokenBucket(c, ms(100), 5)
			if err != nil {
				t.Fatal(err)
			}
			run(t, c, l, tt.steps)
		})
	}
}

func TestLeakyBucket(t *testing.T) {
	tests := []struct {
		name  string
		steps []step
	}{
		{"a burst is queued and released at a constant pace, which is the smoothing", []step{{at: 0, n: 4, want: 4, delay: []time.Duration{0, ms(100), ms(200), ms(300)}}}},
		{"a full queue rejects instead of growing latency without bound", []step{{at: 0, n: 10, want: 4}}},
		{"the queue drains over time so new requests get a slot with a shorter wait", []step{{at: 0, n: 4, want: 4}, {at: ms(250), n: 3, want: 2, delay: []time.Duration{ms(150), ms(250)}}}},
		{"an idle queue serves the next request immediately", []step{{at: 0, n: 1, want: 1}, {at: time.Second, n: 1, want: 1, delay: []time.Duration{0}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := clock.NewManual(start)
			l, err := NewLeakyBucket(c, ms(100), 4)
			if err != nil {
				t.Fatal(err)
			}
			run(t, c, l, tt.steps)
		})
	}
}

func TestGCRA(t *testing.T) {
	tests := []struct {
		name  string
		steps []step
	}{
		{"the burst tolerance admits burst requests at once", []step{{at: 0, n: 8, want: 5}}},
		{"after a burst requests are admitted exactly one interval apart", []step{{at: 0, n: 5, want: 5}, {at: ms(99), n: 1, want: 0}, {at: ms(100), n: 2, want: 1}, {at: ms(200), n: 2, want: 1}}},
		{"idle time does not bank more than the burst", []step{{at: time.Hour, n: 9, want: 5}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := clock.NewManual(start)
			l, err := NewGCRA(c, ms(100), 5)
			if err != nil {
				t.Fatal(err)
			}
			run(t, c, l, tt.steps)
		})
	}
}

func TestGCRAMakesTheSameDecisionsAsTokenBucketWithOneTimestampOfState(t *testing.T) {
	c := clock.NewManual(start)
	tb, err := NewTokenBucket(c, ms(100), 5)
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewGCRA(c, ms(100), 5)
	if err != nil {
		t.Fatal(err)
	}
	offsets := []int{0, 0, 0, 10, 20, 30, 90, 100, 105, 230, 231, 600, 600, 600, 600, 600, 600, 610, 700, 1500, 1500, 1500}
	for i, off := range offsets {
		c.Set(start.Add(ms(off)))
		if a, b := tb.Allow().Allowed, g.Allow().Allowed; a != b {
			t.Fatalf("request %d at %dms: token bucket=%v gcra=%v", i, off, a, b)
		}
	}
}

func boundaryBurst(t *testing.T, newLimiter func(clock.Clock) (Limiter, error), steps []step) {
	t.Helper()
	c := clock.NewManual(start)
	l, err := newLimiter(c)
	if err != nil {
		t.Fatal(err)
	}
	run(t, c, l, steps)
}

func TestWindowsAtTheBoundary(t *testing.T) {
	burst := []step{{at: ms(900), n: 10, want: 10}, {at: ms(1000), n: 10}}
	tests := []struct {
		name      string
		new       func(clock.Clock) (Limiter, error)
		atEdge    int
		laterAt   time.Duration
		laterWant int
	}{
		{
			"fixed window resets at the edge so 20 requests pass within 100ms, double the limit",
			func(c clock.Clock) (Limiter, error) { return NewFixedWindow(c, time.Second, 10) },
			10, ms(1500), 0,
		},
		{
			"sliding log remembers the last second so the edge burst and anything before the old entries expire is rejected",
			func(c clock.Clock) (Limiter, error) { return NewSlidingLog(c, time.Second, 10) },
			0, ms(1500), 0,
		},
		{
			"sliding counter weights the previous window so the edge burst is also rejected",
			func(c clock.Clock) (Limiter, error) { return NewSlidingCounter(c, time.Second, 10) },
			0, ms(1500), 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := append([]step(nil), burst...)
			steps[1].want = tt.atEdge
			steps = append(steps, step{at: tt.laterAt, n: 10, want: tt.laterWant})
			boundaryBurst(t, tt.new, steps)
		})
	}
}

func TestSlidingLogFreesExactlyOneWindowAfterEachRequest(t *testing.T) {
	boundaryBurst(t, func(c clock.Clock) (Limiter, error) { return NewSlidingLog(c, time.Second, 10) }, []step{
		{at: ms(900), n: 10, want: 10},
		{at: ms(1899), n: 1, want: 0},
		{at: ms(1900), n: 12, want: 10},
	})
}

func TestSlidingCounterIsAnApproximationThatAssumesEvenSpread(t *testing.T) {
	tests := []struct {
		name string
		new  func(clock.Clock) (Limiter, error)
		want int
	}{
		{"the exact log sees that all early requests expired and allows a full window", func(c clock.Clock) (Limiter, error) { return NewSlidingLog(c, time.Second, 10) }, 10},
		{"the counter assumes the previous window was spread evenly and allows only half", func(c clock.Clock) (Limiter, error) { return NewSlidingCounter(c, time.Second, 10) }, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boundaryBurst(t, tt.new, []step{{at: 0, n: 10, want: 10}, {at: ms(1500), n: 10, want: tt.want}})
		})
	}
}

func TestSlidingCounterForgetsEverythingAfterAFullyEmptyWindow(t *testing.T) {
	boundaryBurst(t, func(c clock.Clock) (Limiter, error) { return NewSlidingCounter(c, time.Second, 10) }, []step{
		{at: 0, n: 10, want: 10},
		{at: ms(2100), n: 12, want: 10},
	})
}

func TestConstructorsRejectNonsenseConfig(t *testing.T) {
	c := clock.NewManual(start)
	ctors := map[string]func() error{
		"token bucket":    func() error { _, err := NewTokenBucket(c, 0, 5); return err },
		"leaky bucket":    func() error { _, err := NewLeakyBucket(c, ms(100), 0); return err },
		"fixed window":    func() error { _, err := NewFixedWindow(c, -time.Second, 5); return err },
		"sliding log":     func() error { _, err := NewSlidingLog(c, time.Second, 0); return err },
		"sliding counter": func() error { _, err := NewSlidingCounter(c, 0, 0); return err },
		"gcra":            func() error { _, err := NewGCRA(c, ms(100), -1); return err },
	}
	for name, ctor := range ctors {
		t.Run(name+" refuses a zero or negative setting that would divide by zero or allow nothing", func(t *testing.T) {
			if err := ctor(); !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("error = %v, want ErrInvalidConfig", err)
			}
		})
	}
}

func TestConcurrentCallersNeverExceedTheCapacity(t *testing.T) {
	c := clock.NewManual(start)
	build := map[string]func() (Limiter, error){
		"token bucket":    func() (Limiter, error) { return NewTokenBucket(c, ms(100), 7) },
		"leaky bucket":    func() (Limiter, error) { return NewLeakyBucket(c, ms(100), 7) },
		"fixed window":    func() (Limiter, error) { return NewFixedWindow(c, time.Second, 7) },
		"sliding log":     func() (Limiter, error) { return NewSlidingLog(c, time.Second, 7) },
		"sliding counter": func() (Limiter, error) { return NewSlidingCounter(c, time.Second, 7) },
		"gcra":            func() (Limiter, error) { return NewGCRA(c, ms(100), 7) },
	}
	for name, newLimiter := range build {
		t.Run(name, func(t *testing.T) {
			l, err := newLimiter()
			if err != nil {
				t.Fatal(err)
			}
			var ok atomic.Int32
			var wg sync.WaitGroup
			for range 64 {
				wg.Go(func() {
					if l.Allow().Allowed {
						ok.Add(1)
					}
				})
			}
			wg.Wait()
			if got := ok.Load(); got != 7 {
				t.Fatalf("allowed %d concurrent requests, want exactly 7", got)
			}
		})
	}
}
