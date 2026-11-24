package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"time"
)

var ErrNoFiles = errors.New("at least one file is required")

type Config struct {
	Files     []string
	Patterns  []string
	Interval  time.Duration
	Stats     time.Duration
	Duration  time.Duration
	FromStart bool
	All       bool
	Color     bool
}

func Parse(args []string, stderr io.Writer) (Config, error) {
	var cfg Config
	noColor := false
	fs := flag.NewFlagSet("log-tailer-grep", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Func("e", "regex pattern to highlight and count, repeatable", func(v string) error {
		cfg.Patterns = append(cfg.Patterns, v)
		return nil
	})
	fs.DurationVar(&cfg.Interval, "interval", 200*time.Millisecond, "poll interval for each file")
	fs.DurationVar(&cfg.Stats, "stats", 5*time.Second, "print counters every interval, 0 disables periodic counters")
	fs.DurationVar(&cfg.Duration, "for", 0, "stop after this duration, 0 runs until Ctrl-C")
	fs.BoolVar(&cfg.FromStart, "from-start", false, "read files from the beginning instead of the end")
	fs.BoolVar(&cfg.All, "all", false, "print lines that match no pattern too")
	fs.BoolVar(&noColor, "no-color", false, "disable ANSI colors")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: log-tailer-grep [flags] FILE...")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}
	cfg.Files = fs.Args()
	cfg.Color = !noColor
	if len(cfg.Files) == 0 {
		fs.Usage()
		return Config{}, ErrNoFiles
	}
	if cfg.Interval <= 0 {
		return Config{}, fmt.Errorf("interval must be positive, got %s", cfg.Interval)
	}
	return cfg, nil
}
