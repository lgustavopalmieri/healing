package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/telemetry"
)

type ObservabilityComponents struct {
	Provider    *telemetry.Provider
	Factory     *telemetry.Factory
	GRPCMetrics *telemetry.GRPCMetrics
}

func InitObservability(ctx context.Context, cfg *config.Config) (*ObservabilityComponents, error) {
	log.Println("Initializing observability (Tracing, Logging, Metrics)...")

	provider, err := telemetry.NewProvider(ctx, telemetry.ProviderConfig{
		ServiceName:    cfg.Observability.ServiceName,
		ServiceVersion: cfg.Observability.ServiceVersion,
		Environment:    cfg.Observability.Environment,
		OTLPEndpoint:   cfg.Observability.OTLPEndpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize otel provider: %w", err)
	}

	factory := telemetry.NewFactory(cfg.Observability.ServiceName)
	grpcMetrics := telemetry.NewGRPCMetrics(cfg.Observability.ServiceName)

	log.Printf("Observability initialized (Service: %s, OTLP: %s)",
		cfg.Observability.ServiceName,
		cfg.Observability.OTLPEndpoint,
	)

	return &ObservabilityComponents{
		Provider:    provider,
		Factory:     factory,
		GRPCMetrics: grpcMetrics,
	}, nil
}
