package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/checkout"
	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/logx"
	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/telemetry"
)

const maxBodyBytes = 1 << 16

type Placer interface {
	Place(ctx context.Context, o checkout.Order) (checkout.Receipt, error)
}

type errorBody struct {
	Error         string `json:"error"`
	CorrelationID string `json:"correlation_id"`
}

type handler struct {
	placer Placer
	logger *slog.Logger
}

func NewRouter(placer Placer, logger *slog.Logger, prov *telemetry.Providers) http.Handler {
	h := handler{placer: placer, logger: logger}
	r := chi.NewRouter()
	r.Use(logx.Middleware, h.observe)
	r.Get("/health", h.health)
	r.Post("/api/orders", h.placeOrder)
	return otelhttp.NewHandler(r, "http.server",
		otelhttp.WithTracerProvider(prov.Tracer),
		otelhttp.WithMeterProvider(prov.Meter),
		otelhttp.WithPropagators(prov.Propagator),
		otelhttp.WithFilter(func(r *http.Request) bool { return r.URL.Path != "/health" }),
	)
}

func (h handler) observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unmatched"
		}
		span := trace.SpanFromContext(r.Context())
		span.SetName(r.Method + " " + route)
		span.SetAttributes(
			semconv.HTTPRoute(route),
			attribute.String("correlation.id", logx.CorrelationID(r.Context())),
		)
		logx.From(r.Context(), h.logger).Info("request completed",
			"method", r.Method,
			"route", route,
			"status", ww.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (h handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

func (h handler) placeOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var order checkout.Order
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&order); err != nil {
		h.fail(ctx, w, http.StatusBadRequest, "malformed order json", err)
		return
	}
	receipt, err := h.placer.Place(ctx, order)
	switch {
	case errors.Is(err, checkout.ErrInvalidOrder):
		h.fail(ctx, w, http.StatusUnprocessableEntity, err.Error(), err)
		return
	case err != nil:
		h.fail(ctx, w, http.StatusInternalServerError, "order could not be placed", err)
		return
	}
	logx.From(ctx, h.logger).Info("order placed", "order_id", receipt.OrderID, "total_cents", receipt.TotalCents)
	writeJSON(w, http.StatusCreated, receipt)
}

func (h handler) fail(ctx context.Context, w http.ResponseWriter, status int, msg string, err error) {
	logx.From(ctx, h.logger).Warn("order rejected", "status", status, "error", err)
	writeJSON(w, status, errorBody{Error: msg, CorrelationID: logx.CorrelationID(ctx)})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
