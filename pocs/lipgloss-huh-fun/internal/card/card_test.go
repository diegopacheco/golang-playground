package card

import (
	"errors"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/config"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		want   []string
	}{
		{"the card shows every answer so the user can review before scaffolding", config.Config{Name: "billing-api", Stack: config.StackRust, Features: []config.Feature{config.FeatureDocker, config.FeatureCI}}, []string{"billing-api", "rust", "docker", "ci", "✓ valid"}},
		{"a project without features says none instead of leaving a blank row", config.Config{Name: "alpha-service", Stack: config.StackGo}, []string{"none", "✓ valid"}},
		{"a broken config never looks ready to scaffold", config.Config{Name: "Bad", Stack: config.StackGo}, []string{"✗ invalid"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := ansi.Strip(Render(tt.config))
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Fatalf("card is missing %q:\n%s", want, out)
				}
			}
			if strings.Contains(out, "ready to scaffold") == strings.Contains(out, "✗ invalid") {
				t.Fatalf("card must be either ready or invalid:\n%s", out)
			}
		})
	}
}

func TestRenderKeepsAFixedWidthSoTheBorderStaysAligned(t *testing.T) {
	short := Render(config.Config{Name: "abc", Stack: config.StackGo})
	long := Render(config.Config{Name: "a-much-longer-service-name", Stack: config.StackTypeScript, Features: config.Features()})
	if lipgloss.Width(short) != lipgloss.Width(long) {
		t.Fatalf("card widths differ: %d vs %d", lipgloss.Width(short), lipgloss.Width(long))
	}
}

func TestRenderErrorListsEveryProblemOnItsOwnLine(t *testing.T) {
	err := errors.Join(config.ErrNameFormat, config.ErrUnknownStack)
	out := ansi.Strip(RenderError(err))
	if strings.Count(out, "✗ ") != 2 {
		t.Fatalf("want two error lines:\n%s", out)
	}
}
