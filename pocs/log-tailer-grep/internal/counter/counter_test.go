package counter

import (
	"slices"
	"testing"
)

func TestRecord(t *testing.T) {
	cases := []struct {
		name        string
		lines       [][]int
		wantLines   int64
		wantMatched int64
		wantHits    []int64
	}{
		{
			name:        "every line counts toward total so the match ratio is honest",
			lines:       [][]int{nil, nil, {0}},
			wantLines:   3,
			wantMatched: 1,
			wantHits:    []int64{1, 0},
		},
		{
			name:        "a line hitting two patterns counts once per pattern but once as matched",
			lines:       [][]int{{0, 1}},
			wantLines:   1,
			wantMatched: 1,
			wantHits:    []int64{1, 1},
		},
		{
			name:        "counters accumulate across lines so periodic reports show growth",
			lines:       [][]int{{1}, {1}, {0, 1}},
			wantLines:   3,
			wantMatched: 3,
			wantHits:    []int64{1, 3},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := New(2)
			for _, hits := range tc.lines {
				c.Record(hits)
			}
			if c.Lines != tc.wantLines || c.Matched != tc.wantMatched {
				t.Fatalf("lines=%d matched=%d want %d %d", c.Lines, c.Matched, tc.wantLines, tc.wantMatched)
			}
			if !slices.Equal(c.Hits, tc.wantHits) {
				t.Fatalf("hits: got %v want %v", c.Hits, tc.wantHits)
			}
		})
	}
}
