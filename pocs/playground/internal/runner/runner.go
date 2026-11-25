package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ErrEmptyCode = errors.New("code is empty")

type Result struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exitCode"`
	DurationMs int64  `json:"durationMs"`
	TimedOut   bool   `json:"timedOut"`
}

type Runner struct {
	GoBin     string
	Timeout   time.Duration
	MaxOutput int
}

func (r Runner) Run(ctx context.Context, code string) (Result, error) {
	if strings.TrimSpace(code) == "" {
		return Result{}, ErrEmptyCode
	}
	dir, err := os.MkdirTemp("", "goplay-")
	if err != nil {
		return Result{}, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0o600); err != nil {
		return Result{}, fmt.Errorf("write main.go: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module play\n\ngo 1.26\n"), 0o600); err != nil {
		return Result{}, fmt.Errorf("write go.mod: %w", err)
	}
	runCtx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	stdout := &capped{max: r.MaxOutput}
	stderr := &capped{max: r.MaxOutput}
	start := time.Now()
	runErr := r.exec(runCtx, dir, stdout, stderr, r.GoBin, "build", "-o", "prog", ".")
	if runErr == nil {
		runErr = r.exec(runCtx, dir, stdout, stderr, filepath.Join(dir, "prog"))
	}
	res := Result{
		Stdout:     stdout.String(),
		Stderr:     strings.ReplaceAll(stderr.String(), dir+string(filepath.Separator), ""),
		DurationMs: time.Since(start).Milliseconds(),
		TimedOut:   errors.Is(runCtx.Err(), context.DeadlineExceeded),
	}
	var exitErr *exec.ExitError
	switch {
	case runErr == nil:
	case errors.As(runErr, &exitErr):
		res.ExitCode = exitErr.ExitCode()
	case res.TimedOut:
		res.ExitCode = -1
	default:
		return Result{}, fmt.Errorf("run go: %w", runErr)
	}
	if res.TimedOut {
		res.Stderr += fmt.Sprintf("\nprogram killed: timeout after %s\n", r.Timeout)
	}
	return res, nil
}

func (r Runner) exec(ctx context.Context, dir string, stdout, stderr *capped, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = time.Second
	return cmd.Run()
}

type capped struct {
	mu        sync.Mutex
	buf       bytes.Buffer
	max       int
	truncated bool
}

func (c *capped) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	room := c.max - c.buf.Len()
	if room <= 0 {
		c.truncated = true
		return len(p), nil
	}
	if len(p) > room {
		c.buf.Write(p[:room])
		c.truncated = true
		return len(p), nil
	}
	c.buf.Write(p)
	return len(p), nil
}

func (c *capped) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.truncated {
		return c.buf.String() + "\n[output truncated]\n"
	}
	return c.buf.String()
}
