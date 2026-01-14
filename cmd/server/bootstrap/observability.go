package bootstrap

import (
	"context"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/opentelemetry"
)

func InitObservability(ctx context.Context, cfg *config.Config) (*opentelemetry.GrafanaProvider, error) {
	provider, err := opentelemetry.NewGrafanaProvider(ctx, opentelemetry.GrafanaConfig{
		ServiceName:       cfg.Observability.ServiceName,
		ServiceVersion:    cfg.Observability.ServiceVersion,
		Environment:       cfg.Observability.Environment,
		CollectorEndpoint: cfg.Observability.OTLPEndpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize observability: %w", err)
	}

	return provider, nil
}
