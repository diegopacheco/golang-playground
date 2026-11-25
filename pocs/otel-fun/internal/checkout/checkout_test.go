package checkout

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestPrice(t *testing.T) {
	tests := []struct {
		name         string
		order        Order
		wantDiscount int64
		wantTotal    int64
	}{
		{name: "below bulk quantity pays full price", order: Order{Item: "pen", Quantity: 9, UnitPriceCents: 100}, wantDiscount: 0, wantTotal: 900},
		{name: "exactly bulk quantity earns the discount so the threshold is inclusive", order: Order{Item: "pen", Quantity: 10, UnitPriceCents: 100}, wantDiscount: 100, wantTotal: 900},
		{name: "integer cents avoid float rounding drift on odd subtotals", order: Order{Item: "pen", Quantity: 11, UnitPriceCents: 333}, wantDiscount: 366, wantTotal: 3297},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Price(tt.order)
			if r.DiscountCents != tt.wantDiscount || r.TotalCents != tt.wantTotal {
				t.Fatalf("got discount %d total %d, want %d %d", r.DiscountCents, r.TotalCents, tt.wantDiscount, tt.wantTotal)
			}
			if r.SubtotalCents != r.TotalCents+r.DiscountCents {
				t.Fatalf("subtotal %d must equal total plus discount", r.SubtotalCents)
			}
		})
	}
}

func TestPlace(t *testing.T) {
	tests := []struct {
		name        string
		order       Order
		wantErr     bool
		wantOutcome string
		wantStatus  codes.Code
	}{
		{name: "accepted order is counted and its total lands in the histogram", order: Order{Item: "book", Quantity: 2, UnitPriceCents: 1500}, wantOutcome: "accepted", wantStatus: codes.Unset},
		{name: "empty item is rejected and marked as error so Jaeger highlights it", order: Order{Quantity: 1, UnitPriceCents: 10}, wantErr: true, wantOutcome: "rejected", wantStatus: codes.Error},
		{name: "zero quantity is rejected so no free orders are created", order: Order{Item: "book", UnitPriceCents: 10}, wantErr: true, wantOutcome: "rejected", wantStatus: codes.Error},
		{name: "huge unit price is rejected so totals cannot overflow", order: Order{Item: "book", Quantity: MaxQuantity, UnitPriceCents: MaxUnitPriceCents + 1}, wantErr: true, wantOutcome: "rejected", wantStatus: codes.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spans := tracetest.NewSpanRecorder()
			tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
			reader := sdkmetric.NewManualReader()
			mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			svc, err := New(tp, mp)
			if err != nil {
				t.Fatal(err)
			}

			parentCtx, parent := tp.Tracer("test").Start(context.Background(), "parent")
			_, err = svc.Place(parentCtx, tt.order)
			parent.End()

			if tt.wantErr != (err != nil) {
				t.Fatalf("err %v wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrInvalidOrder) {
				t.Fatalf("error %v must wrap ErrInvalidOrder so the API maps it to a client error", err)
			}

			ended := spans.Ended()
			if len(ended) != 2 || ended[0].Name() != PlaceOrderSpanName {
				t.Fatalf("expected business span then parent, got %d spans", len(ended))
			}
			child := ended[0]
			if child.Parent().SpanID() != parent.SpanContext().SpanID() {
				t.Fatal("business span must be a child of the request span")
			}
			if child.Status().Code != tt.wantStatus {
				t.Fatalf("status %v want %v", child.Status().Code, tt.wantStatus)
			}

			var rm metricdata.ResourceMetrics
			if err := reader.Collect(context.Background(), &rm); err != nil {
				t.Fatal(err)
			}
			if got := counterValue(rm, tt.wantOutcome); got != 1 {
				t.Fatalf("counter for %s = %d want 1", tt.wantOutcome, got)
			}
			wantHist := uint64(0)
			if !tt.wantErr {
				wantHist = 1
			}
			if got := histogramCount(rm); got != wantHist {
				t.Fatalf("histogram count %d want %d", got, wantHist)
			}
		})
	}
}

func counterValue(rm metricdata.ResourceMetrics, outcome string) int64 {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			sum, ok := m.Data.(metricdata.Sum[int64])
			if m.Name != OrdersMetric || !ok {
				continue
			}
			for _, dp := range sum.DataPoints {
				if v, _ := dp.Attributes.Value(attribute.Key("outcome")); v.AsString() == outcome {
					return dp.Value
				}
			}
		}
	}
	return 0
}

func histogramCount(rm metricdata.ResourceMetrics) uint64 {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			h, ok := m.Data.(metricdata.Histogram[int64])
			if m.Name != OrderTotalMetric || !ok {
				continue
			}
			var n uint64
			for _, dp := range h.DataPoints {
				n += dp.Count
			}
			return n
		}
	}
	return 0
}
