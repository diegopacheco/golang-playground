package profile

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCapture(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr bool
	}{
		{name: "a good profile is written where the pgo build expects it", status: http.StatusOK, body: "pprof-bytes"},
		{name: "a server error must not leave a profile that would mislead the compiler", status: http.StatusInternalServerError, body: "boom", wantErr: true},
		{name: "an empty profile is rejected because pgo=auto would silently do nothing", status: http.StatusOK, body: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/debug/pprof/profile" || r.URL.Query().Get("seconds") != "2" {
					t.Errorf("unexpected request %s", r.URL)
				}
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			path := filepath.Join(t.TempDir(), "sub", "cpu.pprof")
			err := Capture(context.Background(), srv.Client(), srv.URL, 2, path)
			_, statErr := os.Stat(path)
			if tt.wantErr {
				if !errors.Is(err, ErrCapture) || statErr == nil {
					t.Fatalf("want ErrCapture and no file, got %v and stat %v", err, statErr)
				}
				return
			}
			if err != nil || statErr != nil {
				t.Fatalf("unexpected error %v stat %v", err, statErr)
			}
		})
	}
}
