package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/report"
)

func main() {
	in := flag.String("in", ".run/results.tsv", "results file written by loadgen")
	base := flag.String("base", "off", "baseline label")
	cand := flag.String("cand", "pgo", "candidate label")
	flag.Parse()
	if err := run(*in, *base, *cand); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func run(path, baseLabel, candLabel string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open results: %w", err)
	}
	defer f.Close()
	runs, err := report.Parse(f)
	if err != nil {
		return err
	}
	base, err := report.Summarize(runs, baseLabel)
	if err != nil {
		return err
	}
	cand, err := report.Summarize(runs, candLabel)
	if err != nil {
		return err
	}
	return report.Write(os.Stdout, runs, base, cand)
}
