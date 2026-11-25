package logx

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		incoming string
		wantSame bool
	}{
		{name: "caller id is kept so logs join across services", incoming: "abc-123", wantSame: true},
		{name: "missing id is generated so every request is traceable in logs", incoming: ""},
		{name: "oversized id is replaced so clients cannot flood log lines", incoming: strings.Repeat("x", 200)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seen string
			h := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = CorrelationID(r.Context())
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.incoming != "" {
				req.Header.Set(CorrelationHeader, tt.incoming)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			echoed := rec.Header().Get(CorrelationHeader)
			if seen == "" || echoed != seen {
				t.Fatalf("context id %q and response id %q must match and be set", seen, echoed)
			}
			if tt.wantSame && seen != tt.incoming {
				t.Fatalf("got %q want %q", seen, tt.incoming)
			}
			if !tt.wantSame && seen == tt.incoming {
				t.Fatalf("expected a generated id, got %q", seen)
			}
		})
	}
}

func TestFromAddsCorrelationAndTraceSoLogsLinkToJaeger(t *testing.T) {
	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID})
	ctx := trace.ContextWithSpanContext(WithCorrelationID(context.Background(), "cid-1"), sc)

	var buf bytes.Buffer
	From(ctx, New(&buf)).Info("hello")

	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("log line is not json: %v", err)
	}
	if line["correlation_id"] != "cid-1" || line["trace_id"] != traceID.String() || line["span_id"] != spanID.String() {
		t.Fatalf("missing fields in %v", line)
	}
}
