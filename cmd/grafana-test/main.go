package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/opentelemetry"
)

func main() {
	ctx := context.Background()

	grafanaProvider, err := opentelemetry.NewGrafanaProvider(ctx, opentelemetry.GrafanaConfig{
		ServiceName:       getEnv("SERVICE_NAME", "healing-specialist"),
		ServiceVersion:    getEnv("SERVICE_VERSION", "1.0.0"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		CollectorEndpoint: getEnv("OTEL_COLLECTOR_ENDPOINT", "localhost:4317"),
	})
	if err != nil {
		log.Fatalf("Failed to initialize Grafana provider: %v", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := grafanaProvider.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down provider: %v", err)
		}
	}()

	logger := opentelemetry.NewLogger("healing-specialist")
	tracer := opentelemetry.NewTracer("healing-specialist")
	metrics := opentelemetry.NewMetrics("healing-specialist")

	logger.Info(ctx, "Application started", observability.Field{Key: "version", Value: "1.0.0"})

	counter := metrics.Counter("requests.total")
	histogram := metrics.Histogram("request.duration")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	log.Println("Generating telemetry data... Press Ctrl+C to stop")

	for {
		select {
		case <-ticker.C:
			ctx, span := tracer.Start(ctx, "test.operation")

			logger.Info(ctx, "Processing request",
				observability.Field{Key: "operation", Value: "test"},
				observability.Field{Key: "user_id", Value: "123"},
			)

			counter.Add(ctx, 1, observability.Label{Key: "endpoint", Value: "/test"})
			histogram.Record(ctx, float64(time.Now().UnixMilli()%1000), observability.Label{Key: "method", Value: "GET"})

			time.Sleep(100 * time.Millisecond)

			logger.Info(ctx, "Request completed")
			span.End()

		case <-sigCh:
			log.Println("Shutting down...")
			return
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
