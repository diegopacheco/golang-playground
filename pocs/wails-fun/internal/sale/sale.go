package sale

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/diegopacheco/wails-fun/internal/apperr"
	"github.com/diegopacheco/wails-fun/internal/employee"
	"github.com/diegopacheco/wails-fun/internal/money"
)

type Item struct {
	Description    string      `json:"description"`
	Quantity       int64       `json:"quantity"`
	UnitPriceCents money.Cents `json:"unitPriceCents"`
}

func (i Item) Subtotal() money.Cents {
	return money.Cents(i.Quantity) * i.UnitPriceCents
}

type Sale struct {
	ID           int64       `json:"id"`
	EmployeeID   int64       `json:"employeeId"`
	EmployeeName string      `json:"employeeName"`
	Customer     string      `json:"customer"`
	Items        []Item      `json:"items"`
	TotalCents   money.Cents `json:"totalCents"`
	CreatedAt    time.Time   `json:"createdAt"`
}

type Input struct {
	EmployeeID int64  `json:"employeeId"`
	Customer   string `json:"customer"`
	Items      []Item `json:"items"`
}

func Total(items []Item) money.Cents {
	var total money.Cents
	for _, it := range items {
		total += it.Subtotal()
	}
	return total
}

func (in Input) normalize() (Input, error) {
	out := Input{EmployeeID: in.EmployeeID, Customer: strings.TrimSpace(in.Customer)}
	if out.Customer == "" {
		return Input{}, fmt.Errorf("customer is required: %w", apperr.ErrInvalid)
	}
	if len(in.Items) == 0 {
		return Input{}, fmt.Errorf("a sale needs at least one item: %w", apperr.ErrInvalid)
	}
	for _, it := range in.Items {
		desc := strings.TrimSpace(it.Description)
		if desc == "" || it.Quantity <= 0 || it.UnitPriceCents < 0 {
			return Input{}, fmt.Errorf("item %q needs a description, quantity > 0 and price >= 0: %w", it.Description, apperr.ErrInvalid)
		}
		out.Items = append(out.Items, Item{Description: desc, Quantity: it.Quantity, UnitPriceCents: it.UnitPriceCents})
	}
	return out, nil
}

type employeeGetter interface {
	Get(ctx context.Context, id int64) (employee.Employee, error)
}

type store interface {
	Insert(ctx context.Context, s Sale) (Sale, error)
	Get(ctx context.Context, id int64) (Sale, error)
	List(ctx context.Context) ([]Sale, error)
}

type Service struct {
	store     store
	employees employeeGetter
	now       func() time.Time
}

func NewService(s store, employees employeeGetter, now func() time.Time) *Service {
	return &Service{store: s, employees: employees, now: now}
}

func (s *Service) Create(ctx context.Context, in Input) (Sale, error) {
	valid, err := in.normalize()
	if err != nil {
		return Sale{}, err
	}
	emp, err := s.employees.Get(ctx, valid.EmployeeID)
	if err != nil {
		return Sale{}, fmt.Errorf("sale employee: %w", err)
	}
	return s.store.Insert(ctx, Sale{
		EmployeeID:   emp.ID,
		EmployeeName: emp.Name,
		Customer:     valid.Customer,
		Items:        valid.Items,
		TotalCents:   Total(valid.Items),
		CreatedAt:    s.now().UTC(),
	})
}

func (s *Service) Get(ctx context.Context, id int64) (Sale, error) {
	return s.store.Get(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Sale, error) {
	return s.store.List(ctx)
}
