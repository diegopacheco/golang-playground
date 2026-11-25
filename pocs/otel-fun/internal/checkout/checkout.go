package checkout

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	ScopeName          = "github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/checkout"
	BulkQuantity       = 10
	BulkDiscountPct    = 10
	MaxQuantity        = 1000
	MaxUnitPriceCents  = 10_000_000
	OrdersMetric       = "checkout.orders"
	OrderTotalMetric   = "checkout.order.total"
	PlaceOrderSpanName = "checkout.place_order"
)

var ErrInvalidOrder = errors.New("invalid order")

type Order struct {
	Item           string `json:"item"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
}

type Receipt struct {
	OrderID       string `json:"order_id"`
	Item          string `json:"item"`
	Quantity      int    `json:"quantity"`
	SubtotalCents int64  `json:"subtotal_cents"`
	DiscountCents int64  `json:"discount_cents"`
	TotalCents    int64  `json:"total_cents"`
}

type Service struct {
	tracer trace.Tracer
	orders metric.Int64Counter
	totals metric.Int64Histogram
	seq    atomic.Int64
}

func New(tp trace.TracerProvider, mp metric.MeterProvider) (*Service, error) {
	meter := mp.Meter(ScopeName)
	orders, err := meter.Int64Counter(OrdersMetric,
		metric.WithDescription("orders placed, by outcome"),
		metric.WithUnit("{order}"),
	)
	if err != nil {
		return nil, fmt.Errorf("create %s counter: %w", OrdersMetric, err)
	}
	totals, err := meter.Int64Histogram(OrderTotalMetric,
		metric.WithDescription("order total after discount"),
		metric.WithUnit("{cent}"),
		metric.WithExplicitBucketBoundaries(1000, 5000, 10000, 50000, 100000),
	)
	if err != nil {
		return nil, fmt.Errorf("create %s histogram: %w", OrderTotalMetric, err)
	}
	return &Service{tracer: tp.Tracer(ScopeName), orders: orders, totals: totals}, nil
}

func (s *Service) Place(ctx context.Context, o Order) (Receipt, error) {
	ctx, span := s.tracer.Start(ctx, PlaceOrderSpanName, trace.WithAttributes(
		attribute.String("order.item", o.Item),
		attribute.Int("order.quantity", o.Quantity),
	))
	defer span.End()

	if err := o.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		s.orders.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "rejected")))
		return Receipt{}, err
	}

	r := Price(o)
	r.OrderID = fmt.Sprintf("ord-%06d", s.seq.Add(1))
	span.SetAttributes(
		attribute.String("order.id", r.OrderID),
		attribute.Int64("order.total_cents", r.TotalCents),
		attribute.Bool("order.bulk_discount", r.DiscountCents > 0),
	)
	span.AddEvent("order.priced")
	s.orders.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "accepted")))
	s.totals.Record(ctx, r.TotalCents)
	return r, nil
}

func (o Order) Validate() error {
	switch {
	case strings.TrimSpace(o.Item) == "":
		return fmt.Errorf("%w: item is required", ErrInvalidOrder)
	case o.Quantity < 1 || o.Quantity > MaxQuantity:
		return fmt.Errorf("%w: quantity must be between 1 and %d", ErrInvalidOrder, MaxQuantity)
	case o.UnitPriceCents < 1 || o.UnitPriceCents > MaxUnitPriceCents:
		return fmt.Errorf("%w: unit_price_cents must be between 1 and %d", ErrInvalidOrder, MaxUnitPriceCents)
	}
	return nil
}

func Price(o Order) Receipt {
	subtotal := o.UnitPriceCents * int64(o.Quantity)
	var discount int64
	if o.Quantity >= BulkQuantity {
		discount = subtotal * BulkDiscountPct / 100
	}
	return Receipt{
		Item:          o.Item,
		Quantity:      o.Quantity,
		SubtotalCents: subtotal,
		DiscountCents: discount,
		TotalCents:    subtotal - discount,
	}
}
