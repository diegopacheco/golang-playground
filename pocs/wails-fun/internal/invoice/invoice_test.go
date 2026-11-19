package invoice

import (
	"errors"
	"testing"
	"time"

	"github.com/diegopacheco/wails-fun/internal/apperr"
)

func TestTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    Status
		to      Status
		wantErr bool
	}{
		{"open invoices get paid", Open, Paid, false},
		{"open invoices can be cancelled", Open, Void, false},
		{"paying twice would double count revenue", Paid, Paid, true},
		{"a paid invoice cannot be voided after the money arrived", Paid, Void, true},
		{"a voided invoice cannot be collected", Void, Paid, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Invoice{ID: 1, Status: tt.from}.transition(tt.to)
			if tt.wantErr {
				if !errors.Is(err, apperr.ErrConflict) {
					t.Fatalf("want ErrConflict, got %v", err)
				}
				return
			}
			if err != nil || got.Status != tt.to {
				t.Fatalf("got %v %v, want %s", got.Status, err, tt.to)
			}
		})
	}
}

func TestOverdue(t *testing.T) {
	due := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		status Status
		now    time.Time
		want   bool
	}{
		{"open past the due date needs collection", Open, due.Add(time.Hour), true},
		{"open before the due date is fine", Open, due.Add(-time.Hour), false},
		{"paid invoices are never overdue", Paid, due.AddDate(0, 1, 0), false},
		{"voided invoices are never overdue", Void, due.AddDate(0, 1, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Invoice{ID: 7, Status: tt.status, DueAt: due}.view(tt.now)
			if got.Overdue != tt.want {
				t.Errorf("Overdue = %v, want %v", got.Overdue, tt.want)
			}
			if got.Number != "INV-000007" {
				t.Errorf("Number = %s", got.Number)
			}
		})
	}
}
