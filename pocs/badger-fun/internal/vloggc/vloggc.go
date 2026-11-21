package vloggc

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type Config struct {
	Dir          string
	Keys         int
	ValueSize    int
	Overwrites   int
	DeleteEvery  int
	DiscardRatio float64
}

type Report struct {
	BytesWritten int64
	VlogBefore   int64
	VlogAfter    int64
	FilesBefore  int
	FilesAfter   int
	Rewrites     int
	GCTime       time.Duration
	LiveKeys     int
}

func Run(ctx context.Context, cfg Config) (Report, error) {
	var report Report
	db, err := open(cfg.Dir)
	if err != nil {
		return report, err
	}
	written, err := load(ctx, db, cfg)
	if err != nil {
		return report, errors.Join(err, db.Close())
	}
	report.BytesWritten = written
	if err := db.Close(); err != nil {
		return report, fmt.Errorf("close after load: %w", err)
	}

	db, err = open(cfg.Dir)
	if err != nil {
		return report, err
	}
	defer db.Close()
	if err := db.Flatten(1); err != nil {
		return report, fmt.Errorf("flatten: %w", err)
	}
	if report.VlogBefore, report.FilesBefore, err = VlogSize(cfg.Dir); err != nil {
		return report, err
	}

	start := time.Now()
	if report.Rewrites, err = collect(ctx, db, cfg.DiscardRatio); err != nil {
		return report, err
	}
	report.GCTime = time.Since(start)

	if report.VlogAfter, report.FilesAfter, err = VlogSize(cfg.Dir); err != nil {
		return report, err
	}
	if report.LiveKeys, err = verify(db, cfg); err != nil {
		return report, err
	}
	return report, nil
}

func VlogSize(dir string) (int64, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, fmt.Errorf("read dir: %w", err)
	}
	var total int64
	var files int
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".vlog") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return 0, 0, fmt.Errorf("stat %s: %w", entry.Name(), err)
		}
		total += info.Size()
		files++
	}
	return total, files, nil
}

func open(dir string) (*badger.DB, error) {
	opts := badger.DefaultOptions(filepath.Clean(dir)).
		WithLogger(nil).
		WithValueThreshold(1 << 10).
		WithValueLogFileSize(8 << 20).
		WithNumVersionsToKeep(1).
		WithMemTableSize(64 << 10).
		WithNumLevelZeroTables(1).
		WithBaseTableSize(64 << 10)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dir, err)
	}
	return db, nil
}

func load(ctx context.Context, db *badger.DB, cfg Config) (int64, error) {
	var written int64
	for round := 0; round <= cfg.Overwrites; round++ {
		batch := db.NewWriteBatch()
		for i := range cfg.Keys {
			if err := ctx.Err(); err != nil {
				batch.Cancel()
				return written, err
			}
			value := make([]byte, cfg.ValueSize)
			if _, err := rand.Read(value); err != nil {
				batch.Cancel()
				return written, fmt.Errorf("random value: %w", err)
			}
			if err := batch.Set(key(i), value); err != nil {
				batch.Cancel()
				return written, fmt.Errorf("set round %d: %w", round, err)
			}
			written += int64(cfg.ValueSize)
		}
		if err := batch.Flush(); err != nil {
			return written, fmt.Errorf("flush round %d: %w", round, err)
		}
	}
	batch := db.NewWriteBatch()
	for i := range cfg.Keys {
		if !deleted(i, cfg.DeleteEvery) {
			continue
		}
		if err := batch.Delete(key(i)); err != nil {
			batch.Cancel()
			return written, fmt.Errorf("delete: %w", err)
		}
	}
	if err := batch.Flush(); err != nil {
		return written, fmt.Errorf("flush deletes: %w", err)
	}
	return written, nil
}

func collect(ctx context.Context, db *badger.DB, ratio float64) (int, error) {
	rewrites := 0
	for {
		if err := ctx.Err(); err != nil {
			return rewrites, err
		}
		err := db.RunValueLogGC(ratio)
		if errors.Is(err, badger.ErrNoRewrite) {
			return rewrites, nil
		}
		if err != nil {
			return rewrites, fmt.Errorf("value log gc: %w", err)
		}
		rewrites++
	}
}

func verify(db *badger.DB, cfg Config) (int, error) {
	live := 0
	err := db.View(func(txn *badger.Txn) error {
		for i := range cfg.Keys {
			item, err := txn.Get(key(i))
			if deleted(i, cfg.DeleteEvery) {
				if !errors.Is(err, badger.ErrKeyNotFound) {
					return fmt.Errorf("deleted key %d is still readable", i)
				}
				continue
			}
			if err != nil {
				return fmt.Errorf("live key %d: %w", i, err)
			}
			value, err := item.ValueCopy(nil)
			if err != nil {
				return fmt.Errorf("live key %d value: %w", i, err)
			}
			if len(value) != cfg.ValueSize {
				return fmt.Errorf("live key %d has %d bytes, want %d", i, len(value), cfg.ValueSize)
			}
			live++
		}
		return nil
	})
	return live, err
}

func deleted(i, every int) bool {
	return every > 0 && i%every == 0
}

func key(i int) []byte {
	return fmt.Appendf(nil, "blob-%06d", i)
}
