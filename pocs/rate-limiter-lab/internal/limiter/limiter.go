package limiter

import (
	"errors"
	"time"
)

var ErrInvalidConfig = errors.New("invalid limiter config")

type Decision struct {
	Allowed bool
	Delay   time.Duration
}

type Limiter interface {
	Allow() Decision
}

var (
	allowed  = Decision{Allowed: true}
	rejected = Decision{}
)

func validate(every time.Duration, n int) error {
	if every <= 0 || n <= 0 {
		return ErrInvalidConfig
	}
	return nil
}
