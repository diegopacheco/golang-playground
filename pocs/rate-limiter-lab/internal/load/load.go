package load

import "time"

type Phase struct {
	Name    string
	Ticks   int
	PerTick int
}

type Pattern []Phase

func Bursty() Pattern {
	return Pattern{
		{Name: "steady", Ticks: 8, PerTick: 2},
		{Name: "quiet", Ticks: 3, PerTick: 0},
		{Name: "edge burst", Ticks: 2, PerTick: 30},
		{Name: "quiet", Ticks: 7, PerTick: 0},
		{Name: "overload", Ticks: 16, PerTick: 5},
		{Name: "cooldown", Ticks: 12, PerTick: 1},
	}
}

func (p Pattern) Ticks() int {
	total := 0
	for _, ph := range p {
		total += ph.Ticks
	}
	return total
}

func (p Pattern) PerTick() []int {
	out := make([]int, 0, p.Ticks())
	for _, ph := range p {
		for range ph.Ticks {
			out = append(out, ph.PerTick)
		}
	}
	return out
}

func (p Pattern) Arrivals(tick time.Duration) []time.Duration {
	var out []time.Duration
	for i, n := range p.PerTick() {
		base := time.Duration(i) * tick
		for j := range n {
			out = append(out, base+tick*time.Duration(j)/time.Duration(n))
		}
	}
	return out
}
