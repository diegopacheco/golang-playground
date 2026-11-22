package lines

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/diegopacheco/golang-playground/pocs/iter-seq-fun/internal/seq"
)

func logOf(total, errorEvery int) string {
	var b strings.Builder
	for i := 1; i <= total; i++ {
		level := "INFO"
		if i%errorEvery == 0 {
			level = "ERROR"
		}
		fmt.Fprintf(&b, "%s line %d\n", level, i)
	}
	return b.String()
}

func TestOnlyTheLinesNeededAreRead(t *testing.T) {
	tests := []struct {
		name         string
		take         int
		want         []string
		wantConsumed int
	}{
		{"one error stops after its line, not the whole file", 1, []string{"ERROR line 10"}, 10},
		{"three errors stop after the third match", 3, []string{"ERROR line 10", "ERROR line 20", "ERROR line 30"}, 30},
		{"zero take reads nothing from disk", 0, nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := New(strings.NewReader(logOf(10000, 10)))
			errorsOnly := seq.Filter(src.All(t.Context()), func(l string) bool { return strings.HasPrefix(l, "ERROR") })
			got := slices.Collect(seq.Take(errorsOnly, tt.take))
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			if src.Consumed() != tt.wantConsumed {
				t.Fatalf("consumed %d lines, want %d", src.Consumed(), tt.wantConsumed)
			}
			if src.Err() != nil {
				t.Fatalf("unexpected error: %v", src.Err())
			}
		})
	}
}

func TestReadFailuresAreReportedNotSwallowed(t *testing.T) {
	errDisk := errors.New("disk gone")
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	tests := []struct {
		name string
		ctx  context.Context
		src  *Source
		want error
	}{
		{"broken reader surfaces the IO error to the caller", t.Context(), New(iotest.ErrReader(errDisk)), errDisk},
		{"canceled context stops reading before the first line", canceled, New(strings.NewReader("a\nb\n")), context.Canceled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Collect(tt.src.All(tt.ctx))
			if len(got) != 0 {
				t.Fatalf("got lines %v, want none", got)
			}
			if !errors.Is(tt.src.Err(), tt.want) {
				t.Fatalf("got error %v, want %v", tt.src.Err(), tt.want)
			}
		})
	}
}
