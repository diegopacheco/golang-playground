package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/log-tailer-grep/internal/config"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name    string
		all     bool
		want    []string
		notWant []string
	}{
		{
			name:    "only matching lines print by default so the output behaves like grep",
			want:    []string{"api.log | ERROR boom", "db.log  | WARN slow", "== counters final lines=3 matched=2 ERROR=1 WARN=1"},
			notWant: []string{"INFO ok"},
		},
		{
			name: "all mode prints every line while still counting so context is not lost",
			all:  true,
			want: []string{"api.log | INFO ok", "api.log | ERROR boom", "== counters final lines=3 matched=2 ERROR=1 WARN=1"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			api := filepath.Join(dir, "api.log")
			db := filepath.Join(dir, "db.log")
			if err := os.WriteFile(api, []byte("INFO ok\nERROR boom\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(db, []byte("WARN slow\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			cfg := config.Config{
				Files:     []string{api, db},
				Patterns:  []string{"ERROR", "WARN"},
				Interval:  10 * time.Millisecond,
				Duration:  300 * time.Millisecond,
				FromStart: true,
				All:       tc.all,
			}
			var out bytes.Buffer
			if err := Run(context.Background(), cfg, &out); err != nil {
				t.Fatalf("run: %v", err)
			}
			text := out.String()
			for _, w := range tc.want {
				if !strings.Contains(text, w) {
					t.Fatalf("missing %q in:\n%s", w, text)
				}
			}
			for _, w := range tc.notWant {
				if strings.Contains(text, w) {
					t.Fatalf("unexpected %q in:\n%s", w, text)
				}
			}
		})
	}
}

func TestRunReturnsOnCancelSoCtrlCPrintsFinalCounters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	var out bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, config.Config{Files: []string{path}, Patterns: []string{"x"}, Interval: 10 * time.Millisecond}, &out)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not return after cancel")
	}
	if !strings.Contains(out.String(), "== counters final") {
		t.Fatalf("final counters missing: %q", out.String())
	}
}

func TestRunFailsOnBadRegexBeforeStartingGoroutines(t *testing.T) {
	err := Run(context.Background(), config.Config{Files: []string{"x.log"}, Patterns: []string{"("}, Interval: time.Millisecond}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}
