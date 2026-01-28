package telemetry

import (
	"context"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// GRPCMetrics holds Prometheus metrics for gRPC operations
type GRPCMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	healthStatus    prometheus.Gauge
}

// NewGRPCMetrics creates and registers gRPC metrics with the provided registry
func NewGRPCMetrics(registry *prometheus.Registry) *GRPCMetrics {
	metrics := &GRPCMetrics{
		requestsTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_requests_total",
				Help: "Total number of gRPC requests by service, method and status",
			},
			[]string{"grpc_service", "grpc_method", "grpc_code"},
		),
		requestDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "grpc_request_duration_seconds",
				Help:    "Duration of gRPC requests in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			},
			[]string{"grpc_service", "grpc_method"},
		),
		healthStatus: promauto.With(registry).NewGauge(
			prometheus.GaugeOpts{
				Name: "application_health_status",
				Help: "Application health status (1 = healthy, 0 = unhealthy)",
			},
		),
	}

	// Set initial health status as healthy
	metrics.healthStatus.Set(1)

	return metrics
}

// UnaryServerInterceptor returns a gRPC unary server interceptor that records metrics
func (m *GRPCMetrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Extract service and method names from full method path
		// Example: /pb.SpecialistService/CreateSpecialist -> service=SpecialistService, method=CreateSpecialist
		service, method := splitMethodName(info.FullMethod)

		timer := prometheus.NewTimer(m.requestDuration.WithLabelValues(service, method))
		defer timer.ObserveDuration()

		resp, err := handler(ctx, req)

		grpcCode := "OK"
		if err != nil {
			st, _ := status.FromError(err)
			grpcCode = st.Code().String()
		}

		m.requestsTotal.WithLabelValues(service, method, grpcCode).Inc()

		return resp, err
	}
}

// splitMethodName extracts service and method from gRPC full method name
// Example: /pb.SpecialistService/CreateSpecialist -> (SpecialistService, CreateSpecialist)
func splitMethodName(fullMethod string) (string, string) {
	// Remove leading slash
	fullMethod = strings.TrimPrefix(fullMethod, "/")

	// Split by slash to separate service and method
	parts := strings.Split(fullMethod, "/")
	if len(parts) != 2 {
		return "unknown", "unknown"
	}

	// Extract service name (remove package prefix if present)
	serviceParts := strings.Split(parts[0], ".")
	service := serviceParts[len(serviceParts)-1]

	method := parts[1]

	return service, method
}

// SetHealthy marks the application as healthy
func (m *GRPCMetrics) SetHealthy() {
	m.healthStatus.Set(1)
}

// SetUnhealthy marks the application as unhealthy
func (m *GRPCMetrics) SetUnhealthy() {
	m.healthStatus.Set(0)
}
