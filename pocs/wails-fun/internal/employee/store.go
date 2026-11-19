package employee

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

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

func conflictOnUnique(err error) error {
	if db.IsUnique(err) {
		return fmt.Errorf("email already used: %w", apperr.ErrConflict)
	}
	return err
}

func (s *SQLStore) Insert(ctx context.Context, in Input, at time.Time) (Employee, error) {
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO employees(name, email, role, created_at) VALUES(?, ?, ?, ?)",
		in.Name, in.Email, in.Role, at)
	if err != nil {
		return Employee{}, fmt.Errorf("insert employee: %w", conflictOnUnique(err))
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Employee{}, fmt.Errorf("employee id: %w", err)
	}
	return Employee{ID: id, Name: in.Name, Email: in.Email, Role: in.Role, CreatedAt: at}, nil
}

func (s *SQLStore) Update(ctx context.Context, id int64, in Input) (Employee, error) {
	res, err := s.db.ExecContext(ctx,
		"UPDATE employees SET name = ?, email = ?, role = ? WHERE id = ?",
		in.Name, in.Email, in.Role, id)
	if err != nil {
		return Employee{}, fmt.Errorf("update employee %d: %w", id, conflictOnUnique(err))
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Employee{}, fmt.Errorf("update employee %d: %w", id, err)
	}
	if n == 0 {
		return Employee{}, fmt.Errorf("employee %d: %w", id, apperr.ErrNotFound)
	}
	return s.Get(ctx, id)
}

func (s *SQLStore) Get(ctx context.Context, id int64) (Employee, error) {
	var e Employee
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, email, role, created_at FROM employees WHERE id = ?", id).
		Scan(&e.ID, &e.Name, &e.Email, &e.Role, &e.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Employee{}, fmt.Errorf("employee %d: %w", id, apperr.ErrNotFound)
	}
	if err != nil {
		return Employee{}, fmt.Errorf("get employee %d: %w", id, err)
	}
	return e, nil
}

func (s *SQLStore) List(ctx context.Context) ([]Employee, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, email, role, created_at FROM employees ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()
	out := []Employee{}
	for rows.Next() {
		var e Employee
		if err := rows.Scan(&e.ID, &e.Name, &e.Email, &e.Role, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
