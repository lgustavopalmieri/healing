package telemetry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type ProviderConfig struct {
	ServiceName        string
	ServiceVersion     string
	Environment        string
	OTLPEndpoint       string
	ResourceAttributes string
}

type Provider struct {
	MeterProvider *sdkmetric.MeterProvider
	Resource      *sdkresource.Resource
}

func NewProvider(ctx context.Context, cfg ProviderConfig) (*Provider, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
	}

	if cfg.Environment != "" {
		attrs = append(attrs, attribute.String("deployment.environment", cfg.Environment))
	}

	for _, pair := range strings.Split(cfg.ResourceAttributes, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			attrs = append(attrs, attribute.String(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])))
		}
	}

	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(attrs...),
		sdkresource.WithHost(),
		sdkresource.WithOS(),
		sdkresource.WithProcess(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otel resource: %w", err)
	}

	mp, err := initMeterProvider(ctx, res, cfg.OTLPEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to init meter provider: %w", err)
	}

	otel.SetMeterProvider(mp)

	return &Provider{
		MeterProvider: mp,
		Resource:      res,
	}, nil
}

func (p *Provider) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := p.MeterProvider.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("meter provider shutdown: %w", err)
	}
	return nil
}
