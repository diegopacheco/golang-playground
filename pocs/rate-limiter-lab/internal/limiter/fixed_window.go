package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/rate-limiter-lab/internal/clock"
)

type FixedWindow struct {
	mu     sync.Mutex
	clock  clock.Clock
	window time.Duration
	limit  int
	start  time.Time
	count  int
}

func NewFixedWindow(c clock.Clock, window time.Duration, limit int) (*FixedWindow, error) {
	if err := validate(window, limit); err != nil {
		return nil, fmt.Errorf("fixed window window=%v limit=%d: %w", window, limit, err)
	}
	return &FixedWindow{clock: c, window: window, limit: limit}, nil
}

func (w *FixedWindow) Allow() Decision {
	w.mu.Lock()
	defer w.mu.Unlock()
	start := w.clock.Now().Truncate(w.window)
	if !start.Equal(w.start) {
		w.start, w.count = start, 0
	}
	if w.count >= w.limit {
		return rejected
	}
	w.count++
	return allowed
}
