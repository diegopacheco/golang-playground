package ttl

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type Reading struct {
	Key       string
	Value     string
	Found     bool
	ExpiresAt time.Time
}

func Put(db *badger.DB, key, value string, ttl time.Duration) error {
	entry := badger.NewEntry([]byte(key), []byte(value))
	if ttl > 0 {
		entry = entry.WithTTL(ttl)
	}
	if err := db.Update(func(txn *badger.Txn) error { return txn.SetEntry(entry) }); err != nil {
		return fmt.Errorf("put %q: %w", key, err)
	}
	return nil
}

func Read(db *badger.DB, key string) (Reading, error) {
	reading := Reading{Key: key}
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		reading.Found = true
		reading.Value = string(value)
		if item.ExpiresAt() > 0 {
			reading.ExpiresAt = time.Unix(int64(item.ExpiresAt()), 0)
		}
		return nil
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return reading, nil
	}
	if err != nil {
		return reading, fmt.Errorf("read %q: %w", key, err)
	}
	return reading, nil
}
