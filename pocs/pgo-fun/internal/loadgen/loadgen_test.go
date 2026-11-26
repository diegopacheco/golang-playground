package loadgen

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/api"
	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/store"
)

func TestPercentile(t *testing.T) {
	sorted := make([]time.Duration, 100)
	for i := range sorted {
		sorted[i] = time.Duration(i+1) * time.Millisecond
	}
	tests := []struct {
		name string
		in   []time.Duration
		p    int
		want time.Duration
	}{
		{name: "p50 of 1..100ms is 50ms, the nearest rank median", in: sorted, p: 50, want: 50 * time.Millisecond},
		{name: "p99 reports the tail a user actually waits on", in: sorted, p: 99, want: 99 * time.Millisecond},
		{name: "a single sample is every percentile", in: sorted[:1], p: 99, want: time.Millisecond},
		{name: "no samples reports zero instead of panicking", in: nil, p: 50, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := percentile(tt.in, tt.p); got != tt.want {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestWorkloadIsValidAgainstRealService(t *testing.T) {
	srv := httptest.NewServer(api.New(store.NewMemory(), logx.New(io.Discard, slog.LevelError)))
	defer srv.Close()
	w, err := NewWorkload(1, 8, 2, 20)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := Seed(ctx, srv.Client(), srv.URL, w); err != nil {
		t.Fatalf("seeding must succeed so GETs are hits, got %v", err)
	}
	res, err := Run(ctx, srv.Client(), Config{BaseURL: srv.URL, Concurrency: 4, Duration: 300 * time.Millisecond}, w, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Requests == 0 || res.Errors != 0 {
		t.Fatalf("a valid workload must produce traffic with zero errors or the benchmark measures error paths, got %+v", res)
	}
	if res.P50 <= 0 || res.P99 < res.P50 || res.RPS() <= 0 {
		t.Fatalf("latency stats must be ordered and positive, got %+v rps %.1f", res, res.RPS())
	}
}

func TestRunCountsServerErrors(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1)%2 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()
	w := Workload{Requests: []Request{{Method: http.MethodGet, Path: "/"}}}
	res, err := Run(context.Background(), srv.Client(), Config{BaseURL: srv.URL, Concurrency: 1, Duration: 200 * time.Millisecond}, w, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Errors == 0 || res.Errors == res.Requests {
		t.Fatalf("errors must be counted separately so a broken build cannot look fast, got %+v", res)
	}
}

func TestSeedFailsWhenServiceRejects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()
	w := Workload{Seeds: []Request{{Method: http.MethodPost, Path: "/orders", Body: []byte(strings.Repeat("x", 3))}}}
	if err := Seed(context.Background(), srv.Client(), srv.URL, w); !errors.Is(err, ErrSeed) {
		t.Fatalf("a rejected seed must abort the run, got %v", err)
	}
}

func TestRunStopsWhenCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := Workload{Requests: []Request{{Method: http.MethodGet, Path: "/"}}}
	_, err := Run(ctx, srv.Client(), Config{BaseURL: srv.URL, Concurrency: 2, Warmup: time.Second, Duration: time.Hour}, w, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Ctrl-C must end the run instead of hanging for the full duration, got %v", err)
	}
}
