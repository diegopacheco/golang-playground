package txn

import (
	"errors"
	"testing"

	"github.com/dgraph-io/badger/v4"
)

func TestRace(t *testing.T) {
	tests := []struct {
		name       string
		first      string
		second     string
		wantSecond error
		wantFirst  int
	}{
		{"same key makes the later commit fail so a read-modify-write cannot silently lose an update", "k", "k", badger.ErrConflict, 1},
		{"different keys both commit so conflict detection does not serialize unrelated work", "a", "b", nil, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := badger.Open(badger.DefaultOptions(t.TempDir()).WithLogger(nil))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			outcome, err := Race(db, tt.first, tt.second)
			if err != nil {
				t.Fatal(err)
			}
			if outcome.First != nil {
				t.Fatalf("first commit = %v, want ok", outcome.First)
			}
			if !errors.Is(outcome.Second, tt.wantSecond) {
				t.Fatalf("second commit = %v, want %v", outcome.Second, tt.wantSecond)
			}
			got, err := Counter(db, tt.first)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.wantFirst {
				t.Fatalf("counter %s = %d, want %d", tt.first, got, tt.wantFirst)
			}
		})
	}
}

func TestRetryAfterConflictAppliesTheLostIncrement(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions(t.TempDir()).WithLogger(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := Race(db, "k", "k"); err != nil {
		t.Fatal(err)
	}
	got, err := IncrementWithRetry(db, "k", 3)
	if err != nil {
		t.Fatal(err)
	}
	if got != 2 {
		t.Fatalf("counter = %d, want 2 because the retried transaction must read the committed value", got)
	}
}
