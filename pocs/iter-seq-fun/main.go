package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/diegopacheco/golang-playground/pocs/iter-seq-fun/internal/catalog"
	"github.com/diegopacheco/golang-playground/pocs/iter-seq-fun/internal/lines"
	"github.com/diegopacheco/golang-playground/pocs/iter-seq-fun/internal/pages"
	"github.com/diegopacheco/golang-playground/pocs/iter-seq-fun/internal/seq"
)

type config struct {
	file     string
	level    string
	products int
	pageSize int
	maxCents int
	take     int
}

func main() {
	var cfg config
	flag.StringVar(&cfg.file, "file", "data/app.log", "log file to scan")
	flag.StringVar(&cfg.level, "level", "ERROR", "log level to keep")
	flag.IntVar(&cfg.products, "products", 200, "products in the paginated catalog")
	flag.IntVar(&cfg.pageSize, "page-size", 10, "products per page")
	flag.IntVar(&cfg.maxCents, "max-cents", 120, "keep products cheaper than this")
	flag.IntVar(&cfg.take, "take", 3, "items to take from each pipeline")
	flag.Parse()

	if err := run(context.Background(), os.Stdout, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "iter-seq-fun: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer, cfg config) error {
	f, err := os.Open(cfg.file)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	defer f.Close()
	if err := runLines(ctx, out, f, cfg); err != nil {
		return err
	}
	return runPages(ctx, out, cfg)
}

func runLines(ctx context.Context, out io.Writer, r io.Reader, cfg config) error {
	src := lines.New(r)
	tag := " " + cfg.level + " "
	matching := seq.Filter(src.All(ctx), func(l string) bool { return strings.Contains(l, tag) })
	messages := seq.Map(matching, func(l string) string { return l[strings.Index(l, tag)+len(tag):] })
	fmt.Fprintf(out, "lines pipeline: Filter(%s) -> Map(message) -> Take(%d)\n", cfg.level, cfg.take)
	for n, msg := range seq.Zip(seq.From(1), seq.Take(messages, cfg.take)) {
		fmt.Fprintf(out, "  %d. %s\n", n, msg)
	}
	if err := src.Err(); err != nil {
		return err
	}
	fmt.Fprintf(out, "  lines read: %d (stopped as soon as %d matches were taken)\n\n", src.Consumed(), cfg.take)
	return nil
}

func runPages(ctx context.Context, out io.Writer, cfg config) error {
	store := catalog.NewStore(cfg.products, cfg.pageSize)
	pager := pages.New(store)
	cheap := seq.Filter(pager.All(ctx), func(p catalog.Product) bool { return p.PriceCents < cfg.maxCents })
	labels := seq.Map(cheap, func(p catalog.Product) string {
		return fmt.Sprintf("%s $%d.%02d", p.Name, p.PriceCents/100, p.PriceCents%100)
	})
	fmt.Fprintf(out, "pages pipeline: Filter(price < %d cents) -> Map(label) -> Take(%d)\n", cfg.maxCents, cfg.take)
	picked := slices.Collect(seq.Take(labels, cfg.take))
	if err := pager.Err(); err != nil {
		return err
	}
	for n, label := range seq.Zip(seq.From(1), slices.Values(picked)) {
		fmt.Fprintf(out, "  %d. %s\n", n, label)
	}
	fmt.Fprintf(out, "  pages fetched: %d of %d\n", store.Calls(), store.Pages())
	return nil
}
