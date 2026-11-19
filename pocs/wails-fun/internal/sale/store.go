package sale

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/diegopacheco/wails-fun/internal/apperr"
)

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

const selectSale = `SELECT s.id, s.employee_id, e.name, s.customer, s.total_cents, s.created_at
FROM sales s JOIN employees e ON e.id = s.employee_id`

func (s *SQLStore) Insert(ctx context.Context, sl Sale) (Sale, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Sale{}, fmt.Errorf("begin sale: %w", err)
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx,
		"INSERT INTO sales(employee_id, customer, total_cents, created_at) VALUES(?, ?, ?, ?)",
		sl.EmployeeID, sl.Customer, sl.TotalCents, sl.CreatedAt)
	if err != nil {
		return Sale{}, fmt.Errorf("insert sale: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Sale{}, fmt.Errorf("sale id: %w", err)
	}
	for _, it := range sl.Items {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO sale_items(sale_id, description, quantity, unit_price_cents) VALUES(?, ?, ?, ?)",
			id, it.Description, it.Quantity, it.UnitPriceCents); err != nil {
			return Sale{}, fmt.Errorf("insert sale item: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return Sale{}, fmt.Errorf("commit sale: %w", err)
	}
	sl.ID = id
	return sl, nil
}

func (s *SQLStore) Get(ctx context.Context, id int64) (Sale, error) {
	var sl Sale
	err := s.db.QueryRowContext(ctx, selectSale+" WHERE s.id = ?", id).
		Scan(&sl.ID, &sl.EmployeeID, &sl.EmployeeName, &sl.Customer, &sl.TotalCents, &sl.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Sale{}, fmt.Errorf("sale %d: %w", id, apperr.ErrNotFound)
	}
	if err != nil {
		return Sale{}, fmt.Errorf("get sale %d: %w", id, err)
	}
	items, err := s.items(ctx, id)
	if err != nil {
		return Sale{}, err
	}
	sl.Items = items
	return sl, nil
}

func (s *SQLStore) List(ctx context.Context) ([]Sale, error) {
	rows, err := s.db.QueryContext(ctx, selectSale+" ORDER BY s.id DESC")
	if err != nil {
		return nil, fmt.Errorf("list sales: %w", err)
	}
	defer rows.Close()
	out := []Sale{}
	for rows.Next() {
		var sl Sale
		if err := rows.Scan(&sl.ID, &sl.EmployeeID, &sl.EmployeeName, &sl.Customer, &sl.TotalCents, &sl.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan sale: %w", err)
		}
		out = append(out, sl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list sales: %w", err)
	}
	byID, err := s.allItems(ctx)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Items = byID[out[i].ID]
	}
	return out, nil
}

func (s *SQLStore) items(ctx context.Context, saleID int64) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT description, quantity, unit_price_cents FROM sale_items WHERE sale_id = ? ORDER BY id", saleID)
	if err != nil {
		return nil, fmt.Errorf("sale %d items: %w", saleID, err)
	}
	defer rows.Close()
	out := []Item{}
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.Description, &it.Quantity, &it.UnitPriceCents); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *SQLStore) allItems(ctx context.Context) (map[int64][]Item, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT sale_id, description, quantity, unit_price_cents FROM sale_items ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("sale items: %w", err)
	}
	defer rows.Close()
	out := map[int64][]Item{}
	for rows.Next() {
		var id int64
		var it Item
		if err := rows.Scan(&id, &it.Description, &it.Quantity, &it.UnitPriceCents); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		out[id] = append(out[id], it)
	}
	return out, rows.Err()
}
