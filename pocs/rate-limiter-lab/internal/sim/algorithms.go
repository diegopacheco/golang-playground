package sim

import (
	"fmt"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/clock"
	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/limiter"
)

type Limits struct {
	Every  time.Duration
	Burst  int
	Window time.Duration
	Limit  int
}

func DefaultLimits() Limits {
	return Limits{Every: 100 * time.Millisecond, Burst: 10, Window: time.Second, Limit: 10}
}

func Algorithms(l Limits) []Algorithm {
	return []Algorithm{
		{Name: "token bucket", State: "credit + last time", New: func(c clock.Clock) (limiter.Limiter, error) {
			return limiter.NewTokenBucket(c, l.Every, l.Burst)
		}},
		{Name: "leaky bucket", State: "next free slot", New: func(c clock.Clock) (limiter.Limiter, error) {
			return limiter.NewLeakyBucket(c, l.Every, l.Burst)
		}},
		{Name: "fixed window", State: "window start + count", New: func(c clock.Clock) (limiter.Limiter, error) {
			return limiter.NewFixedWindow(c, l.Window, l.Limit)
		}},
		{Name: "sliding log", State: fmt.Sprintf("up to %d timestamps", l.Limit), New: func(c clock.Clock) (limiter.Limiter, error) {
			return limiter.NewSlidingLog(c, l.Window, l.Limit)
		}},
		{Name: "sliding counter", State: "two counters", New: func(c clock.Clock) (limiter.Limiter, error) {
			return limiter.NewSlidingCounter(c, l.Window, l.Limit)
		}},
		{Name: "gcra", State: "one timestamp", New: func(c clock.Clock) (limiter.Limiter, error) {
			return limiter.NewGCRA(c, l.Every, l.Burst)
		}},
	}
}
