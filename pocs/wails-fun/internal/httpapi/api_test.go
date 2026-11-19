package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/diegopacheco/wails-fun/internal/db"
	"github.com/diegopacheco/wails-fun/internal/employee"
	"github.com/diegopacheco/wails-fun/internal/httpapi"
	"github.com/diegopacheco/wails-fun/internal/invoice"
	"github.com/diegopacheco/wails-fun/internal/receipt"
	"github.com/diegopacheco/wails-fun/internal/sale"
)

func newServer(t *testing.T, now time.Time) *httptest.Server {
	t.Helper()
	conn, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "biz.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	clock := func() time.Time { return now }
	employees := employee.NewService(employee.NewSQLStore(conn), clock)
	sales := sale.NewService(sale.NewSQLStore(conn), employees, clock)
	svc := httpapi.Services{
		Employees: employees,
		Sales:     sales,
		Receipts:  receipt.NewService(receipt.NewSQLStore(conn), sales, clock),
		Invoices:  invoice.NewService(invoice.NewSQLStore(conn), sales, clock),
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(httpapi.NewRouter(logger, svc, fstest.MapFS{}))
	t.Cleanup(srv.Close)
	return srv
}

func call[T any](t *testing.T, srv *httptest.Server, method, path string, body any, wantStatus int) T {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, srv.URL+path, &buf)
	if err != nil {
		t.Fatal(err)
	}
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != wantStatus {
		raw, _ := io.ReadAll(res.Body)
		t.Fatalf("%s %s = %d (%s), want %d", method, path, res.StatusCode, raw, wantStatus)
	}
	if res.Header.Get("X-Correlation-Id") == "" {
		t.Errorf("%s %s missing correlation id", method, path)
	}
	var out T
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

type errBody struct {
	Error string `json:"error"`
}

func TestSaleToDocumentsFlow(t *testing.T) {
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	srv := newServer(t, now)

	emp := call[employee.Employee](t, srv, "POST", "/api/employees",
		employee.Input{Name: "Ana", Email: "ANA@shop.com", Role: "Sales Rep"}, http.StatusCreated)
	if emp.Email != "ana@shop.com" {
		t.Errorf("email must be normalized to avoid duplicates by case, got %s", emp.Email)
	}
	call[errBody](t, srv, "POST", "/api/employees",
		employee.Input{Name: "Other", Email: "ana@shop.com", Role: "x"}, http.StatusConflict)

	call[errBody](t, srv, "POST", "/api/sales",
		sale.Input{EmployeeID: 999, Customer: "Acme", Items: []sale.Item{{Description: "pen", Quantity: 1, UnitPriceCents: 1}}}, http.StatusNotFound)

	sl := call[sale.Sale](t, srv, "POST", "/api/sales", sale.Input{
		EmployeeID: emp.ID,
		Customer:   "Acme",
		Items: []sale.Item{
			{Description: "Desk", Quantity: 2, UnitPriceCents: 15000},
			{Description: "Lamp", Quantity: 1, UnitPriceCents: 2599},
		},
	}, http.StatusCreated)
	if sl.TotalCents != 32599 {
		t.Fatalf("server must compute the total, got %d", sl.TotalCents)
	}
	stored := call[sale.Sale](t, srv, "GET", "/api/sales/1", nil, http.StatusOK)
	if len(stored.Items) != 2 || stored.EmployeeName != "Ana" || !stored.CreatedAt.Equal(now) {
		t.Fatalf("sale not persisted with items, employee and time: %+v", stored)
	}

	rc := call[receipt.Receipt](t, srv, "POST", "/api/receipts", receipt.Input{SaleID: sl.ID, Method: receipt.Card}, http.StatusCreated)
	if rc.Number != "RCP-000001" || rc.AmountCents != sl.TotalCents {
		t.Errorf("receipt must carry the sale total and a number, got %+v", rc)
	}
	call[errBody](t, srv, "POST", "/api/receipts", receipt.Input{SaleID: sl.ID, Method: receipt.Cash}, http.StatusConflict)
	call[errBody](t, srv, "POST", "/api/receipts", receipt.Input{SaleID: sl.ID, Method: "BITCOIN"}, http.StatusBadRequest)

	call[errBody](t, srv, "POST", "/api/invoices", invoice.Input{SaleID: sl.ID, DueDays: 0}, http.StatusBadRequest)
	inv := call[invoice.Invoice](t, srv, "POST", "/api/invoices", invoice.Input{SaleID: sl.ID, DueDays: 15}, http.StatusCreated)
	if inv.Status != invoice.Open || !inv.DueAt.Equal(now.AddDate(0, 0, 15)) {
		t.Fatalf("invoice must start open and due in 15 days, got %+v", inv)
	}
	paid := call[invoice.Invoice](t, srv, "POST", "/api/invoices/1/pay", nil, http.StatusOK)
	if paid.Status != invoice.Paid || paid.PaidAt == nil {
		t.Fatalf("paying must record status and time, got %+v", paid)
	}
	call[errBody](t, srv, "POST", "/api/invoices/1/void", nil, http.StatusConflict)

	list := call[[]invoice.Invoice](t, srv, "GET", "/api/invoices", nil, http.StatusOK)
	if len(list) != 1 || list[0].Status != invoice.Paid || list[0].PaidAt == nil {
		t.Fatalf("paid status must persist, got %+v", list)
	}
	call[errBody](t, srv, "GET", "/api/sales/abc", nil, http.StatusBadRequest)
}
