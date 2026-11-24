package tail

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type scenario struct {
	name      string
	fromStart bool
	setup     func(t *testing.T, path string)
	act       func(t *testing.T, path string)
	want      []string
}

func appendTo(t *testing.T, path, text string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func collect(t *testing.T, ch <-chan Line, n int) []string {
	t.Helper()
	var got []string
	deadline := time.After(3 * time.Second)
	for len(got) < n {
		select {
		case l := <-ch:
			got = append(got, l.Text)
		case <-deadline:
			t.Fatalf("timed out after %d of %d lines: %q", len(got), n, got)
		}
	}
	return got
}

func settle() {
	time.Sleep(60 * time.Millisecond)
}

func TestFollow(t *testing.T) {
	cases := []scenario{
		{
			name:  "starting at the end skips history so old noise is not replayed",
			setup: func(t *testing.T, p string) { writeFile(t, p, "old 1\nold 2\n") },
			act:   func(t *testing.T, p string) { appendTo(t, p, "new 1\nnew 2\n") },
			want:  []string{"new 1", "new 2"},
		},
		{
			name:      "from-start replays history so a whole file can be grepped",
			fromStart: true,
			setup:     func(t *testing.T, p string) { writeFile(t, p, "old 1\n") },
			act:       func(t *testing.T, p string) { appendTo(t, p, "new 1\n") },
			want:      []string{"old 1", "new 1"},
		},
		{
			name:  "partial lines wait for the newline so a line is never split in two",
			setup: func(t *testing.T, p string) { writeFile(t, p, "") },
			act: func(t *testing.T, p string) {
				appendTo(t, p, "half ")
				settle()
				appendTo(t, p, "done\r\n")
			},
			want: []string{"half done"},
		},
		{
			name:  "truncation restarts from offset zero so copytruncate rotation keeps flowing",
			setup: func(t *testing.T, p string) { writeFile(t, p, "") },
			act: func(t *testing.T, p string) {
				appendTo(t, p, "before truncate long line\n")
				settle()
				writeFile(t, p, "after\n")
			},
			want: []string{"before truncate long line", "after"},
		},
		{
			name:  "rename rotation switches to the new file so logrotate create mode keeps flowing",
			setup: func(t *testing.T, p string) { writeFile(t, p, "") },
			act: func(t *testing.T, p string) {
				appendTo(t, p, "in old file\n")
				settle()
				if err := os.Rename(p, p+".1"); err != nil {
					t.Fatalf("rename: %v", err)
				}
				settle()
				writeFile(t, p, "in new file\n")
			},
			want: []string{"in old file", "in new file"},
		},
		{
			name:  "a missing file is picked up once created so tailing can start before the app",
			setup: func(t *testing.T, p string) {},
			act:   func(t *testing.T, p string) { writeFile(t, p, "first\n") },
			want:  []string{"first"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "app.log")
			tc.setup(t, path)
			ctx, cancel := context.WithCancel(context.Background())
			lines := make(chan Line, 16)
			done := make(chan error, 1)
			go func() {
				done <- Follow(ctx, path, Options{Interval: 10 * time.Millisecond, FromStart: tc.fromStart}, lines)
			}()
			settle()
			tc.act(t, path)
			got := collect(t, lines, len(tc.want))
			cancel()
			if err := <-done; err != nil {
				t.Fatalf("follow returned error: %v", err)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("line %d: got %q want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestFollowExitsWhenContextIsCancelledEvenIfNobodyReads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	writeFile(t, path, "a\nb\nc\n")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Follow(ctx, path, Options{Interval: 10 * time.Millisecond, FromStart: true}, make(chan Line))
	}()
	settle()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("follow returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("follow goroutine leaked after cancel")
	}
}
