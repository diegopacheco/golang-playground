package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/clock"
)

type SlidingCounter struct {
	mu     sync.Mutex
	clock  clock.Clock
	window time.Duration
	limit  int
	start  time.Time
	prev   int
	curr   int
}

func NewSlidingCounter(c clock.Clock, window time.Duration, limit int) (*SlidingCounter, error) {
	if err := validate(window, limit); err != nil {
		return nil, fmt.Errorf("sliding counter window=%v limit=%d: %w", window, limit, err)
	}
	return &SlidingCounter{clock: c, window: window, limit: limit}, nil
}

func (s *SlidingCounter) Allow() Decision {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock.Now()
	start := now.Truncate(s.window)
	switch {
	case start.Equal(s.start):
	case start.Equal(s.start.Add(s.window)):
		s.start, s.prev, s.curr = start, s.curr, 0
	default:
		s.start, s.prev, s.curr = start, 0, 0
	}
	remaining := s.window - now.Sub(s.start)
	weighted := int64(s.prev)*int64(remaining) + int64(s.curr)*int64(s.window)
	if weighted+int64(s.window) > int64(s.limit)*int64(s.window) {
		return rejected
	}
	s.curr++
	return allowed
}
