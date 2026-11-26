package loadgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync"
	"time"
)

var ErrSeed = errors.New("seed request failed")

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Config struct {
	BaseURL     string
	Concurrency int
	Warmup      time.Duration
	Duration    time.Duration
}

type Result struct {
	Requests int
	Errors   int
	Elapsed  time.Duration
	P50      time.Duration
	P99      time.Duration
}

func (r Result) RPS() float64 {
	if r.Elapsed <= 0 {
		return 0
	}
	return float64(r.Requests) / r.Elapsed.Seconds()
}

func Seed(ctx context.Context, client Doer, baseURL string, w Workload) error {
	for _, req := range w.Seeds {
		status, err := send(ctx, client, baseURL, req)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrSeed, err)
		}
		if status >= 300 {
			return fmt.Errorf("%w: %s %s returned %d", ErrSeed, req.Method, req.Path, status)
		}
	}
	return nil
}

func Run(ctx context.Context, client Doer, cfg Config, w Workload, onMeasure func()) (Result, error) {
	if len(w.Requests) == 0 || cfg.Concurrency < 1 {
		return Result{}, errors.New("workload needs requests and concurrency >= 1")
	}
	if cfg.Warmup > 0 {
		phase(ctx, client, cfg, w, cfg.Warmup)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("canceled during warmup: %w", err)
	}
	if onMeasure != nil {
		onMeasure()
	}
	start := time.Now()
	lat, errs := phase(ctx, client, cfg, w, cfg.Duration)
	res := Result{Requests: len(lat) + errs, Errors: errs, Elapsed: time.Since(start)}
	slices.Sort(lat)
	res.P50 = percentile(lat, 50)
	res.P99 = percentile(lat, 99)
	return res, nil
}

func phase(ctx context.Context, client Doer, cfg Config, w Workload, d time.Duration) ([]time.Duration, int) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	var (
		mu   sync.Mutex
		all  []time.Duration
		errs int
		wg   sync.WaitGroup
	)
	for worker := range cfg.Concurrency {
		wg.Go(func() {
			var lat []time.Duration
			failed := 0
			for i := worker; ctx.Err() == nil; i += cfg.Concurrency {
				req := w.Requests[i%len(w.Requests)]
				t0 := time.Now()
				status, err := send(ctx, client, cfg.BaseURL, req)
				if ctx.Err() != nil {
					break
				}
				if err != nil || status >= 400 {
					failed++
					continue
				}
				lat = append(lat, time.Since(t0))
			}
			mu.Lock()
			all = append(all, lat...)
			errs += failed
			mu.Unlock()
		})
	}
	wg.Wait()
	return all, errs
}

func send(ctx context.Context, client Doer, baseURL string, r Request) (int, error) {
	var body io.Reader
	if r.Body != nil {
		body = bytes.NewReader(r.Body)
	}
	req, err := http.NewRequestWithContext(ctx, r.Method, baseURL+r.Path, body)
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	if r.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%s %s: %w", r.Method, r.Path, err)
	}
	defer res.Body.Close()
	if _, err := io.Copy(io.Discard, res.Body); err != nil {
		return 0, fmt.Errorf("read body: %w", err)
	}
	return res.StatusCode, nil
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := (len(sorted)*p+99)/100 - 1
	return sorted[max(idx, 0)]
}
