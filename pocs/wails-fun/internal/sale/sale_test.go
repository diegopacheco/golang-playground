package sale

import (
	"errors"
	"testing"

	"github.com/diegopacheco/wails-fun/internal/apperr"
	"github.com/diegopacheco/wails-fun/internal/money"
)

func TestTotal(t *testing.T) {
	tests := []struct {
		name  string
		items []Item
		want  money.Cents
	}{
		{"quantity multiplies unit price so customers pay for every unit", []Item{{Description: "pen", Quantity: 3, UnitPriceCents: 150}}, 450},
		{"every line adds to the bill", []Item{{Description: "a", Quantity: 1, UnitPriceCents: 999}, {Description: "b", Quantity: 2, UnitPriceCents: 1}}, 1001},
		{"free items do not change the bill", []Item{{Description: "gift", Quantity: 5, UnitPriceCents: 0}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Total(tt.items); got != tt.want {
				t.Errorf("Total() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestInputValidation(t *testing.T) {
	ok := Item{Description: "pen", Quantity: 1, UnitPriceCents: 100}
	tests := []struct {
		name    string
		in      Input
		wantErr bool
	}{
		{"a sale with a customer and items is accepted", Input{Customer: " Acme ", Items: []Item{ok}}, false},
		{"a sale without a customer cannot be billed", Input{Customer: "  ", Items: []Item{ok}}, true},
		{"an empty sale makes no sense", Input{Customer: "Acme"}, true},
		{"zero quantity would hide a fake line", Input{Customer: "Acme", Items: []Item{{Description: "pen", Quantity: 0, UnitPriceCents: 100}}}, true},
		{"negative prices would turn a sale into a refund", Input{Customer: "Acme", Items: []Item{{Description: "pen", Quantity: 1, UnitPriceCents: -1}}}, true},
		{"an item needs a description to show on documents", Input{Customer: "Acme", Items: []Item{{Description: " ", Quantity: 1, UnitPriceCents: 1}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.normalize()
			if tt.wantErr {
				if !errors.Is(err, apperr.ErrInvalid) {
					t.Fatalf("want ErrInvalid, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if got.Customer != "Acme" {
				t.Errorf("customer not trimmed: %q", got.Customer)
			}
		})
	}
}
