package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/clock"
)

type LeakyBucket struct {
	mu       sync.Mutex
	clock    clock.Clock
	interval time.Duration
	capacity int
	next     time.Time
}

func NewLeakyBucket(c clock.Clock, every time.Duration, queue int) (*LeakyBucket, error) {
	if err := validate(every, queue); err != nil {
		return nil, fmt.Errorf("leaky bucket every=%v queue=%d: %w", every, queue, err)
	}
	return &LeakyBucket{clock: c, interval: every, capacity: queue}, nil
}

func (b *LeakyBucket) Allow() Decision {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.clock.Now()
	if b.next.Before(now) {
		b.next = now
	}
	wait := b.next.Sub(now)
	if int((wait+b.interval-1)/b.interval) >= b.capacity {
		return rejected
	}
	b.next = b.next.Add(b.interval)
	return Decision{Allowed: true, Delay: wait}
}
