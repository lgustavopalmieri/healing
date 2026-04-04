package telemetry

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

type HTTPMetrics struct {
	RequestsTotal    otelmetric.Float64Counter
	ResponseDuration otelmetric.Float64Histogram
	AppName          string
}

func NewHTTPMetrics(serviceName string) *HTTPMetrics {
	meter := otel.Meter(serviceName)

	requestsTotal, _ := meter.Float64Counter(
		"app_requests_total",
		otelmetric.WithDescription("Total number of requests processed"),
	)

	responseDuration, _ := meter.Float64Histogram(
		"app_response_time_seconds",
		otelmetric.WithDescription("Response time per request in seconds"),
		otelmetric.WithUnit("s"),
	)

	return &HTTPMetrics{
		RequestsTotal:    requestsTotal,
		ResponseDuration: responseDuration,
		AppName:          serviceName,
	}
}

func (m *HTTPMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		attrs := otelmetric.WithAttributes(
			attribute.String("app", m.AppName),
			attribute.String("endpoint", endpoint),
		)

		m.RequestsTotal.Add(c.Request.Context(), 1, attrs)
		m.ResponseDuration.Record(c.Request.Context(), duration, attrs)
	}
}
