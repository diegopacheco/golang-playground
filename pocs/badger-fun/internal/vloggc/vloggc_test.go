package vloggc

import (
	"context"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		overwrites  int
		deleteEvery int
		wantShrink  bool
		wantLive    int
	}{
		{"overwritten and deleted values are reclaimed so the value log does not grow forever", 3, 2, true, 200},
		{"values written once are all live so gc has nothing to rewrite and keeps every file", 0, 0, false, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Dir:          t.TempDir(),
				Keys:         400,
				ValueSize:    16 << 10,
				Overwrites:   tt.overwrites,
				DeleteEvery:  tt.deleteEvery,
				DiscardRatio: 0.5,
			}
			report, err := Run(context.Background(), cfg)
			if err != nil {
				t.Fatal(err)
			}
			shrunk := report.VlogAfter < report.VlogBefore
			if shrunk != tt.wantShrink {
				t.Fatalf("vlog %d -> %d bytes, shrink = %v, want %v", report.VlogBefore, report.VlogAfter, shrunk, tt.wantShrink)
			}
			if tt.wantShrink && report.Rewrites == 0 {
				t.Fatal("space was reclaimed without any gc rewrite")
			}
			if report.LiveKeys != tt.wantLive {
				t.Fatalf("live keys = %d, want %d so gc never drops data that is still referenced", report.LiveKeys, tt.wantLive)
			}
		})
	}
}

func TestRunStopsWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Run(ctx, Config{Dir: t.TempDir(), Keys: 10, ValueSize: 4 << 10, DiscardRatio: 0.5})
	if err == nil {
		t.Fatal("want context error so a long benchmark can be interrupted")
	}
}
