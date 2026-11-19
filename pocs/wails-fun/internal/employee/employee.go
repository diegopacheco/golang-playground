package employee

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/diegopacheco/wails-fun/internal/apperr"
)

type Employee struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type Input struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (in Input) normalize() (Input, error) {
	out := Input{
		Name:  strings.TrimSpace(in.Name),
		Email: strings.ToLower(strings.TrimSpace(in.Email)),
		Role:  strings.TrimSpace(in.Role),
	}
	if out.Name == "" || out.Role == "" {
		return Input{}, fmt.Errorf("name and role are required: %w", apperr.ErrInvalid)
	}
	if _, err := mail.ParseAddress(out.Email); err != nil {
		return Input{}, fmt.Errorf("email %q: %w", in.Email, apperr.ErrInvalid)
	}
	return out, nil
}

type store interface {
	Insert(ctx context.Context, in Input, at time.Time) (Employee, error)
	Update(ctx context.Context, id int64, in Input) (Employee, error)
	Get(ctx context.Context, id int64) (Employee, error)
	List(ctx context.Context) ([]Employee, error)
}

type Service struct {
	store store
	now   func() time.Time
}

func NewService(s store, now func() time.Time) *Service {
	return &Service{store: s, now: now}
}

func (s *Service) Create(ctx context.Context, in Input) (Employee, error) {
	valid, err := in.normalize()
	if err != nil {
		return Employee{}, err
	}
	return s.store.Insert(ctx, valid, s.now().UTC())
}

func (s *Service) Update(ctx context.Context, id int64, in Input) (Employee, error) {
	valid, err := in.normalize()
	if err != nil {
		return Employee{}, err
	}
	return s.store.Update(ctx, id, valid)
}

func (s *Service) Get(ctx context.Context, id int64) (Employee, error) {
	return s.store.Get(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]Employee, error) {
	return s.store.List(ctx)
}
