package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/checkout"
	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/telemetry"
)

func newTestRouter(t *testing.T) (http.Handler, *tracetest.SpanRecorder) {
	t.Helper()
	spans := tracetest.NewSpanRecorder()
	prov := &telemetry.Providers{
		Tracer:     sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)),
		Meter:      sdkmetric.NewMeterProvider(),
		Propagator: propagation.TraceContext{},
	}
	svc, err := checkout.New(prov.Tracer, prov.Meter)
	if err != nil {
		t.Fatal(err)
	}
	return NewRouter(svc, logx.New(io.Discard), prov), spans
}

func TestOrdersEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "valid order is created so clients get a receipt", body: `{"item":"book","quantity":2,"unit_price_cents":1500}`, wantStatus: http.StatusCreated},
		{name: "business rule violation is 422 so clients know to fix the data", body: `{"item":"","quantity":2,"unit_price_cents":1500}`, wantStatus: http.StatusUnprocessableEntity},
		{name: "malformed json is 400 so it is not confused with a rule violation", body: `{"item":`, wantStatus: http.StatusBadRequest},
		{name: "unknown fields are 400 so typos in field names do not pass silently", body: `{"item":"book","qty":2}`, wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, _ := newTestRouter(t)
			req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(tt.body))
			req.Header.Set(logx.CorrelationHeader, "cid-42")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status %d want %d body %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if rec.Header().Get(logx.CorrelationHeader) != "cid-42" {
				t.Fatal("correlation id must be echoed so callers can find their logs")
			}
			if tt.wantStatus != http.StatusCreated {
				var body errorBody
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.CorrelationID != "cid-42" {
					t.Fatalf("error body must carry the correlation id, got %s", rec.Body.String())
				}
			}
		})
	}
}

func TestServerSpanParentsBusinessSpanSoJaegerShowsOneTree(t *testing.T) {
	router, spans := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(`{"item":"book","quantity":1,"unit_price_cents":100}`))
	req.Header.Set("traceparent", "00-0102030405060708090a0b0c0d0e0f10-0102030405060708-01")
	router.ServeHTTP(httptest.NewRecorder(), req)

	ended := spans.Ended()
	if len(ended) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(ended))
	}
	child, server := ended[0], ended[1]
	if server.Name() != "POST /api/orders" {
		t.Fatalf("server span name %q must use the chi route", server.Name())
	}
	if child.Name() != checkout.PlaceOrderSpanName || child.Parent().SpanID() != server.SpanContext().SpanID() {
		t.Fatal("business span must hang under the server span")
	}
	if server.SpanContext().TraceID().String() != "0102030405060708090a0b0c0d0e0f10" {
		t.Fatal("incoming traceparent must be continued so upstream traces connect")
	}
}

func TestHealthIsNotTracedSoProbesDoNotFloodJaeger(t *testing.T) {
	router, spans := newTestRouter(t)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if n := len(spans.Ended()); n != 0 {
		t.Fatalf("health produced %d spans", n)
	}
}
