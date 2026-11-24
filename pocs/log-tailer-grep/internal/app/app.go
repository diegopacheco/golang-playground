package app

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/log-tailer-grep/internal/config"
	"github.com/diegopacheco/golang-playground/pocs/log-tailer-grep/internal/counter"
	"github.com/diegopacheco/golang-playground/pocs/log-tailer-grep/internal/match"
	"github.com/diegopacheco/golang-playground/pocs/log-tailer-grep/internal/tail"
)

const prefixColor = "\x1b[90m"

type printer struct {
	w        io.Writer
	matcher  *match.Matcher
	counters *counter.Counters
	cfg      config.Config
	names    map[string]string
	start    time.Time
}

func Run(ctx context.Context, cfg config.Config, w io.Writer) error {
	matcher, err := match.New(cfg.Patterns, cfg.Color)
	if err != nil {
		return fmt.Errorf("build matcher: %w", err)
	}
	if cfg.Duration > 0 {
		var stop context.CancelFunc
		ctx, stop = context.WithTimeout(ctx, cfg.Duration)
		defer stop()
	}
	ctx, cancel := context.WithCancel(ctx)
	lines := make(chan tail.Line, 256)
	errc := make(chan error, len(cfg.Files))
	opts := tail.Options{Interval: cfg.Interval, FromStart: cfg.FromStart}
	var wg sync.WaitGroup
	for _, file := range cfg.Files {
		wg.Go(func() {
			if err := tail.Follow(ctx, file, opts, lines); err != nil {
				errc <- fmt.Errorf("follow %s: %w", file, err)
			}
		})
	}
	p := &printer{
		w:        w,
		matcher:  matcher,
		counters: counter.New(len(cfg.Patterns)),
		cfg:      cfg,
		names:    displayNames(cfg.Files),
		start:    time.Now(),
	}
	err = p.loop(ctx, lines, errc)
	cancel()
	wg.Wait()
	p.report("final")
	return err
}

func (p *printer) loop(ctx context.Context, lines <-chan tail.Line, errc <-chan error) error {
	var stats <-chan time.Time
	if p.cfg.Stats > 0 {
		ticker := time.NewTicker(p.cfg.Stats)
		defer ticker.Stop()
		stats = ticker.C
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errc:
			return err
		case <-stats:
			p.report(time.Since(p.start).Round(time.Second).String())
		case line := <-lines:
			p.print(line)
		}
	}
}

func (p *printer) print(line tail.Line) {
	res := p.matcher.Match(line.Text)
	p.counters.Record(res.Hits)
	if len(res.Hits) == 0 && len(p.cfg.Patterns) > 0 && !p.cfg.All {
		return
	}
	prefix := p.names[line.File] + " |"
	if p.cfg.Color {
		prefix = prefixColor + prefix + match.Reset
	}
	fmt.Fprintf(p.w, "%s %s\n", prefix, res.Text)
}

func (p *printer) report(label string) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "== counters %s lines=%d matched=%d", label, p.counters.Lines, p.counters.Matched)
	for i, pat := range p.matcher.Patterns() {
		fmt.Fprintf(&sb, " %s=%d", p.matcher.Paint(pat.Expr, i), p.counters.Hits[i])
	}
	fmt.Fprintln(p.w, sb.String())
}

func displayNames(files []string) map[string]string {
	width := 0
	for _, f := range files {
		width = max(width, len(filepath.Base(f)))
	}
	names := make(map[string]string, len(files))
	for _, f := range files {
		names[f] = fmt.Sprintf("%-*s", width, filepath.Base(f))
	}
	return names
}
