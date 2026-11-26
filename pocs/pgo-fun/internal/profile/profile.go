package profile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var ErrCapture = errors.New("cpu profile capture failed")

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func Capture(ctx context.Context, client Doer, baseURL string, seconds int, path string) error {
	url := fmt.Sprintf("%s/debug/pprof/profile?seconds=%d", baseURL, seconds)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCapture, err)
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCapture, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrCapture, res.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("%w: %w", ErrCapture, err)
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCapture, err)
	}
	n, err := io.Copy(f, res.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && n == 0 {
		err = errors.New("empty profile")
	}
	if err != nil {
		os.Remove(tmp)
		return fmt.Errorf("%w: %w", ErrCapture, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("%w: %w", ErrCapture, err)
	}
	return nil
}
