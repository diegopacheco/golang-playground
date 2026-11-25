package telemetry

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	"github.com/diegopacheco/golang-playground/pocs/otel-fun/internal/config"
)

type Providers struct {
	Tracer     *sdktrace.TracerProvider
	Meter      *sdkmetric.MeterProvider
	Propagator propagation.TextMapPropagator
}

func Setup(ctx context.Context, cfg config.Config, console io.Writer) (*Providers, error) {
	res := resource.NewSchemaless(semconv.ServiceName(cfg.ServiceName))

	otlp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}
	traceOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(otlp),
	}
	metricOpts := []sdkmetric.Option{sdkmetric.WithResource(res)}

	if cfg.ConsoleExport {
		consoleTraces, err := stdouttrace.New(stdouttrace.WithWriter(console))
		if err != nil {
			return nil, fmt.Errorf("create console trace exporter: %w", err)
		}
		consoleMetrics, err := stdoutmetric.New(stdoutmetric.WithWriter(console))
		if err != nil {
			return nil, fmt.Errorf("create console metric exporter: %w", err)
		}
		traceOpts = append(traceOpts, sdktrace.WithBatcher(consoleTraces))
		metricOpts = append(metricOpts, sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(consoleMetrics, sdkmetric.WithInterval(cfg.MetricInterval)),
		))
	}

	return &Providers{
		Tracer:     sdktrace.NewTracerProvider(traceOpts...),
		Meter:      sdkmetric.NewMeterProvider(metricOpts...),
		Propagator: propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	}, nil
}

func (p *Providers) Shutdown(ctx context.Context) error {
	return errors.Join(p.Tracer.Shutdown(ctx), p.Meter.Shutdown(ctx))
}
