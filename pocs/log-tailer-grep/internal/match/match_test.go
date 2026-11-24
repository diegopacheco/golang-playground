package match

import (
	"errors"
	"regexp/syntax"
	"slices"
	"testing"
)

func TestMatch(t *testing.T) {
	red, yellow := palette[0], palette[1]
	cases := []struct {
		name     string
		patterns []string
		color    bool
		line     string
		wantHits []int
		wantText string
	}{
		{
			name:     "a hit is wrapped in its pattern color so the eye finds it fast",
			patterns: []string{"ERROR"},
			color:    true,
			line:     "x ERROR y",
			wantHits: []int{0},
			wantText: "x " + red + "ERROR" + Reset + " y",
		},
		{
			name:     "each pattern keeps its own color so different signals stay distinct",
			patterns: []string{"ERROR", "timeout"},
			color:    true,
			line:     "ERROR db timeout",
			wantHits: []int{0, 1},
			wantText: red + "ERROR" + Reset + " db " + yellow + "timeout" + Reset,
		},
		{
			name:     "every occurrence is highlighted so repeated tokens are not missed",
			patterns: []string{"a+"},
			color:    true,
			line:     "aa b aaa",
			wantHits: []int{0},
			wantText: red + "aa" + Reset + " b " + red + "aaa" + Reset,
		},
		{
			name:     "overlapping matches keep the first span so escape codes never nest",
			patterns: []string{"ERROR", "ROR"},
			color:    true,
			line:     "ERROR",
			wantHits: []int{0, 1},
			wantText: red + "ERROR" + Reset,
		},
		{
			name:     "a line without hits is untouched so the caller can filter it",
			patterns: []string{"ERROR"},
			color:    true,
			line:     "all good",
			wantHits: nil,
			wantText: "all good",
		},
		{
			name:     "no-color still reports hits so counters work when piping to files",
			patterns: []string{"(?i)warn"},
			color:    false,
			line:     "Warn disk",
			wantHits: []int{0},
			wantText: "Warn disk",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := New(tc.patterns, tc.color)
			if err != nil {
				t.Fatalf("new: %v", err)
			}
			got := m.Match(tc.line)
			if !slices.Equal(got.Hits, tc.wantHits) {
				t.Fatalf("hits: got %v want %v", got.Hits, tc.wantHits)
			}
			if got.Text != tc.wantText {
				t.Fatalf("text: got %q want %q", got.Text, tc.wantText)
			}
		})
	}
}

func TestNewRejectsBadRegexSoTyposFailBeforeTailing(t *testing.T) {
	_, err := New([]string{"("}, true)
	var syntaxErr *syntax.Error
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("expected wrapped regexp syntax error, got %v", err)
	}
}

func BenchmarkMatch(b *testing.B) {
	m, err := New([]string{"ERROR", "WARN", `timeout after \d+ms`}, true)
	if err != nil {
		b.Fatal(err)
	}
	line := "2026-11-19T10:00:01 ERROR orders-api timeout after 532ms path=/api/orders"
	for b.Loop() {
		m.Match(line)
	}
}
