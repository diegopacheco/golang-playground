package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/order"
)

const maxBody = 4 << 20

type Store interface {
	Put(p order.Priced)
	Get(id string) (order.Priced, bool)
}

type Handler struct {
	store  Store
	logger *slog.Logger
}

func New(store Store, logger *slog.Logger) http.Handler {
	h := &Handler{store: store, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /orders", h.createOrder)
	mux.HandleFunc("GET /orders/{id}", h.getOrder)
	mux.HandleFunc("POST /orders/summary", h.summary)
	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	return logx.Middleware(logger, mux)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	h.write(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var o order.Order
	if !h.decode(w, r, &o) {
		return
	}
	p, err := order.Price(o)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	h.store.Put(p)
	h.write(w, r, http.StatusCreated, p)
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	p, ok := h.store.Get(r.PathValue("id"))
	if !ok {
		h.write(w, r, http.StatusNotFound, errorBody{Error: "order not found"})
		return
	}
	h.write(w, r, http.StatusOK, p)
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	var orders []order.Order
	if !h.decode(w, r, &orders) {
		return
	}
	s, err := order.Summarize(orders)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	h.write(w, r, http.StatusOK, s)
}

type errorBody struct {
	Error string `json:"error"`
}

func (h *Handler) decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		h.write(w, r, http.StatusBadRequest, errorBody{Error: "malformed json: " + err.Error()})
		return false
	}
	return true
}

func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, order.ErrInvalid), errors.Is(err, order.ErrUnknownCoupon):
		h.write(w, r, http.StatusUnprocessableEntity, errorBody{Error: err.Error()})
	default:
		logx.FromContext(r.Context(), h.logger).Error("request failed", slog.Any("error", err))
		h.write(w, r, http.StatusInternalServerError, errorBody{Error: "internal error"})
	}
}

func (h *Handler) write(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logx.FromContext(r.Context(), h.logger).Warn("write response", slog.Any("error", err))
	}
}
