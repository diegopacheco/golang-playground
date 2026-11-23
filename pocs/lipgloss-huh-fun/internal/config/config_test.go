package config

import (
	"errors"
	"slices"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  error
	}{
		{"kebab case names become folder and module names so they are accepted", "billing-api", nil},
		{"digits after the first letter are common in service names", "auth2-gateway", nil},
		{"an empty name would create a nameless project", "", ErrNameRequired},
		{"two letters is too short to be a meaningful project name", "ab", ErrNameLength},
		{"very long names break paths and docker tags", "a-very-long-project-name-that-keeps-going", ErrNameLength},
		{"uppercase breaks module paths on case sensitive systems", "BillingApi", ErrNameFormat},
		{"spaces and punctuation are not valid in folder or image names", "my app!", ErrNameFormat},
		{"a leading digit is not a valid package identifier", "9lives", ErrNameFormat},
		{"a trailing dash produces ugly and ambiguous names", "billing-", ErrNameFormat},
		{"double dashes hide typos so they are rejected", "billing--api", ErrNameFormat},
		{"names that clash with go tooling folders are reserved", "internal", ErrNameReserved},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if tt.want == nil && err != nil {
				t.Fatalf("ValidateName(%q) = %v, want nil", tt.input, err)
			}
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatalf("ValidateName(%q) = %v, want %v", tt.input, err, tt.want)
			}
		})
	}
}

func TestParseStack(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Stack
		wantErr error
	}{
		{"flags are forgiving about case and spaces typed in scripts", " Rust ", StackRust, nil},
		{"a supported stack maps to its typed constant", "typescript", StackTypeScript, nil},
		{"an unsupported stack must fail instead of generating a broken project", "cobol", "", ErrUnknownStack},
		{"an empty stack is not silently defaulted at the domain level", "", "", ErrUnknownStack},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStack(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseStack(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ParseStack(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFeatures(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Feature
		wantErr error
	}{
		{"no features is a valid minimal project", "", nil, nil},
		{"order is canonical so the same choices always give the same config", "metrics, docker,CI", []Feature{FeatureDocker, FeatureCI, FeatureMetrics}, nil},
		{"empty items from trailing commas are ignored", "lint,,", []Feature{FeatureLint}, nil},
		{"unknown features are rejected so typos are not lost", "docker,kubernetes", nil, ErrUnknownFeature},
		{"repeating a feature signals a mistake in the script", "ci,ci", nil, ErrDuplicateFeature},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFeatures(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseFeatures(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("ParseFeatures(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    [3]string
		want     Config
		wantErrs []error
	}{
		{"flags for a scripted run produce the same config the wizard would", [3]string{"billing-api", "go", "ci,docker"}, Config{Name: "billing-api", Stack: StackGo, Features: []Feature{FeatureDocker, FeatureCI}}, nil},
		{"all flag mistakes surface together so a script fails once with the full picture", [3]string{"Billing API", "cobol", "docker,k8s"}, Config{}, []error{ErrNameFormat, ErrUnknownStack, ErrUnknownFeature}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input[0], tt.input[1], tt.input[2])
			for _, want := range tt.wantErrs {
				if !errors.Is(err, want) {
					t.Fatalf("Parse error = %v, want it to include %v", err, want)
				}
			}
			if len(tt.wantErrs) == 0 && err != nil {
				t.Fatalf("Parse error = %v, want nil", err)
			}
			if got.Name != tt.want.Name || got.Stack != tt.want.Stack || !slices.Equal(got.Features, tt.want.Features) {
				t.Fatalf("Parse = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   []error
	}{
		{"a complete config from the wizard is valid", Config{Name: "billing-api", Stack: StackGo, Features: []Feature{FeatureDocker}}, nil},
		{"every problem is reported at once so the user fixes them in one pass", Config{Name: "Bad Name", Stack: "cobol", Features: []Feature{"k8s"}}, []error{ErrNameFormat, ErrUnknownStack, ErrUnknownFeature}},
		{"a missing stack is caught even when the name is fine", Config{Name: "billing-api"}, []error{ErrUnknownStack}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if len(tt.want) == 0 && err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
			for _, want := range tt.want {
				if !errors.Is(err, want) {
					t.Fatalf("Validate() = %v, want it to include %v", err, want)
				}
			}
		})
	}
}
