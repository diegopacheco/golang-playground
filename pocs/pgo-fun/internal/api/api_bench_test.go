package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/order"
	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/store"
)

func benchBatch(b *testing.B, n int) []byte {
	b.Helper()
	batch := make([]order.Order, n)
	for i := range batch {
		batch[i] = order.Order{ID: "o", Customer: string(rune('a' + i%26)), Country: "US", Items: []order.Item{{SKU: "pen", Qty: 2, UnitCents: 150}, {SKU: "ink", Qty: 1, UnitCents: 990}}}
	}
	body, err := json.Marshal(batch)
	if err != nil {
		b.Fatal(err)
	}
	return body
}

func BenchmarkCreateOrder(b *testing.B) {
	h := New(store.NewMemory(), logx.New(io.Discard, slog.LevelError))
	body := []byte(validOrder)
	for b.Loop() {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body)))
		if rr.Code != http.StatusCreated {
			b.Fatalf("status %d", rr.Code)
		}
	}
}

func BenchmarkSummary100(b *testing.B) {
	h := New(store.NewMemory(), logx.New(io.Discard, slog.LevelError))
	body := benchBatch(b, 100)
	for b.Loop() {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/orders/summary", bytes.NewReader(body)))
		if rr.Code != http.StatusOK {
			b.Fatalf("status %d", rr.Code)
		}
	}
}
