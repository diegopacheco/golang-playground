package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/clock"
)

type SlidingLog struct {
	mu     sync.Mutex
	clock  clock.Clock
	window time.Duration
	limit  int
	log    []time.Time
}

func NewSlidingLog(c clock.Clock, window time.Duration, limit int) (*SlidingLog, error) {
	if err := validate(window, limit); err != nil {
		return nil, fmt.Errorf("sliding log window=%v limit=%d: %w", window, limit, err)
	}
	return &SlidingLog{clock: c, window: window, limit: limit, log: make([]time.Time, 0, limit)}, nil
}

func (l *SlidingLog) Allow() Decision {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.clock.Now()
	cutoff := now.Add(-l.window)
	expired := 0
	for expired < len(l.log) && !l.log[expired].After(cutoff) {
		expired++
	}
	l.log = append(l.log[:0], l.log[expired:]...)
	if len(l.log) >= l.limit {
		return rejected
	}
	l.log = append(l.log, now)
	return allowed
}
