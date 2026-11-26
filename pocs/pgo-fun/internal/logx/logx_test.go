package logx

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareCorrelationID(t *testing.T) {
	tests := []struct {
		name     string
		incoming string
		keep     bool
	}{
		{name: "an upstream id is kept so one request can be traced across services", incoming: "order-abc_123", keep: true},
		{name: "a missing id is generated so every log line is still traceable", incoming: "", keep: false},
		{name: "an id with unsafe characters is replaced to keep logs free of injected content", incoming: "bad id\n{json}", keep: false},
		{name: "an oversized id is replaced to bound log line size", incoming: strings.Repeat("a", 65), keep: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf, slog.LevelDebug)
			var seen string
			h := Middleware(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = CorrelationID(r.Context())
				w.WriteHeader(http.StatusTeapot)
			}))
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			if tt.incoming != "" {
				req.Header.Set(Header, tt.incoming)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			got := rr.Header().Get(Header)
			if got == "" || got != seen {
				t.Fatalf("response id %q and context id %q must match and be set", got, seen)
			}
			if tt.keep && got != tt.incoming {
				t.Fatalf("want kept id %q, got %q", tt.incoming, got)
			}
			if !tt.keep && got == tt.incoming {
				t.Fatalf("id %q should have been replaced", tt.incoming)
			}
			if !strings.Contains(buf.String(), `"correlation_id":"`+got+`"`) || !strings.Contains(buf.String(), `"status":418`) {
				t.Fatalf("request log missing id or status: %s", buf.String())
			}
		})
	}
}

func TestMiddlewareSkipsRequestLogAboveDebug(t *testing.T) {
	var buf bytes.Buffer
	h := Middleware(New(&buf, slog.LevelInfo), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
	if buf.Len() != 0 {
		t.Fatalf("info level must not log every request under load, got %s", buf.String())
	}
}
