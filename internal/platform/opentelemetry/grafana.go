package opentelemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrafanaConfig struct {
	ServiceName       string
	ServiceVersion    string
	Environment       string
	CollectorEndpoint string // e.g., "localhost:4317"
}

type GrafanaProvider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
	resource       *resource.Resource
}

func NewGrafanaProvider(ctx context.Context, cfg GrafanaConfig) (*GrafanaProvider, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := &GrafanaProvider{resource: res}

	if err := provider.initTraceProvider(ctx, cfg); err != nil {
		return nil, err
	}

	if err := provider.initMetricProvider(ctx, cfg); err != nil {
		return nil, err
	}

	if err := provider.initLogProvider(ctx, cfg); err != nil {
		return nil, err
	}

	return provider, nil
}

func (p *GrafanaProvider) initTraceProvider(ctx context.Context, cfg GrafanaConfig) error {
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.CollectorEndpoint),
		otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	p.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(p.resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(p.tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

func (p *GrafanaProvider) initMetricProvider(ctx context.Context, cfg GrafanaConfig) error {
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.CollectorEndpoint),
		otlpmetricgrpc.WithTLSCredentials(insecure.NewCredentials()),
		otlpmetricgrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	p.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(p.resource),
	)

	otel.SetMeterProvider(p.meterProvider)
	return nil
}

func (p *GrafanaProvider) initLogProvider(ctx context.Context, cfg GrafanaConfig) error {
	logExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(cfg.CollectorEndpoint),
		otlploggrpc.WithTLSCredentials(insecure.NewCredentials()),
		otlploggrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		return fmt.Errorf("failed to create log exporter: %w", err)
	}

	p.loggerProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(p.resource),
	)

	global.SetLoggerProvider(p.loggerProvider)
	return nil
}

func (p *GrafanaProvider) Shutdown(ctx context.Context) error {
	var errs []error

	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if p.loggerProvider != nil {
		if err := p.loggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}
