package bootstrap

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/telemetry"
)

func InitOtel(ctx context.Context, cfg *config.Config) (*telemetry.Provider, error) {
	if cfg.Otel.ExporterEndpoint == "" {
		log.Println("OTEL_EXPORTER_OTLP_ENDPOINT not set, skipping OTel initialization")
		return nil, nil
	}

	endpoint := cfg.Otel.ExporterEndpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	log.Printf("Initializing OTel metrics (endpoint=%s, service=%s)...", endpoint, cfg.Otel.ServiceName)

	provider, err := telemetry.NewProvider(ctx, telemetry.ProviderConfig{
		ServiceName:        cfg.Otel.ServiceName,
		ServiceVersion:     "1.0.0",
		OTLPEndpoint:       endpoint,
		ResourceAttributes: cfg.Otel.ResourceAttributes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OTel: %w", err)
	}

	log.Println("OTel metrics initialized successfully")
	return provider, nil
}
