package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/telemetry"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ObservabilityComponents struct {
	TracerProvider *sdktrace.TracerProvider
	Tracer         observability.Tracer
	Logger         observability.Logger
	Metrics        observability.Metrics
	GRPCMetrics    *telemetry.GRPCMetrics
}

func InitObservability(ctx context.Context, cfg *config.Config) (*ObservabilityComponents, error) {
	log.Println("📊 Initializing observability (Tracing, Logging, Metrics)...")

	tracerProvider, err := telemetry.NewTracerProvider(
		ctx,
		cfg.Observability.ServiceName,
		cfg.Observability.OTLPEndpoint,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
	}

	otelTracer := telemetry.GetTracer(cfg.Observability.ServiceName)
	tracer := telemetry.NewOtelTracer(otelTracer)

	logger := telemetry.NewSlogLogger(cfg.Observability.ServiceName)

	prometheusMetrics := telemetry.NewPrometheusMetrics()

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
