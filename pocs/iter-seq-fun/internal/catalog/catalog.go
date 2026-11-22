package catalog

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/diegopacheco/golang-playground/pocs/iter-seq-fun/internal/pages"
)

var ErrBadCursor = errors.New("bad cursor")

type Product struct {
	ID         int
	Name       string
	PriceCents int
}

type Store struct {
	products []Product
	pageSize int
	calls    int
}

func NewStore(count, pageSize int) *Store {
	products := make([]Product, count)
	for i := range products {
		id := i + 1
		products[i] = Product{
			ID:         id,
			Name:       fmt.Sprintf("product-%03d", id),
			PriceCents: 100 + (id*37)%900,
		}
	}
	return &Store{products: products, pageSize: pageSize}
}

func (s *Store) Fetch(ctx context.Context, cursor string) (pages.Page[Product], error) {
	if err := ctx.Err(); err != nil {
		return pages.Page[Product]{}, fmt.Errorf("catalog fetch: %w", err)
	}
	start := 0
	if cursor != "" {
		n, err := strconv.Atoi(cursor)
		if err != nil || n < 0 || n > len(s.products) {
			return pages.Page[Product]{}, fmt.Errorf("cursor %q: %w", cursor, ErrBadCursor)
		}
		start = n
	}
	s.calls++
	end := min(start+s.pageSize, len(s.products))
	page := pages.Page[Product]{Items: s.products[start:end]}
	if end < len(s.products) {
		page.Next = strconv.Itoa(end)
	}
	return page, nil
}

func (s *Store) Pages() int {
	return (len(s.products) + s.pageSize - 1) / s.pageSize
}

func (s *Store) Calls() int {
	return s.calls
}
