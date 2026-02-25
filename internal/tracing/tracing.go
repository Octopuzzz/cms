// Package tracing provides OpenTelemetry tracing integration.
package tracing

import (
	"context"

	"cms-backend/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

// InitTracer initializes the OpenTelemetry tracer
func InitTracer(cfg *config.JaegerConfig) (*sdktrace.TracerProvider, error) {
	if !cfg.Enabled {
		// Use a no-op tracer when disabled
		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		tracer = otel.Tracer(cfg.ServiceName)
		return tp, nil
	}

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.Endpoint)))
	if err != nil {
		return nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			attribute.String("environment", "production"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	tracer = otel.Tracer(cfg.ServiceName)
	return tp, nil
}

// StartSpan starts a new trace span
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	if tracer == nil {
		tracer = otel.Tracer("cms-backend")
	}
	return tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// Shutdown gracefully shuts down the tracer provider
func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider) {
	if tp != nil {
		_ = tp.Shutdown(ctx)
	}
}
