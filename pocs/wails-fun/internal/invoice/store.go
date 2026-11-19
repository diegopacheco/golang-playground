package invoice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/diegopacheco/wails-fun/internal/apperr"
	"github.com/diegopacheco/wails-fun/internal/db"
)

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(conn *sql.DB) *SQLStore {
	return &SQLStore{db: conn}
}

const selectInvoice = `SELECT id, sale_id, customer, amount_cents, status, issued_at, due_at, paid_at FROM invoices`

type scanner interface {
	Scan(dest ...any) error
}

func scan(row scanner) (Invoice, error) {
	var inv Invoice
	var paid sql.NullTime
	if err := row.Scan(&inv.ID, &inv.SaleID, &inv.Customer, &inv.AmountCents, &inv.Status, &inv.IssuedAt, &inv.DueAt, &paid); err != nil {
		return Invoice{}, err
	}
	if paid.Valid {
		inv.PaidAt = &paid.Time
	}
	return inv, nil
}

func (s *SQLStore) Insert(ctx context.Context, inv Invoice) (Invoice, error) {
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO invoices(sale_id, customer, amount_cents, status, issued_at, due_at) VALUES(?, ?, ?, ?, ?, ?)",
		inv.SaleID, inv.Customer, inv.AmountCents, inv.Status, inv.IssuedAt, inv.DueAt)
	if db.IsUnique(err) {
		return Invoice{}, fmt.Errorf("sale %d already has an invoice: %w", inv.SaleID, apperr.ErrConflict)
	}
	if err != nil {
		return Invoice{}, fmt.Errorf("insert invoice: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Invoice{}, fmt.Errorf("invoice id: %w", err)
	}
	inv.ID = id
	return inv, nil
}

func (s *SQLStore) Get(ctx context.Context, id int64) (Invoice, error) {
	inv, err := scan(s.db.QueryRowContext(ctx, selectInvoice+" WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return Invoice{}, fmt.Errorf("invoice %d: %w", id, apperr.ErrNotFound)
	}
	if err != nil {
		return Invoice{}, fmt.Errorf("get invoice %d: %w", id, err)
	}
	return inv, nil
}

func (s *SQLStore) SetStatus(ctx context.Context, id int64, status Status, paidAt *time.Time) error {
	if _, err := s.db.ExecContext(ctx, "UPDATE invoices SET status = ?, paid_at = ? WHERE id = ?", status, paidAt, id); err != nil {
		return fmt.Errorf("update invoice %d: %w", id, err)
	}
	return nil
}

func (s *SQLStore) List(ctx context.Context) ([]Invoice, error) {
	rows, err := s.db.QueryContext(ctx, selectInvoice+" ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}
	defer rows.Close()
	out := []Invoice{}
	for rows.Next() {
		inv, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan invoice: %w", err)
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}
