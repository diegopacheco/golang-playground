package txn

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger/v4"
)

type Outcome struct {
	First  error
	Second error
}

func Race(db *badger.DB, firstKey, secondKey string) (Outcome, error) {
	first := db.NewTransaction(true)
	defer first.Discard()
	second := db.NewTransaction(true)
	defer second.Discard()

	if err := increment(first, firstKey); err != nil {
		return Outcome{}, err
	}
	if err := increment(second, secondKey); err != nil {
		return Outcome{}, err
	}
	return Outcome{First: first.Commit(), Second: second.Commit()}, nil
}

func IncrementWithRetry(db *badger.DB, key string, attempts int) (int, error) {
	var err error
	for range attempts {
		err = db.Update(func(t *badger.Txn) error { return increment(t, key) })
		if !errors.Is(err, badger.ErrConflict) {
			break
		}
	}
	if err != nil {
		return 0, fmt.Errorf("increment %q: %w", key, err)
	}
	return Counter(db, key)
}

func Counter(db *badger.DB, key string) (int, error) {
	var value int
	err := db.View(func(t *badger.Txn) error {
		var err error
		value, err = read(t, key)
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("counter %q: %w", key, err)
	}
	return value, nil
}

func increment(t *badger.Txn, key string) error {
	value, err := read(t, key)
	if err != nil {
		return err
	}
	if err := t.Set([]byte(key), []byte(strconv.Itoa(value+1))); err != nil {
		return fmt.Errorf("set %q: %w", key, err)
	}
	return nil
}

func read(t *badger.Txn, key string) (int, error) {
	item, err := t.Get([]byte(key))
	if errors.Is(err, badger.ErrKeyNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get %q: %w", key, err)
	}
	raw, err := item.ValueCopy(nil)
	if err != nil {
		return 0, fmt.Errorf("value %q: %w", key, err)
	}
	value, err := strconv.Atoi(string(raw))
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", key, err)
	}
	return value, nil
}
