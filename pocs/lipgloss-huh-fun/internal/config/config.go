package config

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type Stack string

const (
	StackGo         Stack = "go"
	StackRust       Stack = "rust"
	StackJava       Stack = "java"
	StackTypeScript Stack = "typescript"
	StackPython     Stack = "python"
)

type Feature string

const (
	FeatureDocker  Feature = "docker"
	FeatureCI      Feature = "ci"
	FeatureLint    Feature = "lint"
	FeatureMetrics Feature = "metrics"
	FeatureAuth    Feature = "auth"
)

const (
	MinNameLength = 3
	MaxNameLength = 32
)

var (
	ErrNameRequired     = errors.New("name is required")
	ErrNameLength       = fmt.Errorf("name must be %d to %d characters", MinNameLength, MaxNameLength)
	ErrNameFormat       = errors.New("name must be lowercase kebab-case starting with a letter")
	ErrNameReserved     = errors.New("name is reserved")
	ErrUnknownStack     = errors.New("unknown stack")
	ErrUnknownFeature   = errors.New("unknown feature")
	ErrDuplicateFeature = errors.New("duplicate feature")
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

var reservedNames = []string{"main", "test", "internal", "vendor"}

type Config struct {
	Name     string    `json:"name"`
	Stack    Stack     `json:"stack"`
	Features []Feature `json:"features"`
}

func Stacks() []Stack {
	return []Stack{StackGo, StackRust, StackJava, StackTypeScript, StackPython}
}

func Features() []Feature {
	return []Feature{FeatureDocker, FeatureCI, FeatureLint, FeatureMetrics, FeatureAuth}
}

func ValidateName(name string) error {
	switch {
	case name == "":
		return ErrNameRequired
	case len(name) < MinNameLength || len(name) > MaxNameLength:
		return ErrNameLength
	case !namePattern.MatchString(name):
		return ErrNameFormat
	case slices.Contains(reservedNames, name):
		return fmt.Errorf("%w: %q", ErrNameReserved, name)
	}
	return nil
}

func ParseStack(s string) (Stack, error) {
	stack := Stack(strings.ToLower(strings.TrimSpace(s)))
	if !slices.Contains(Stacks(), stack) {
		return "", fmt.Errorf("%w: %q", ErrUnknownStack, s)
	}
	return stack, nil
}

func ParseFeatures(csv string) ([]Feature, error) {
	var features []Feature
	for part := range strings.SplitSeq(csv, ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" {
			features = append(features, Feature(part))
		}
	}
	if err := ValidateFeatures(features); err != nil {
		return nil, err
	}
	return Canonical(features), nil
}

func ValidateFeatures(features []Feature) error {
	seen := make(map[Feature]bool, len(features))
	for _, f := range features {
		if !slices.Contains(Features(), f) {
			return fmt.Errorf("%w: %q", ErrUnknownFeature, f)
		}
		if seen[f] {
			return fmt.Errorf("%w: %q", ErrDuplicateFeature, f)
		}
		seen[f] = true
	}
	return nil
}

func Canonical(features []Feature) []Feature {
	var out []Feature
	for _, f := range Features() {
		if slices.Contains(features, f) {
			out = append(out, f)
		}
	}
	return out
}

func Parse(name, stack, features string) (Config, error) {
	var errs []error
	if err := ValidateName(name); err != nil {
		errs = append(errs, err)
	}
	s, err := ParseStack(stack)
	if err != nil {
		errs = append(errs, err)
	}
	fs, err := ParseFeatures(features)
	if err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}
	return Config{Name: name, Stack: s, Features: fs}, nil
}

func (c Config) Validate() error {
	var errs []error
	if err := ValidateName(c.Name); err != nil {
		errs = append(errs, err)
	}
	if !slices.Contains(Stacks(), c.Stack) {
		errs = append(errs, fmt.Errorf("%w: %q", ErrUnknownStack, c.Stack))
	}
	if err := ValidateFeatures(c.Features); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
