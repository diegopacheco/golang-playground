package pages

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/diegopacheco/golang-playground/pocs/iter-seq-fun/internal/seq"
)

type fakeAPI struct {
	pages   map[string]Page[int]
	failAt  string
	fetched []string
}

var errUpstream = errors.New("upstream 503")

func (f *fakeAPI) Fetch(_ context.Context, cursor string) (Page[int], error) {
	if cursor == f.failAt && f.failAt != "" {
		return Page[int]{}, errUpstream
	}
	f.fetched = append(f.fetched, cursor)
	return f.pages[cursor], nil
}

func threePages() *fakeAPI {
	return &fakeAPI{pages: map[string]Page[int]{
		"":   {Items: []int{1, 2, 3}, Next: "p2"},
		"p2": {Items: []int{4, 5, 6}, Next: "p3"},
		"p3": {Items: []int{7, 8}},
	}}
}

func TestOnlyThePagesNeededAreFetched(t *testing.T) {
	tests := []struct {
		name        string
		take        int
		want        []int
		wantFetched []string
	}{
		{"items on the first page never touch page two", 2, []int{1, 2}, []string{""}},
		{"crossing a page boundary fetches exactly one more page", 4, []int{1, 2, 3, 4}, []string{"", "p2"}},
		{"draining everything follows Next until it is empty", 100, []int{1, 2, 3, 4, 5, 6, 7, 8}, []string{"", "p2", "p3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := threePages()
			p := New[int](api)
			got := slices.Collect(seq.Take(p.All(t.Context()), tt.take))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			if !slices.Equal(api.fetched, tt.wantFetched) {
				t.Fatalf("fetched %q, want %q", api.fetched, tt.wantFetched)
			}
			if p.Fetched() != len(tt.wantFetched) {
				t.Fatalf("Fetched() = %d, want %d", p.Fetched(), len(tt.wantFetched))
			}
		})
	}
}

func TestUpstreamErrorStopsIterationAndIsWrapped(t *testing.T) {
	api := threePages()
	api.failAt = "p2"
	p := New[int](api)
	got := slices.Collect(p.All(t.Context()))
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("items before the failure must still be delivered, got %v", got)
	}
	if !errors.Is(p.Err(), errUpstream) {
		t.Fatalf("got %v, want wrapped %v", p.Err(), errUpstream)
	}
}

func TestZipOverTwoPaginatedSourcesFetchesLazilyOnBothSides(t *testing.T) {
	left, right := threePages(), threePages()
	pl, pr := New[int](left), New[int](right)
	var sums []int
	for a, b := range seq.Zip(pl.All(t.Context()), pr.All(t.Context())) {
		sums = append(sums, a+b)
		if len(sums) == 4 {
			break
		}
	}
	if !slices.Equal(sums, []int{2, 4, 6, 8}) {
		t.Fatalf("got %v", sums)
	}
	if len(left.fetched) != 2 || len(right.fetched) != 2 {
		t.Fatalf("zip of 4 pairs must fetch 2 pages per side, got left=%q right=%q", left.fetched, right.fetched)
	}
}
