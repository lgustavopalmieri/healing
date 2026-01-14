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
	"google.golang.org/grpc/credentials"
)

// DatadogConfig holds configuration for Datadog integration
type DatadogConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	DatadogSite    string // e.g., "datadoghq.com", "datadoghq.eu", "us3.datadoghq.com"
	APIKey         string
}

// DatadogProvider manages OpenTelemetry providers configured for Datadog
type DatadogProvider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
	resource       *resource.Resource
}

// NewDatadogProvider initializes OpenTelemetry with Datadog exporters
func NewDatadogProvider(ctx context.Context, cfg DatadogConfig) (*DatadogProvider, error) {
	// Create resource with service information
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

	provider := &DatadogProvider{
		resource: res,
	}

	// Initialize trace provider
	if err := provider.initTraceProvider(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize trace provider: %w", err)
	}

	// Initialize metric provider
	if err := provider.initMetricProvider(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize metric provider: %w", err)
	}

	// Initialize log provider
	if err := provider.initLogProvider(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize log provider: %w", err)
	}

	return provider, nil
}

func (p *DatadogProvider) initTraceProvider(ctx context.Context, cfg DatadogConfig) error {
	// Datadog OTLP gRPC endpoint
	endpoint := fmt.Sprintf("api.%s:443", cfg.DatadogSite)

	// Create OTLP gRPC trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		otlptracegrpc.WithHeaders(map[string]string{
			"dd-api-key": cfg.APIKey,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	p.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(p.resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global trace provider
	otel.SetTracerProvider(p.tracerProvider)

	// Set global propagator for distributed tracing
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

func (p *DatadogProvider) initMetricProvider(ctx context.Context, cfg DatadogConfig) error {
	// Datadog OTLP gRPC endpoint
	endpoint := fmt.Sprintf("api.%s:443", cfg.DatadogSite)

	// Create OTLP gRPC metric exporter
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		otlpmetricgrpc.WithHeaders(map[string]string{
			"dd-api-key": cfg.APIKey,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Create metric provider
	p.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(p.resource),
	)

	// Set global meter provider
	otel.SetMeterProvider(p.meterProvider)

	return nil
}

func (p *DatadogProvider) initLogProvider(ctx context.Context, cfg DatadogConfig) error {
	// Datadog OTLP gRPC endpoint
	endpoint := fmt.Sprintf("api.%s:443", cfg.DatadogSite)

	// Create OTLP gRPC log exporter
	logExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		otlploggrpc.WithHeaders(map[string]string{
			"dd-api-key": cfg.APIKey,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create log exporter: %w", err)
	}

	// Create log provider
	p.loggerProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(p.resource),
	)

	// Set global logger provider
	global.SetLoggerProvider(p.loggerProvider)

	return nil
}

// Shutdown gracefully shuts down all providers
func (p *DatadogProvider) Shutdown(ctx context.Context) error {
	var errs []error

	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("trace provider shutdown: %w", err))
		}
	}

	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
		}
	}

	if p.loggerProvider != nil {
		if err := p.loggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("logger provider shutdown: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}
