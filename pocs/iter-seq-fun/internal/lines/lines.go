package lines

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"iter"
)

type Source struct {
	scanner  *bufio.Scanner
	consumed int
	err      error
}

func New(r io.Reader) *Source {
	return &Source{scanner: bufio.NewScanner(r)}
}

func (s *Source) All(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		for {
			if err := ctx.Err(); err != nil {
				s.err = fmt.Errorf("read lines: %w", err)
				return
			}
			if !s.scanner.Scan() {
				break
			}
			s.consumed++
			if !yield(s.scanner.Text()) {
				return
			}
		}
		if err := s.scanner.Err(); err != nil {
			s.err = fmt.Errorf("read lines: %w", err)
		}
	}
}

func (s *Source) Consumed() int {
	return s.consumed
}

func (s *Source) Err() error {
	return s.err
}
