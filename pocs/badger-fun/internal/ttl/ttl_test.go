package ttl

import (
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
)

func openDB(t *testing.T) *badger.DB {
	t.Helper()
	db, err := badger.Open(badger.DefaultOptions(t.TempDir()).WithLogger(nil))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestTTL(t *testing.T) {
	db := openDB(t)
	if err := Put(db, "session", "token", time.Second); err != nil {
		t.Fatal(err)
	}
	if err := Put(db, "user", "ana", 0); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2100 * time.Millisecond)

	tests := []struct {
		name      string
		key       string
		wantFound bool
	}{
		{"session with ttl disappears after it expires so stale tokens cannot be reused", "session", false},
		{"entry without ttl survives the same wait so permanent data is never dropped", "user", true},
		{"missing key reads as not found instead of an error so callers need no special case", "nothing", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, err := Read(db, tt.key)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			if reading.Found != tt.wantFound {
				t.Fatalf("found = %v, want %v", reading.Found, tt.wantFound)
			}
		})
	}
}

func TestTTLIsVisibleBeforeExpiry(t *testing.T) {
	db := openDB(t)
	if err := Put(db, "session", "token", time.Hour); err != nil {
		t.Fatal(err)
	}
	reading, err := Read(db, "session")
	if err != nil {
		t.Fatal(err)
	}
	if !reading.Found || reading.Value != "token" {
		t.Fatalf("reading = %+v, want live token", reading)
	}
	left := time.Until(reading.ExpiresAt)
	if left < 59*time.Minute || left > time.Hour+time.Second {
		t.Fatalf("expires in %s, want about one hour so the ttl given by the caller is the one stored", left)
	}
}
