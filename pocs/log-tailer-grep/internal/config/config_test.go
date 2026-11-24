package config

import (
	"errors"
	"io"
	"slices"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr error
		check   func(t *testing.T, c Config)
	}{
		{
			name: "repeated -e keeps order so pattern colors are stable between runs",
			args: []string{"-e", "ERROR", "-e", "WARN", "a.log", "b.log"},
			check: func(t *testing.T, c Config) {
				if !slices.Equal(c.Patterns, []string{"ERROR", "WARN"}) || !slices.Equal(c.Files, []string{"a.log", "b.log"}) {
					t.Fatalf("got %+v", c)
				}
			},
		},
		{
			name: "defaults color on so a terminal user sees highlights without flags",
			args: []string{"a.log"},
			check: func(t *testing.T, c Config) {
				if !c.Color || c.Interval != 200*time.Millisecond || c.Stats != 5*time.Second {
					t.Fatalf("got %+v", c)
				}
			},
		},
		{
			name: "no-color turns colors off so output can be piped to files",
			args: []string{"-no-color", "a.log"},
			check: func(t *testing.T, c Config) {
				if c.Color {
					t.Fatal("color should be off")
				}
			},
		},
		{
			name:    "no files is an error because there is nothing to follow",
			args:    []string{"-e", "x"},
			wantErr: ErrNoFiles,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := Parse(tc.args, io.Discard)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got %v want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			tc.check(t, c)
		})
	}
}
