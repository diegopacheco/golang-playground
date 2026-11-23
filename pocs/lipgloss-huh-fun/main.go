package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"charm.land/lipgloss/v2"

	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/ansihtml"
	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/card"
	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/config"
	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/shots"
	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/wizard"
)

type options struct {
	name     string
	stack    string
	features string
	out      string
	capture  string
}

func main() {
	var o options
	flag.StringVar(&o.name, "name", "", "project name, setting it runs non-interactively")
	flag.StringVar(&o.stack, "stack", string(config.StackGo), "stack for non-interactive mode")
	flag.StringVar(&o.features, "features", "", "comma separated features for non-interactive mode")
	flag.StringVar(&o.out, "out", "", "write the config as JSON to this file")
	flag.StringVar(&o.capture, "capture", "", "write every wizard screen as HTML into this folder and exit")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, o); err != nil {
		if errors.Is(err, wizard.ErrCancelled) {
			fmt.Fprintln(os.Stderr, "cancelled, nothing written")
			os.Exit(130)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, o options) error {
	if o.capture != "" {
		return capture(o.capture)
	}
	cfg, err := collect(ctx, o)
	if err != nil {
		return err
	}
	lipgloss.Println(card.Render(cfg))
	if o.out == "" {
		return nil
	}
	return write(o.out, cfg)
}

func collect(ctx context.Context, o options) (config.Config, error) {
	if o.name == "" {
		return wizard.Run(ctx, &wizard.Answers{})
	}
	cfg, err := config.Parse(o.name, o.stack, o.features)
	if err != nil {
		lipgloss.Fprintln(os.Stderr, card.RenderError(err))
		return config.Config{}, fmt.Errorf("invalid flags: %w", err)
	}
	return cfg, nil
}

func write(path string, cfg config.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func capture(dir string) error {
	list, _, err := shots.Capture()
	if err != nil {
		return fmt.Errorf("capture screens: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create capture dir: %w", err)
	}
	for _, s := range list {
		page := ansihtml.Page(s.Title, s.Content, ansihtml.Dark)
		if err := os.WriteFile(filepath.Join(dir, s.Name+".html"), []byte(page), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", s.Name, err)
		}
	}
	return nil
}
