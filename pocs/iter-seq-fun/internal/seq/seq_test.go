package seq

import (
	"iter"
	"slices"
	"testing"
)

type probe struct {
	pulls  int
	closed bool
}

func (p *probe) seq(values []int) iter.Seq[int] {
	return func(yield func(int) bool) {
		defer func() { p.closed = true }()
		for _, v := range values {
			p.pulls++
			if !yield(v) {
				return
			}
		}
	}
}

func TestTakeStopsPullingTheSource(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		want      []int
		wantPulls int
	}{
		{"zero limit never touches an expensive source", 0, nil, 0},
		{"negative limit is treated as nothing to take", -1, nil, 0},
		{"limit inside the source pulls exactly that many", 2, []int{1, 2}, 2},
		{"limit past the end returns everything without failing", 10, []int{1, 2, 3, 4, 5}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &probe{}
			got := slices.Collect(Take(p.seq([]int{1, 2, 3, 4, 5}), tt.n))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			if p.pulls != tt.wantPulls {
				t.Fatalf("pulled %d values, want %d", p.pulls, tt.wantPulls)
			}
		})
	}
}

func TestPipelinePullsOnlyWhatTheConsumerNeeds(t *testing.T) {
	tests := []struct {
		name      string
		take      int
		want      []int
		wantPulls int
	}{
		{"first even square needs only two source values", 1, []int{4}, 2},
		{"two even squares need four source values", 2, []int{4, 16}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &probe{}
			even := Filter(p.seq([]int{1, 2, 3, 4, 5, 6, 7, 8}), func(v int) bool { return v%2 == 0 })
			squares := Map(even, func(v int) int { return v * v })
			got := slices.Collect(Take(squares, tt.take))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			if p.pulls != tt.wantPulls {
				t.Fatalf("pulled %d values, want %d", p.pulls, tt.wantPulls)
			}
		})
	}
}

func TestZipPairsUntilTheShorterSequenceEnds(t *testing.T) {
	tests := []struct {
		name  string
		left  []int
		right []int
		want  [][2]int
	}{
		{"left shorter stops the pairing", []int{1, 2}, []int{10, 20, 30}, [][2]int{{1, 10}, {2, 20}}},
		{"right shorter stops the pairing", []int{1, 2, 3}, []int{10}, [][2]int{{1, 10}}},
		{"empty side yields no pairs and never starts the other side", nil, []int{10}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left, right := &probe{}, &probe{}
			var got [][2]int
			for a, b := range Zip(left.seq(tt.left), right.seq(tt.right)) {
				got = append(got, [2]int{a, b})
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for side, p := range map[string]*probe{"left": left, "right": right} {
				if p.pulls > 0 && !p.closed {
					t.Fatalf("%s sequence was started but never released", side)
				}
			}
			if len(tt.left) == 0 && right.pulls != 0 {
				t.Fatalf("right side pulled %d values although left was empty", right.pulls)
			}
		})
	}
}

func TestZipReleasesAnInfiniteSequenceWhenTheConsumerBreaks(t *testing.T) {
	right := &probe{}
	count := 0
	for range Zip(From(1), right.seq([]int{7, 8, 9})) {
		count++
		if count == 2 {
			break
		}
	}
	if right.pulls != 2 {
		t.Fatalf("pulled %d values from the finite side, want 2", right.pulls)
	}
	if !right.closed {
		t.Fatal("breaking out of Zip must stop the pulled coroutines so they do not leak")
	}
}
