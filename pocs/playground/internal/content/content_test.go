package content

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/playground/internal/runner"
)

var goBlock = regexp.MustCompile("(?s)```go\n(.*?)```")

func newRunner(t *testing.T) runner.Runner {
	t.Helper()
	bin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not found")
	}
	return runner.Runner{GoBin: bin, Timeout: 30 * time.Second, MaxOutput: 64 * 1024}
}

func TestEverySampleRunsCleanlySoUsersNeverLoadBrokenCode(t *testing.T) {
	r := newRunner(t)
	samples, err := Samples()
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) < 10 {
		t.Fatalf("samples = %d, want at least 10", len(samples))
	}
	for _, s := range samples {
		t.Run(s.ID, func(t *testing.T) {
			t.Parallel()
			res, err := r.Run(context.Background(), s.Code)
			if err != nil {
				t.Fatal(err)
			}
			if res.ExitCode != 0 || res.Stdout == "" {
				t.Errorf("exit=%d stdout=%q stderr=%q", res.ExitCode, res.Stdout, res.Stderr)
			}
		})
	}
}

func TestEveryRunnableDocBlockRunsSoRunInPlaygroundWorks(t *testing.T) {
	r := newRunner(t)
	docs, err := Docs()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, d := range docs {
		for i, m := range goBlock.FindAllStringSubmatch(d.Markdown, -1) {
			if !strings.Contains(m[1], "package main") {
				continue
			}
			count++
			t.Run(d.ID+"/"+string(rune('a'+i)), func(t *testing.T) {
				t.Parallel()
				res, err := r.Run(context.Background(), m[1])
				if err != nil {
					t.Fatal(err)
				}
				if res.ExitCode != 0 {
					t.Errorf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
				}
			})
		}
	}
	if count < 15 {
		t.Errorf("runnable doc blocks = %d, want at least 15", count)
	}
}

func TestDocsHaveTitleSummaryAndConcurrencyTopics(t *testing.T) {
	docs, err := Docs()
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, d := range docs {
		ids[d.ID] = true
		if d.Title == "" || d.Summary == "" || strings.HasPrefix(d.Summary, "#") || strings.Contains(d.Summary, "`") {
			t.Errorf("doc %s title=%q summary=%q", d.ID, d.Title, d.Summary)
		}
	}
	for _, want := range []string{"goroutines", "channels", "buffered-channels", "select", "sync", "context", "worker-pools", "pipelines"} {
		if !ids[want] {
			t.Errorf("missing doc %q", want)
		}
	}
}

func TestDocByIDUnknownIsNotFound(t *testing.T) {
	if _, err := DocByID("nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
