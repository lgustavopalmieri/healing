package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/telemetry"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ObservabilityComponents holds all observability-related components
type ObservabilityComponents struct {
	TracerProvider *sdktrace.TracerProvider
	Tracer         observability.Tracer
	Logger         observability.Logger
	Metrics        observability.Metrics
	GRPCMetrics    *telemetry.GRPCMetrics
}

// InitObservability initializes all observability components (tracing, logging, metrics)
func InitObservability(ctx context.Context, cfg *config.Config) (*ObservabilityComponents, error) {
	log.Println("📊 Initializing observability (Tracing, Logging, Metrics)...")

	// Initialize Tracer Provider (OpenTelemetry -> Tempo)
	tracerProvider, err := telemetry.NewTracerProvider(
		ctx,
		cfg.Observability.ServiceName,
		cfg.Observability.OTLPEndpoint,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
	}

	// Create Tracer wrapper
	otelTracer := telemetry.GetTracer(cfg.Observability.ServiceName)
	tracer := telemetry.NewOtelTracer(otelTracer)

	// Initialize Logger (slog with JSON output)
	logger := telemetry.NewSlogLogger(cfg.Observability.ServiceName)

	// Initialize Prometheus Metrics
	prometheusMetrics := telemetry.NewPrometheusMetrics()

	// Initialize gRPC-specific metrics
	grpcMetrics := telemetry.NewGRPCMetrics(prometheusMetrics.Registry())

	log.Printf("✅ Observability initialized (Service: %s, Endpoint: %s)",
		cfg.Observability.ServiceName,
		cfg.Observability.OTLPEndpoint,
	)

	return &ObservabilityComponents{
		TracerProvider: tracerProvider,
		Tracer:         tracer,
		Logger:         logger,
		Metrics:        prometheusMetrics,
		GRPCMetrics:    grpcMetrics,
	}, nil
}
