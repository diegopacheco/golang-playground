package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/raft-kv/internal/node"
)

var ErrJoinRejected = errors.New("join rejected")

var client = &http.Client{Timeout: 10 * time.Second}

func Join(ctx context.Context, target string, req node.JoinRequest, attempts int) error {
	var last error
	for i := range attempts {
		if i > 0 {
			t := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				t.Stop()
				return fmt.Errorf("join %s: %w", target, ctx.Err())
			case <-t.C:
			}
		}
		if last = post(ctx, target, req); last == nil {
			return nil
		}
		logx.FromContext(ctx).Warn("join attempt failed", "attempt", i+1, "err", last)
	}
	return fmt.Errorf("join %s after %d attempts: %w", target, attempts, last)
}

func post(ctx context.Context, target string, req node.JoinRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("encode join: %w", err)
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+target+"/join", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build join request: %w", err)
	}
	r.Header.Set("Content-Type", "application/json")
	if id := logx.CorrelationID(ctx); id != "" {
		r.Header.Set(logx.HeaderCorrelationID, id)
	}
	resp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("post join: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%w: status %d: %s", ErrJoinRejected, resp.StatusCode, bytes.TrimSpace(msg))
	}
	return nil
}
