package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/node"
)

type fakeStore struct {
	leader     bool
	leaderHTTP string
	setErr     error
	data       map[string]string
}

func (f *fakeStore) Get(key string) (string, bool) { v, ok := f.data[key]; return v, ok }
func (f *fakeStore) Set(_ context.Context, key, value string) error {
	if f.setErr != nil {
		return f.setErr
	}
	f.data[key] = value
	return nil
}
func (f *fakeStore) Delete(_ context.Context, key string) error {
	delete(f.data, key)
	return nil
}
func (f *fakeStore) Join(context.Context, node.JoinRequest) error { return nil }
func (f *fakeStore) Status() node.Status                          { return node.Status{} }
func (f *fakeStore) IsLeader() bool                               { return f.leader }
func (f *fakeStore) LeaderHTTP() (string, bool)                   { return f.leaderHTTP, f.leaderHTTP != "" }

func TestWritesOnlyLandOnTheLeader(t *testing.T) {
	tests := []struct {
		name     string
		store    *fakeStore
		method   string
		path     string
		wantCode int
		wantLoc  string
	}{
		{"leader applies the write itself", &fakeStore{leader: true}, http.MethodPut, "/kv/a", http.StatusOK, ""},
		{"follower redirects put to the leader keeping method and body", &fakeStore{leaderHTTP: "localhost:9101"}, http.MethodPut, "/kv/a", http.StatusTemporaryRedirect, "http://localhost:9101/kv/a"},
		{"follower redirects delete to the leader", &fakeStore{leaderHTTP: "localhost:9102"}, http.MethodDelete, "/kv/a", http.StatusTemporaryRedirect, "http://localhost:9102/kv/a"},
		{"follower redirects join so only the leader changes membership", &fakeStore{leaderHTTP: "localhost:9101"}, http.MethodPost, "/join", http.StatusTemporaryRedirect, "http://localhost:9101/join"},
		{"no known leader means unavailable rather than a lost write", &fakeStore{}, http.MethodPut, "/kv/a", http.StatusServiceUnavailable, ""},
		{"leadership lost mid write still redirects the client", &fakeStore{leader: true, leaderHTTP: "localhost:9103", setErr: fmt.Errorf("set: %w", node.ErrNotLeader)}, http.MethodPut, "/kv/a", http.StatusTemporaryRedirect, "http://localhost:9103/kv/a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.store.data = map[string]string{}
			rec := httptest.NewRecorder()
			New(tt.store).Handler().ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, strings.NewReader("v")))
			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d want %d body %s", rec.Code, tt.wantCode, rec.Body)
			}
			if got := rec.Header().Get("Location"); got != tt.wantLoc {
				t.Fatalf("Location = %q want %q", got, tt.wantLoc)
			}
		})
	}
}

func TestReadsAreServedLocallyOnAnyNode(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		wantCode int
	}{
		{"present key is returned by a follower without a leader round trip", map[string]string{"a": "1"}, http.StatusOK},
		{"missing key is a 404 so clients can tell absent from failure", map[string]string{}, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			New(&fakeStore{data: tt.data}).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/kv/a", nil))
			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d want %d", rec.Code, tt.wantCode)
			}
		})
	}
}
