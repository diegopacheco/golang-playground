package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/diegopacheco/golang-playground/pocs/playground/internal/complete"
	"github.com/diegopacheco/golang-playground/pocs/playground/internal/runner"
)

type fakeRunner struct{}

func (fakeRunner) Run(_ context.Context, code string) (runner.Result, error) {
	if strings.TrimSpace(code) == "" {
		return runner.Result{}, runner.ErrEmptyCode
	}
	return runner.Result{Stdout: "ran"}, nil
}

type fakeCompleter struct{}

func (fakeCompleter) Query(prefix string, _ int) []complete.Item {
	return []complete.Item{{Label: prefix + "X", Kind: complete.Func}}
}

func (fakeCompleter) Size() int { return 1 }

func TestEndpoints(t *testing.T) {
	mux := http.NewServeMux()
	New(fakeRunner{}, fakeCompleter{}, slog.New(slog.NewTextHandler(io.Discard, nil)), "go1.26").Routes(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	tests := []struct {
		name     string
		method   string
		path     string
		body     string
		wantCode int
		wantBody string
	}{
		{"run returns program output", "POST", "/api/run", `{"code":"package main"}`, 200, `"stdout":"ran"`},
		{"run with empty code is a client error", "POST", "/api/run", `{"code":"  "}`, 400, "code is empty"},
		{"run with broken json is a client error", "POST", "/api/run", `{`, 400, "bad request"},
		{"format returns gofmt output", "POST", "/api/format", `{"code":"package main\nfunc main(){x:=1;_=x}"}`, 200, "x := 1"},
		{"format reports syntax errors with line", "POST", "/api/format", `{"code":"package main\nfunc {"}`, 400, "2:"},
		{"complete passes the prefix", "GET", "/api/complete?prefix=fmt.P", "", 200, "fmt.PX"},
		{"docs lists topics", "GET", "/api/docs", "", 200, `"id":"channels"`},
		{"doc renders markdown to html without escaping tags", "GET", "/api/docs/channels", "", 200, "<h1 id="},
		{"unknown doc is 404", "GET", "/api/docs/nope", "", 404, "not found"},
		{"samples include code", "GET", "/api/samples", "", 200, `"id":"worker-pool"`},
		{"health lists services for the boot screen", "GET", "/api/health", "", 200, "Completion index"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, srv.URL+tt.path, strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Errorf("status = %d, want %d, body %s", resp.StatusCode, tt.wantCode, raw)
			}
			if !json.Valid(raw) {
				t.Errorf("body is not json: %s", raw)
			}
			if !strings.Contains(string(raw), tt.wantBody) {
				t.Errorf("body %s does not contain %q", raw, tt.wantBody)
			}
		})
	}
}

func TestRunRejectsGetSoBrowsersCannotTriggerRunsByLink(t *testing.T) {
	mux := http.NewServeMux()
	New(fakeRunner{}, fakeCompleter{}, slog.New(slog.NewTextHandler(io.Discard, nil)), "go1.26").Routes(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/run", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
