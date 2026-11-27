package logx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
)

const HeaderCorrelationID = "X-Correlation-Id"

type ctxKey struct{}

type entry struct {
	id     string
	logger *slog.Logger
}

func parseLevel(level string) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return slog.LevelInfo
	}
	return l
}

func New(w io.Writer, level string) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: parseLevel(level)}))
}

func Raft(w io.Writer, name, level string) hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{Name: name, Output: w, Level: hclog.LevelFromString(strings.ToUpper(level))})
}

func WithCorrelationID(ctx context.Context, base *slog.Logger, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, entry{id: id, logger: base.With("correlation_id", id)})
}

func CorrelationID(ctx context.Context) string {
	if e, ok := ctx.Value(ctxKey{}).(entry); ok {
		return e.id
	}
	return ""
}

func FromContext(ctx context.Context) *slog.Logger {
	if e, ok := ctx.Value(ctxKey{}).(entry); ok {
		return e.logger
	}
	return slog.Default()
}

func NewID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type recorder struct {
	http.ResponseWriter
	status int
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func Middleware(base *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderCorrelationID)
		if id == "" {
			id = NewID()
		}
		w.Header().Set(HeaderCorrelationID, id)
		ctx := WithCorrelationID(r.Context(), base, id)
		rec := &recorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r.WithContext(ctx))
		FromContext(ctx).Info("http", "method", r.Method, "path", r.URL.Path, "status", rec.status, "duration", time.Since(start))
	})
}
