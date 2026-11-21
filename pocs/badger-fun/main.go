package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"

	"github.com/diegopacheco/golang-playground/pocs/badger-fun/internal/ttl"
	"github.com/diegopacheco/golang-playground/pocs/badger-fun/internal/txn"
	"github.com/diegopacheco/golang-playground/pocs/badger-fun/internal/vloggc"
)

func main() {
	dir := flag.String("dir", ".run", "data directory")
	keys := flag.Int("keys", 2000, "keys written by the gc benchmark")
	valueSize := flag.Int("value-size", 16<<10, "value size in bytes for the gc benchmark")
	overwrites := flag.Int("overwrites", 3, "times every key is overwritten")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := vloggc.Config{
		Dir:          filepath.Join(*dir, "gc"),
		Keys:         *keys,
		ValueSize:    *valueSize,
		Overwrites:   *overwrites,
		DeleteEvery:  2,
		DiscardRatio: 0.5,
	}
	if err := run(ctx, filepath.Join(*dir, "kv"), cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, kvDir string, cfg vloggc.Config) error {
	for _, dir := range []string{kvDir, cfg.Dir} {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("reset %s: %w", dir, err)
		}
	}
	db, err := badger.Open(badger.DefaultOptions(kvDir).WithLogger(nil))
	if err != nil {
		return fmt.Errorf("open %s: %w", kvDir, err)
	}
	err = errors.Join(showTTL(ctx, db), showConflict(db))
	if closeErr := db.Close(); closeErr != nil {
		err = errors.Join(err, fmt.Errorf("close: %w", closeErr))
	}
	if err != nil {
		return err
	}
	return showGC(ctx, cfg)
}

func showTTL(ctx context.Context, db *badger.DB) error {
	fmt.Println("== TTL entries")
	if err := ttl.Put(db, "session:42", "token-abc", 2*time.Second); err != nil {
		return err
	}
	if err := ttl.Put(db, "user:42", "ana", 0); err != nil {
		return err
	}
	if err := printReadings(db, "t=0s"); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
	}
	return printReadings(db, "t=3s")
}

func printReadings(db *badger.DB, label string) error {
	for _, key := range []string{"session:42", "user:42"} {
		reading, err := ttl.Read(db, key)
		if err != nil {
			return err
		}
		switch {
		case !reading.Found:
			fmt.Printf("  %-5s %-11s expired, ErrKeyNotFound\n", label, key)
		case reading.ExpiresAt.IsZero():
			fmt.Printf("  %-5s %-11s %q no ttl\n", label, key, reading.Value)
		default:
			fmt.Printf("  %-5s %-11s %q expires in %s\n", label, key, reading.Value, time.Until(reading.ExpiresAt).Round(time.Second))
		}
	}
	return nil
}

func showConflict(db *badger.DB) error {
	fmt.Println("== Read-write transactions")
	cases := []struct {
		label  string
		first  string
		second string
	}{
		{"same key", "counter:a", "counter:a"},
		{"different keys", "counter:b", "counter:c"},
	}
	for _, c := range cases {
		outcome, err := txn.Race(db, c.first, c.second)
		if err != nil {
			return err
		}
		fmt.Printf("  %-14s txn1 commit: %s, txn2 commit: %s\n", c.label, describe(outcome.First), describe(outcome.Second))
	}
	value, err := txn.Counter(db, "counter:a")
	if err != nil {
		return err
	}
	fmt.Printf("  counter:a after conflict = %d\n", value)
	value, err = txn.IncrementWithRetry(db, "counter:a", 3)
	if err != nil {
		return err
	}
	fmt.Printf("  counter:a after retry    = %d\n", value)
	return nil
}

func describe(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, badger.ErrConflict):
		return "ErrConflict"
	default:
		return err.Error()
	}
}

func showGC(ctx context.Context, cfg vloggc.Config) error {
	fmt.Println("== Value log GC")
	fmt.Printf("  %d keys x %d KiB values, written %d times, every %dnd key deleted\n",
		cfg.Keys, cfg.ValueSize>>10, cfg.Overwrites+1, cfg.DeleteEvery)
	report, err := vloggc.Run(ctx, cfg)
	if err != nil {
		return err
	}
	fmt.Printf("  bytes written      %s\n", mib(report.BytesWritten))
	fmt.Printf("  vlog before gc     %s in %d files\n", mib(report.VlogBefore), report.FilesBefore)
	fmt.Printf("  vlog after gc      %s in %d files\n", mib(report.VlogAfter), report.FilesAfter)
	fmt.Printf("  reclaimed          %s (%.0f%%)\n", mib(report.VlogBefore-report.VlogAfter),
		100*float64(report.VlogBefore-report.VlogAfter)/float64(max(report.VlogBefore, 1)))
	fmt.Printf("  gc rewrites        %d\n", report.Rewrites)
	fmt.Printf("  gc time            %s\n", report.GCTime.Round(time.Millisecond))
	fmt.Printf("  live keys verified %d\n", report.LiveKeys)
	return nil
}

func mib(n int64) string {
	return fmt.Sprintf("%.1f MiB", float64(n)/(1<<20))
}
