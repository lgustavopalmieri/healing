package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type ProviderConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
}

type Provider struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	PromRegistry   *prometheus.Registry
	Resource       *sdkresource.Resource
}

func NewProvider(ctx context.Context, cfg ProviderConfig) (*Provider, error) {
	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("deployment.environment.name", cfg.Environment),
		),
		sdkresource.WithHost(),
		sdkresource.WithOS(),
		sdkresource.WithProcess(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otel resource: %w", err)
	}

	tp, err := initTracerProvider(ctx, res, cfg.OTLPEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to init tracer provider: %w", err)
	}

	promRegistry := prometheus.NewRegistry()

	mp, err := initMeterProvider(ctx, res, cfg.OTLPEndpoint, promRegistry)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, fmt.Errorf("failed to init meter provider: %w", err)
	}

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{
		TracerProvider: tp,
		MeterProvider:  mp,
		PromRegistry:   promRegistry,
		Resource:       res,
	}, nil
}

func (p *Provider) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var errs []error

	if err := p.TracerProvider.Shutdown(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("tracer provider shutdown: %w", err))
	}

	if err := p.MeterProvider.Shutdown(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("otel shutdown errors: %v", errs)
	}
	return nil
}

func (p *Provider) MetricsHandler() http.Handler {
	return promhttp.HandlerFor(p.PromRegistry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}
