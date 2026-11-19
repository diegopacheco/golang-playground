package receipt

import (
	"context"
	"fmt"
	"time"

	"github.com/diegopacheco/wails-fun/internal/apperr"
	"github.com/diegopacheco/wails-fun/internal/money"
	"github.com/diegopacheco/wails-fun/internal/sale"
)

type Method string

const (
	Cash     Method = "CASH"
	Card     Method = "CARD"
	Transfer Method = "TRANSFER"
)

func (m Method) valid() bool {
	return m == Cash || m == Card || m == Transfer
}

type Receipt struct {
	ID          int64       `json:"id"`
	Number      string      `json:"number"`
	SaleID      int64       `json:"saleId"`
	Customer    string      `json:"customer"`
	Method      Method      `json:"method"`
	AmountCents money.Cents `json:"amountCents"`
	IssuedAt    time.Time   `json:"issuedAt"`
}

type Input struct {
	SaleID int64  `json:"saleId"`
	Method Method `json:"method"`
}

func Number(id int64) string {
	return fmt.Sprintf("RCP-%06d", id)
}

type saleGetter interface {
	Get(ctx context.Context, id int64) (sale.Sale, error)
}

type store interface {
	Insert(ctx context.Context, r Receipt) (Receipt, error)
	List(ctx context.Context) ([]Receipt, error)
}

type Service struct {
	store store
	sales saleGetter
	now   func() time.Time
}

func NewService(s store, sales saleGetter, now func() time.Time) *Service {
	return &Service{store: s, sales: sales, now: now}
}

func (s *Service) Issue(ctx context.Context, in Input) (Receipt, error) {
	if !in.Method.valid() {
		return Receipt{}, fmt.Errorf("payment method %q: %w", in.Method, apperr.ErrInvalid)
	}
	sl, err := s.sales.Get(ctx, in.SaleID)
	if err != nil {
		return Receipt{}, fmt.Errorf("receipt sale: %w", err)
	}
	r, err := s.store.Insert(ctx, Receipt{
		SaleID:      sl.ID,
		Customer:    sl.Customer,
		Method:      in.Method,
		AmountCents: sl.TotalCents,
		IssuedAt:    s.now().UTC(),
	})
	if err != nil {
		return Receipt{}, err
	}
	r.Number = Number(r.ID)
	return r, nil
}

func (s *Service) List(ctx context.Context) ([]Receipt, error) {
	list, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].Number = Number(list[i].ID)
	}
	return list, nil
}
