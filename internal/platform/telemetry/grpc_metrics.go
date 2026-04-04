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
	AppName         string
}

func NewGRPCMetrics(serviceName string) *GRPCMetrics {
	meter := otel.Meter(serviceName)

	requestsTotal, _ := meter.Float64Counter(
		"app_requests_total",
		otelmetric.WithDescription("Total number of requests processed"),
	)

	requestDuration, _ := meter.Float64Histogram(
		"app_response_time_seconds",
		otelmetric.WithDescription("Response time per request in seconds"),
		otelmetric.WithUnit("s"),
	)

	return &GRPCMetrics{
		RequestsTotal:   requestsTotal,
		RequestDuration: requestDuration,
		AppName:         serviceName,
	}
}

func (m *GRPCMetrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()

		grpcCode := "OK"
		if err != nil {
			st, _ := status.FromError(err)
			grpcCode = st.Code().String()
		}

		endpoint := info.FullMethod

		m.RequestsTotal.Add(ctx, 1, otelmetric.WithAttributes(
			attribute.String("app", m.AppName),
			attribute.String("endpoint", endpoint),
			attribute.String("grpc_code", grpcCode),
		))

		m.RequestDuration.Record(ctx, duration, otelmetric.WithAttributes(
			attribute.String("app", m.AppName),
			attribute.String("endpoint", endpoint),
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
