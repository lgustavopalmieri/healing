package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/grpcserver/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/opentelemetry"
)

func InitObservability(ctx context.Context, cfg *config.Config) (*opentelemetry.GrafanaProvider, error) {
	log.Println("📊 Initializing observability (OpenTelemetry + Grafana Stack)...")

	provider, err := opentelemetry.NewGrafanaProvider(ctx, opentelemetry.GrafanaConfig{
		ServiceName:       cfg.Observability.ServiceName,
		ServiceVersion:    cfg.Observability.ServiceVersion,
		Environment:       cfg.Observability.Environment,
		CollectorEndpoint: cfg.Observability.OTLPEndpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize observability: %w", err)
	}

	log.Printf("✅ Observability initialized (Endpoint: %s)", cfg.Observability.OTLPEndpoint)

	return provider, nil
}
