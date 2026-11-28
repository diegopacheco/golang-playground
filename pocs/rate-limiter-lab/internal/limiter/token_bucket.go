package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/clock"
)

type TokenBucket struct {
	mu       sync.Mutex
	clock    clock.Clock
	interval time.Duration
	capacity time.Duration
	credit   time.Duration
	last     time.Time
}

func NewTokenBucket(c clock.Clock, every time.Duration, burst int) (*TokenBucket, error) {
	if err := validate(every, burst); err != nil {
		return nil, fmt.Errorf("token bucket every=%v burst=%d: %w", every, burst, err)
	}
	capacity := every * time.Duration(burst)
	return &TokenBucket{clock: c, interval: every, capacity: capacity, credit: capacity, last: c.Now()}, nil
}

func (b *TokenBucket) Allow() Decision {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.clock.Now()
	b.credit = min(b.capacity, b.credit+max(0, now.Sub(b.last)))
	b.last = now
	if b.credit < b.interval {
		return rejected
	}
	b.credit -= b.interval
	return allowed
}
