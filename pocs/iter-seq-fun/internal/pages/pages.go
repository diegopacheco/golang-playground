package pages

import (
	"context"
	"fmt"
	"iter"
)

type Page[T any] struct {
	Items []T
	Next  string
}

type Fetcher[T any] interface {
	Fetch(ctx context.Context, cursor string) (Page[T], error)
}

type Paginator[T any] struct {
	fetcher Fetcher[T]
	fetched int
	err     error
}

func New[T any](f Fetcher[T]) *Paginator[T] {
	return &Paginator[T]{fetcher: f}
}

func (p *Paginator[T]) All(ctx context.Context) iter.Seq[T] {
	return func(yield func(T) bool) {
		cursor := ""
		for {
			page, err := p.fetcher.Fetch(ctx, cursor)
			if err != nil {
				p.err = fmt.Errorf("fetch page at cursor %q: %w", cursor, err)
				return
			}
			p.fetched++
			for _, item := range page.Items {
				if !yield(item) {
					return
				}
			}
			if page.Next == "" {
				return
			}
			cursor = page.Next
		}
	}
}

func (p *Paginator[T]) Fetched() int {
	return p.fetched
}

func (p *Paginator[T]) Err() error {
	return p.err
}
