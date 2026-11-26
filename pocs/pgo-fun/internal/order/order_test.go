package order

import (
	"errors"
	"testing"
)

func TestPrice(t *testing.T) {
	items := func(qty int, cents int64) []Item { return []Item{{SKU: "A", Qty: qty, UnitCents: cents}} }
	tests := []struct {
		name     string
		in       Order
		subtotal int64
		discount int64
		tax      int64
		total    int64
		err      error
	}{
		{name: "US orders pay 7 percent tax on the subtotal", in: Order{ID: "1", Customer: "c", Country: "us", Items: items(2, 1000)}, subtotal: 2000, tax: 140, total: 2140},
		{name: "unknown countries fall back to the default 10 percent tax", in: Order{ID: "1", Customer: "c", Country: "ZZ", Items: items(1, 1000)}, subtotal: 1000, tax: 100, total: 1100},
		{name: "SAVE10 discounts before tax so customers are not taxed on money they did not pay", in: Order{ID: "1", Customer: "c", Country: "DE", Coupon: "SAVE10", Items: items(1, 10000)}, subtotal: 10000, discount: 1000, tax: 1710, total: 10710},
		{name: "BULK only applies from 10 items to protect margins on small orders", in: Order{ID: "1", Customer: "c", Country: "US", Coupon: "BULK", Items: items(9, 100)}, subtotal: 900, tax: 63, total: 963},
		{name: "BULK gives 5 percent once the order reaches 10 items", in: Order{ID: "1", Customer: "c", Country: "US", Coupon: "BULK", Items: items(10, 100)}, subtotal: 1000, discount: 50, tax: 67, total: 1017},
		{name: "tax rounds half up to the nearest cent", in: Order{ID: "1", Customer: "c", Country: "US", Items: items(1, 50)}, subtotal: 50, tax: 4, total: 54},
		{name: "an unknown coupon is rejected instead of silently charging full price", in: Order{ID: "1", Customer: "c", Coupon: "FREE", Items: items(1, 1)}, err: ErrUnknownCoupon},
		{name: "a zero quantity line is rejected because it would hide data entry errors", in: Order{ID: "1", Customer: "c", Items: items(0, 1)}, err: ErrInvalid},
		{name: "a negative price is rejected so totals cannot go below zero", in: Order{ID: "1", Customer: "c", Items: items(1, -1)}, err: ErrInvalid},
		{name: "an order without customer cannot be attributed and is rejected", in: Order{ID: "1", Items: items(1, 1)}, err: ErrInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Price(tt.in)
			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("want %v, got %v", tt.err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.SubtotalCents != tt.subtotal || p.DiscountCents != tt.discount || p.TaxCents != tt.tax || p.TotalCents != tt.total {
				t.Fatalf("got subtotal %d discount %d tax %d total %d", p.SubtotalCents, p.DiscountCents, p.TaxCents, p.TotalCents)
			}
		})
	}
}

func TestSummarize(t *testing.T) {
	orders := []Order{
		{ID: "1", Customer: "ana", Country: "US", Items: []Item{{SKU: "pen", Qty: 1, UnitCents: 100}, {SKU: "ink", Qty: 3, UnitCents: 10}}},
		{ID: "2", Customer: "bob", Country: "US", Items: []Item{{SKU: "book", Qty: 1, UnitCents: 5000}}},
		{ID: "3", Customer: "ana", Country: "US", Items: []Item{{SKU: "pen", Qty: 1, UnitCents: 100}}},
		{ID: "4", Customer: "cid", Country: "US", Items: []Item{{SKU: "book", Qty: 1, UnitCents: 5000}}},
	}
	got, err := Summarize(orders)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("want one row per customer, got %d", len(got))
	}
	if got[0].Customer != "bob" || got[1].Customer != "cid" {
		t.Fatalf("biggest spenders must come first and ties break by name for stable output, got %+v", got)
	}
	if got[2].Orders != 2 || got[2].TotalCents != 246 || got[2].TopSKU != "ink" {
		t.Fatalf("ana aggregates both orders and top sku is by quantity, got %+v", got[2])
	}
}

func TestSummarizeRejectsBatchWithInvalidOrder(t *testing.T) {
	_, err := Summarize([]Order{{ID: "x", Customer: "c", Coupon: "NOPE", Items: []Item{{SKU: "a", Qty: 1}}}})
	if !errors.Is(err, ErrUnknownCoupon) {
		t.Fatalf("a bad order must fail the batch so reports never undercount, got %v", err)
	}
}
