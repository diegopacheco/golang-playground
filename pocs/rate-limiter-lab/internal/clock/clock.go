package clock

import "time"

type Clock interface {
	Now() time.Time
}

type System struct{}

func (System) Now() time.Time {
	return time.Now()
}

type Manual struct {
	now time.Time
}

func NewManual(start time.Time) *Manual {
	return &Manual{now: start}
}

func (m *Manual) Now() time.Time {
	return m.now
}

func (m *Manual) Set(t time.Time) {
	m.now = t
}

func (m *Manual) Advance(d time.Duration) {
	m.now = m.now.Add(d)
}
