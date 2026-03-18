package telemetry

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type GRPCMetrics struct {
	RequestsTotal   otelmetric.Float64Counter
	RequestDuration otelmetric.Float64Histogram
	HealthStatus    otelmetric.Float64Gauge
}

func NewGRPCMetrics(serviceName string) *GRPCMetrics {
	meter := otel.Meter(serviceName)

	requestsTotal, _ := meter.Float64Counter("grpc_requests_total")
	requestDuration, _ := meter.Float64Histogram("grpc_request_duration_seconds")
	healthStatus, _ := meter.Float64Gauge("application_health_status")

	m := &GRPCMetrics{
		RequestsTotal:   requestsTotal,
		RequestDuration: requestDuration,
		HealthStatus:    healthStatus,
	}

	m.SetHealthy()

	return m
}

func (m *GRPCMetrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		service, method := splitMethodName(info.FullMethod)
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()

		grpcCode := "OK"
		if err != nil {
			st, _ := status.FromError(err)
			grpcCode = st.Code().String()
		}

		attrs := otelmetric.WithAttributes(
			attribute.String("grpc_service", service),
			attribute.String("grpc_method", method),
			attribute.String("grpc_code", grpcCode),
		)

		m.RequestsTotal.Add(ctx, 1, attrs)
		m.RequestDuration.Record(ctx, duration, otelmetric.WithAttributes(
			attribute.String("grpc_service", service),
			attribute.String("grpc_method", method),
		))

		return resp, err
	}
}

func splitMethodName(fullMethod string) (string, string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/")
	parts := strings.Split(fullMethod, "/")
	if len(parts) != 2 {
		return "unknown", "unknown"
	}
	serviceParts := strings.Split(parts[0], ".")
	return serviceParts[len(serviceParts)-1], parts[1]
}

func (m *GRPCMetrics) SetHealthy() {
	m.HealthStatus.Record(context.Background(), 1)
}

func (m *GRPCMetrics) SetUnhealthy() {
	m.HealthStatus.Record(context.Background(), 0)
}
