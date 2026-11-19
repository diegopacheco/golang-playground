package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/diegopacheco/wails-fun/internal/apperr"
)

type errorBody struct {
	Error string `json:"error"`
}

type responder struct {
	logger *slog.Logger
}

func (rs responder) json(ctx context.Context, w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		rs.logger.ErrorContext(ctx, "encode response", "err", err)
	}
}

func (rs responder) error(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		rs.json(ctx, w, http.StatusNotFound, errorBody{Error: err.Error()})
	case errors.Is(err, apperr.ErrInvalid):
		rs.json(ctx, w, http.StatusBadRequest, errorBody{Error: err.Error()})
	case errors.Is(err, apperr.ErrConflict):
		rs.json(ctx, w, http.StatusConflict, errorBody{Error: err.Error()})
	default:
		rs.logger.ErrorContext(ctx, "request failed", "err", err)
		rs.json(ctx, w, http.StatusInternalServerError, errorBody{Error: "internal error"})
	}
}

func (rs responder) reply(w http.ResponseWriter, r *http.Request, status int, v any, err error) {
	if err != nil {
		rs.error(r.Context(), w, err)
		return
	}
	rs.json(r.Context(), w, status, v)
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return v, fmt.Errorf("body: %v: %w", err, apperr.ErrInvalid)
	}
	return v, nil
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("id %q: %w", r.PathValue("id"), apperr.ErrInvalid)
	}
	return id, nil
}
