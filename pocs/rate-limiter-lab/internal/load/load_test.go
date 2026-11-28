package load

import (
	"slices"
	"testing"
	"time"
)

func TestArrivals(t *testing.T) {
	p := Pattern{{Ticks: 1, PerTick: 4}, {Ticks: 1, PerTick: 0}, {Ticks: 1, PerTick: 1}}
	got := p.Arrivals(100 * time.Millisecond)
	want := []time.Duration{0, 25 * time.Millisecond, 50 * time.Millisecond, 75 * time.Millisecond, 200 * time.Millisecond}
	if !slices.Equal(got, want) {
		t.Fatalf("requests must be spread evenly inside their tick so window edges are hit deterministically: got %v, want %v", got, want)
	}
}

func TestBurstyStraddlesAWindowEdge(t *testing.T) {
	per := Bursty().PerTick()
	const ticksPerWindow = 4
	edge := -1
	for i, n := range per {
		if n >= 30 {
			edge = i
			break
		}
	}
	if edge < 0 || (edge+1)%ticksPerWindow != 0 || per[edge+1] < 30 {
		t.Fatalf("the burst must cover the last tick of one 1s window and the first of the next to expose the fixed window boundary effect, got burst at tick %d in %v", edge, per)
	}
	if per[edge-1] != 0 {
		t.Fatal("the tick before the burst must be quiet so the fixed window has its full limit left")
	}
}
