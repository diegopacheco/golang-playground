package content

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:embed docs/*.md samples/*/main.go
var files embed.FS

var ErrNotFound = errors.New("not found")

type Doc struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Markdown string `json:"-"`
}

type Sample struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Code  string `json:"code"`
}

var sampleOrder = []Sample{
	{ID: "hello", Title: "Hello, Go"},
	{ID: "goroutines", Title: "Goroutines with WaitGroup"},
	{ID: "ping-pong", Title: "Ping Pong over Channels"},
	{ID: "buffered-channels", Title: "Buffered Channel Queue"},
	{ID: "select-timeout", Title: "Select with Timeout"},
	{ID: "mutex-counter", Title: "Mutex Protected Map"},
	{ID: "atomic-counter", Title: "Atomic Counter"},
	{ID: "context-cancel", Title: "Context Cancellation"},
	{ID: "worker-pool", Title: "Worker Pool"},
	{ID: "pipeline", Title: "Pipeline Fan-out Fan-in"},
	{ID: "rate-limiter", Title: "Rate Limiter with Ticker"},
	{ID: "first-error", Title: "Cancel All on First Error"},
	{ID: "generics-channel", Title: "Generic Channel Helpers"},
}

func Docs() ([]Doc, error) {
	entries, err := fs.ReadDir(files, "docs")
	if err != nil {
		return nil, fmt.Errorf("read docs: %w", err)
	}
	docs := make([]Doc, 0, len(entries))
	for _, e := range entries {
		raw, err := files.ReadFile(path.Join("docs", e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read doc %s: %w", e.Name(), err)
		}
		docs = append(docs, parseDoc(e.Name(), string(raw)))
	}
	return docs, nil
}

func DocByID(id string) (Doc, error) {
	docs, err := Docs()
	if err != nil {
		return Doc{}, err
	}
	for _, d := range docs {
		if d.ID == id {
			return d, nil
		}
	}
	return Doc{}, fmt.Errorf("doc %q: %w", id, ErrNotFound)
}

func Samples() ([]Sample, error) {
	out := make([]Sample, 0, len(sampleOrder))
	for _, s := range sampleOrder {
		raw, err := files.ReadFile(path.Join("samples", s.ID, "main.go"))
		if err != nil {
			return nil, fmt.Errorf("read sample %s: %w", s.ID, err)
		}
		out = append(out, Sample{ID: s.ID, Title: s.Title, Code: string(raw)})
	}
	return out, nil
}

func parseDoc(name, raw string) Doc {
	id := strings.TrimSuffix(name, ".md")
	if _, rest, ok := strings.Cut(id, "-"); ok {
		id = rest
	}
	d := Doc{ID: id, Markdown: raw}
	for line := range strings.SplitSeq(raw, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case d.Title == "" && strings.HasPrefix(t, "# "):
			d.Title = strings.TrimPrefix(t, "# ")
		case d.Title != "" && d.Summary == "" && t != "":
			d.Summary = strings.ReplaceAll(t, "`", "")
		}
		if d.Summary != "" {
			break
		}
	}
	return d
}
