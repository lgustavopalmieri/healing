package telemetry

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// NewTracerProvider creates and configures a new TracerProvider that exports traces via OTLP HTTP to Tempo
func NewTracerProvider(ctx context.Context, serviceName, tempoEndpoint string) (*sdktrace.TracerProvider, error) {
	// Create OTLP HTTP exporter configured for Tempo endpoint
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(tempoEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create TracerProvider with Batcher, Resource and AlwaysSample sampler
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global TracerProvider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// ShutdownTracer gracefully shuts down the TracerProvider, flushing any pending traces
func ShutdownTracer(ctx context.Context, tp *sdktrace.TracerProvider) {
	// Create context with timeout of 5 seconds
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Shutdown TracerProvider with error handling
	if err := tp.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	}
}

// GetTracer returns a tracer for the given service name
func GetTracer(serviceName string) trace.Tracer {
	return otel.Tracer(serviceName)
}
