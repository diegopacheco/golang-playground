package catalog

import (
	"context"
	"errors"
	"testing"
)

func TestFetchWalksPagesAndRejectsBadInput(t *testing.T) {
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	tests := []struct {
		name      string
		ctx       context.Context
		cursor    string
		wantItems int
		wantNext  string
		wantErr   error
	}{
		{"first page starts at the beginning and points to the next", t.Context(), "", 10, "10", nil},
		{"last page is short and ends pagination with an empty Next", t.Context(), "20", 5, "", nil},
		{"garbage cursor is rejected so callers cannot loop forever", t.Context(), "abc", 0, "", ErrBadCursor},
		{"cursor past the end is rejected", t.Context(), "99", 0, "", ErrBadCursor},
		{"canceled request does no work", canceled, "", 0, "", context.Canceled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore(25, 10)
			page, err := s.Fetch(tt.ctx, tt.cursor)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error %v, want %v", err, tt.wantErr)
			}
			if len(page.Items) != tt.wantItems || page.Next != tt.wantNext {
				t.Fatalf("got %d items next=%q, want %d next=%q", len(page.Items), page.Next, tt.wantItems, tt.wantNext)
			}
			wantCalls := 0
			if tt.wantErr == nil {
				wantCalls = 1
			}
			if s.Calls() != wantCalls {
				t.Fatalf("Calls() = %d, want %d so the laziness counter stays honest", s.Calls(), wantCalls)
			}
		})
	}
}

func TestPagesRoundsUpForAPartialLastPage(t *testing.T) {
	if got := NewStore(25, 10).Pages(); got != 3 {
		t.Fatalf("Pages() = %d, want 3", got)
	}
}
