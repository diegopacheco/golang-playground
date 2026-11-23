package shots

import (
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/config"
)

func TestCaptureDrivesTheRealFormEndToEnd(t *testing.T) {
	shots, cfg, err := Capture()
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
	want := config.Config{Name: "billing-api", Stack: config.StackRust, Features: []config.Feature{config.FeatureDocker, config.FeatureCI, config.FeatureMetrics}}
	if cfg.Name != want.Name || cfg.Stack != want.Stack || !slices.Equal(cfg.Features, want.Features) {
		t.Fatalf("keystrokes produced %+v, want %+v", cfg, want)
	}
	if len(shots) != 8 {
		t.Fatalf("got %d shots, want 8", len(shots))
	}
	byName := map[string]string{}
	for _, s := range shots {
		if strings.TrimSpace(s.Content) == "" {
			t.Fatalf("shot %s is empty", s.Name)
		}
		byName[s.Name] = ansi.Strip(s.Content)
	}
	checks := []struct {
		name string
		shot string
		want string
	}{
		{"the form must block an invalid name and tell the user why", "02-name-invalid", config.ErrNameFormat.Error()},
		{"the stack step is reached only after a valid name", "04-stack", "Rust"},
		{"the confirm step previews what will be created", "06-confirm", "billing-api on rust with docker, ci, metrics"},
		{"the summary card reflects the confirmed answers", "07-summary", "metrics"},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(byName[c.shot], c.want) {
				t.Fatalf("%s is missing %q:\n%s", c.shot, c.want, byName[c.shot])
			}
		})
	}
}
