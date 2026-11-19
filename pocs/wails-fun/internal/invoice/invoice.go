package invoice

import (
	"context"
	"fmt"
	"time"

	"github.com/diegopacheco/wails-fun/internal/apperr"
	"github.com/diegopacheco/wails-fun/internal/money"
	"github.com/diegopacheco/wails-fun/internal/sale"
)

type Status string

const (
	Open Status = "OPEN"
	Paid Status = "PAID"
	Void Status = "VOID"
)

const (
	minDueDays = 1
	maxDueDays = 90
)

type Invoice struct {
	ID          int64       `json:"id"`
	Number      string      `json:"number"`
	SaleID      int64       `json:"saleId"`
	Customer    string      `json:"customer"`
	AmountCents money.Cents `json:"amountCents"`
	Status      Status      `json:"status"`
	Overdue     bool        `json:"overdue"`
	IssuedAt    time.Time   `json:"issuedAt"`
	DueAt       time.Time   `json:"dueAt"`
	PaidAt      *time.Time  `json:"paidAt"`
}

type Input struct {
	SaleID  int64 `json:"saleId"`
	DueDays int   `json:"dueDays"`
}

func Number(id int64) string {
	return fmt.Sprintf("INV-%06d", id)
}

func (inv Invoice) transition(to Status) (Invoice, error) {
	if inv.Status != Open {
		return Invoice{}, fmt.Errorf("invoice %s is %s and cannot become %s: %w", inv.Number, inv.Status, to, apperr.ErrConflict)
	}
	inv.Status = to
	return inv, nil
}

func (inv Invoice) view(now time.Time) Invoice {
	inv.Number = Number(inv.ID)
	inv.Overdue = inv.Status == Open && now.After(inv.DueAt)
	return inv
}

type saleGetter interface {
	Get(ctx context.Context, id int64) (sale.Sale, error)
}

type store interface {
	Insert(ctx context.Context, inv Invoice) (Invoice, error)
	Get(ctx context.Context, id int64) (Invoice, error)
	SetStatus(ctx context.Context, id int64, status Status, paidAt *time.Time) error
	List(ctx context.Context) ([]Invoice, error)
}

type Service struct {
	store store
	sales saleGetter
	now   func() time.Time
}

func NewService(s store, sales saleGetter, now func() time.Time) *Service {
	return &Service{store: s, sales: sales, now: now}
}

func (s *Service) Create(ctx context.Context, in Input) (Invoice, error) {
	if in.DueDays < minDueDays || in.DueDays > maxDueDays {
		return Invoice{}, fmt.Errorf("due days must be between %d and %d: %w", minDueDays, maxDueDays, apperr.ErrInvalid)
	}
	sl, err := s.sales.Get(ctx, in.SaleID)
	if err != nil {
		return Invoice{}, fmt.Errorf("invoice sale: %w", err)
	}
	now := s.now().UTC()
	inv, err := s.store.Insert(ctx, Invoice{
		SaleID:      sl.ID,
		Customer:    sl.Customer,
		AmountCents: sl.TotalCents,
		Status:      Open,
		IssuedAt:    now,
		DueAt:       now.AddDate(0, 0, in.DueDays),
	})
	if err != nil {
		return Invoice{}, err
	}
	return inv.view(now), nil
}

func (s *Service) Pay(ctx context.Context, id int64) (Invoice, error) {
	return s.move(ctx, id, Paid)
}

func (s *Service) Void(ctx context.Context, id int64) (Invoice, error) {
	return s.move(ctx, id, Void)
}

func (s *Service) move(ctx context.Context, id int64, to Status) (Invoice, error) {
	current, err := s.store.Get(ctx, id)
	if err != nil {
		return Invoice{}, err
	}
	now := s.now().UTC()
	next, err := current.view(now).transition(to)
	if err != nil {
		return Invoice{}, err
	}
	if to == Paid {
		next.PaidAt = &now
	}
	if err := s.store.SetStatus(ctx, id, next.Status, next.PaidAt); err != nil {
		return Invoice{}, err
	}
	return next.view(now), nil
}

func (s *Service) List(ctx context.Context) ([]Invoice, error) {
	list, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	for i := range list {
		list[i] = list[i].view(now)
	}
	return list, nil
}
