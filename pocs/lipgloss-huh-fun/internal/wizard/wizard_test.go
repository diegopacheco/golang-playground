package wizard

import (
	"slices"
	"testing"

	"github.com/diegopacheco/golang-playground/pocs/lipgloss-huh-fun/internal/config"
)

func TestAnswersConfigUsesCanonicalFeatureOrder(t *testing.T) {
	a := Answers{Name: "billing-api", Stack: config.StackGo, Features: []config.Feature{config.FeatureMetrics, config.FeatureDocker}}
	got := a.Config().Features
	want := []config.Feature{config.FeatureDocker, config.FeatureMetrics}
	if !slices.Equal(got, want) {
		t.Fatalf("toggle order leaked into the config: got %v, want %v", got, want)
	}
}

func TestNewFormDefaultsTheStackSoEnterOnStepTwoIsAlwaysValid(t *testing.T) {
	var a Answers
	NewForm(&a)
	if a.Stack != config.StackGo {
		t.Fatalf("default stack = %q, want %q", a.Stack, config.StackGo)
	}
}

func TestPreview(t *testing.T) {
	tests := []struct {
		name    string
		answers Answers
		want    string
	}{
		{"the confirm step spells out every choice before anything is created", Answers{Name: "billing-api", Stack: config.StackRust, Features: []config.Feature{config.FeatureCI, config.FeatureDocker}}, "billing-api on rust with docker, ci"},
		{"an empty feature list is stated explicitly", Answers{Name: "alpha-service", Stack: config.StackGo}, "alpha-service on go with no features"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := preview(tt.answers); got != tt.want {
				t.Fatalf("preview = %q, want %q", got, tt.want)
			}
		})
	}
}
