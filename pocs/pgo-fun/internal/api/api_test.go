package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/order"
	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/store"
)

const validOrder = `{"id":"o-1","customer":"ana","country":"US","items":[{"sku":"pen","qty":2,"unit_cents":1000}]}`

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(New(store.NewMemory(), logx.New(io.Discard, slog.LevelError)))
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, srv *httptest.Server, method, path, body string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(logx.Header, "test-id")
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res, string(b)
}

func TestRoutes(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
		want   string
	}{
		{name: "health lets scripts know the service is ready", method: http.MethodGet, path: "/health", status: http.StatusOK, want: `"ok"`},
		{name: "creating an order returns the priced totals the client must charge", method: http.MethodPost, path: "/orders", body: validOrder, status: http.StatusCreated, want: `"total_cents":2140`},
		{name: "malformed json is a client error, not a server crash", method: http.MethodPost, path: "/orders", body: `{"id":`, status: http.StatusBadRequest, want: "malformed json"},
		{name: "unknown fields are rejected so typos in field names do not silently drop data", method: http.MethodPost, path: "/orders", body: `{"id":"1","customer":"c","itemz":[]}`, status: http.StatusBadRequest, want: "unknown field"},
		{name: "business rule violations map to 422 so clients can fix the payload", method: http.MethodPost, path: "/orders", body: `{"id":"1","customer":"c","coupon":"NOPE","items":[{"sku":"a","qty":1,"unit_cents":1}]}`, status: http.StatusUnprocessableEntity, want: "unknown coupon"},
		{name: "reading a missing order is 404 so clients can tell absent from broken", method: http.MethodGet, path: "/orders/nope", status: http.StatusNotFound, want: "order not found"},
		{name: "method patterns reject the wrong verb", method: http.MethodDelete, path: "/orders/o-1", status: http.StatusMethodNotAllowed},
		{name: "summary aggregates a batch per customer", method: http.MethodPost, path: "/orders/summary", body: "[" + validOrder + "," + validOrder + "]", status: http.StatusOK, want: `"orders":2`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newServer(t)
			res, body := do(t, srv, tt.method, tt.path, tt.body)
			if res.StatusCode != tt.status {
				t.Fatalf("want %d, got %d: %s", tt.status, res.StatusCode, body)
			}
			if !strings.Contains(body, tt.want) {
				t.Fatalf("body %q missing %q", body, tt.want)
			}
			if res.Header.Get(logx.Header) != "test-id" {
				t.Fatalf("every response must echo the correlation id")
			}
		})
	}
}

func TestCreatedOrderCanBeReadBack(t *testing.T) {
	srv := newServer(t)
	if res, body := do(t, srv, http.MethodPost, "/orders", validOrder); res.StatusCode != http.StatusCreated {
		t.Fatalf("create failed: %d %s", res.StatusCode, body)
	}
	res, body := do(t, srv, http.MethodGet, "/orders/o-1", "")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("stored order must be readable, got %d", res.StatusCode)
	}
	var p order.Priced
	if err := json.Unmarshal([]byte(body), &p); err != nil {
		t.Fatal(err)
	}
	if p.TotalCents != 2140 || p.Customer != "ana" {
		t.Fatalf("read back must match what was priced, got %+v", p)
	}
}
