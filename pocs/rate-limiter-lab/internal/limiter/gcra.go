package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/clock"
)

type GCRA struct {
	mu        sync.Mutex
	clock     clock.Clock
	interval  time.Duration
	tolerance time.Duration
	tat       time.Time
}

func NewGCRA(c clock.Clock, every time.Duration, burst int) (*GCRA, error) {
	if err := validate(every, burst); err != nil {
		return nil, fmt.Errorf("gcra every=%v burst=%d: %w", every, burst, err)
	}
	return &GCRA{clock: c, interval: every, tolerance: every * time.Duration(burst-1)}, nil
}

func (g *GCRA) Allow() Decision {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.clock.Now()
	tat := g.tat
	if tat.Before(now) {
		tat = now
	}
	if tat.Sub(now) > g.tolerance {
		return rejected
	}
	g.tat = tat.Add(g.interval)
	return allowed
}
