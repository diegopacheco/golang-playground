package receipt

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/diegopacheco/wails-fun/internal/apperr"
	"github.com/diegopacheco/wails-fun/internal/db"
)

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(conn *sql.DB) *SQLStore {
	return &SQLStore{db: conn}
}

func (s *SQLStore) Insert(ctx context.Context, r Receipt) (Receipt, error) {
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO receipts(sale_id, method, amount_cents, issued_at) VALUES(?, ?, ?, ?)",
		r.SaleID, r.Method, r.AmountCents, r.IssuedAt)
	if db.IsUnique(err) {
		return Receipt{}, fmt.Errorf("sale %d already has a receipt: %w", r.SaleID, apperr.ErrConflict)
	}
	if err != nil {
		return Receipt{}, fmt.Errorf("insert receipt: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Receipt{}, fmt.Errorf("receipt id: %w", err)
	}
	r.ID = id
	return r, nil
}

func (s *SQLStore) List(ctx context.Context) ([]Receipt, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT r.id, r.sale_id, s.customer, r.method, r.amount_cents, r.issued_at
FROM receipts r JOIN sales s ON s.id = r.sale_id ORDER BY r.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list receipts: %w", err)
	}
	defer rows.Close()
	out := []Receipt{}
	for rows.Next() {
		var r Receipt
		if err := rows.Scan(&r.ID, &r.SaleID, &r.Customer, &r.Method, &r.AmountCents, &r.IssuedAt); err != nil {
			return nil, fmt.Errorf("scan receipt: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
