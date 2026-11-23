package wizard

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"

	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/config"
)

const Width = 80

var ErrCancelled = errors.New("wizard cancelled")

type Answers struct {
	Name      string
	Stack     config.Stack
	Features  []config.Feature
	Confirmed bool
}

func (a Answers) Config() config.Config {
	return config.Config{Name: a.Name, Stack: a.Stack, Features: config.Canonical(a.Features)}
}

func NewForm(a *Answers) *huh.Form {
	if a.Stack == "" {
		a.Stack = config.StackGo
	}
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Description("Step 1 of 4 · lowercase kebab-case, 3 to 32 characters").
				Placeholder("billing-api").
				CharLimit(config.MaxNameLength).
				Validate(config.ValidateName).
				Value(&a.Name),
		),
		huh.NewGroup(
			huh.NewSelect[config.Stack]().
				Title("Stack").
				Description("Step 2 of 4 · the language the project is built with").
				Options(stackOptions()...).
				Height(len(config.Stacks())+2).
				Value(&a.Stack),
		),
		huh.NewGroup(
			huh.NewMultiSelect[config.Feature]().
				Title("Features").
				Description("Step 3 of 4 · space toggles, enter continues").
				Options(featureOptions()...).
				Height(len(config.Features())+2).
				Validate(config.ValidateFeatures).
				Value(&a.Features),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Create this project config?").
				DescriptionFunc(func() string { return "Step 4 of 4 · " + preview(*a) }, a).
				Affirmative("Create").
				Negative("Cancel").
				Value(&a.Confirmed),
		),
	).WithTheme(huh.ThemeFunc(huh.ThemeCharm)).WithWidth(Width).WithShowHelp(true)
}

func Run(ctx context.Context, a *Answers) (config.Config, error) {
	if err := NewForm(a).RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return config.Config{}, ErrCancelled
		}
		return config.Config{}, fmt.Errorf("run wizard: %w", err)
	}
	if !a.Confirmed {
		return config.Config{}, ErrCancelled
	}
	cfg := a.Config()
	if err := cfg.Validate(); err != nil {
		return config.Config{}, fmt.Errorf("invalid answers: %w", err)
	}
	return cfg, nil
}

func preview(a Answers) string {
	features := "no features"
	if len(a.Features) > 0 {
		names := make([]string, 0, len(a.Features))
		for _, f := range config.Canonical(a.Features) {
			names = append(names, string(f))
		}
		features = strings.Join(names, ", ")
	}
	return fmt.Sprintf("%s on %s with %s", a.Name, a.Stack, features)
}

func stackOptions() []huh.Option[config.Stack] {
	labels := map[config.Stack]string{
		config.StackGo:         "Go",
		config.StackRust:       "Rust",
		config.StackJava:       "Java",
		config.StackTypeScript: "TypeScript",
		config.StackPython:     "Python",
	}
	var opts []huh.Option[config.Stack]
	for _, s := range config.Stacks() {
		opts = append(opts, huh.NewOption(labels[s], s))
	}
	return opts
}

func featureOptions() []huh.Option[config.Feature] {
	labels := map[config.Feature]string{
		config.FeatureDocker:  "Docker image",
		config.FeatureCI:      "CI pipeline",
		config.FeatureLint:    "Linters",
		config.FeatureMetrics: "Prometheus metrics",
		config.FeatureAuth:    "Auth middleware",
	}
	var opts []huh.Option[config.Feature]
	for _, f := range config.Features() {
		opts = append(opts, huh.NewOption(labels[f], f))
	}
	return opts
}
