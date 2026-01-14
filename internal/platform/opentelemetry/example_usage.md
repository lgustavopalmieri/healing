# Example usage of the Datadog OpenTelemetry integration

This file demonstrates how to initialize and use the observability stack
with Datadog as the backend.
Usage in main.go or application bootstrap:

```go
import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/opentelemetry"
)
func main() {
	ctx := context.Background()
	Initialize Datadog provider
	ddProvider, err := opentelemetry.NewDatadogProvider(ctx, opentelemetry.DatadogConfig{
		ServiceName:    "healing-specialist",
		ServiceVersion: "1.0.0",
		Environment:    os.Getenv("ENV"), // e.g., "production", "staging", "development"
		DatadogSite:    "datadoghq.com",  // or "datadoghq.eu", "us3.datadoghq.com", etc.
		APIKey:         os.Getenv("DD_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to initialize Datadog provider: %v", err)
	}
	Ensure graceful shutdown
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := ddProvider.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down Datadog provider: %v", err)
		}
	}()
	Create observability instances
	logger := opentelemetry.NewLogger("healing-specialist")
	tracer := opentelemetry.NewTracer("healing-specialist")
	metrics := opentelemetry.NewMetrics("healing-specialist")
	Use in your application
	logger.Info(ctx, "Application started", observability.Field{Key: "version", Value: "1.0.0"})
	ctx, span := tracer.Start(ctx, "main.operation")
	defer span.End()
	counter := metrics.Counter("requests.total")
	counter.Add(ctx, 1, observability.Label{Key: "endpoint", Value: "/health"})
	Your application logic here...
	Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}
```

Environment Variables:
  - DD_API_KEY: Your Datadog API key (required)
  - ENV: Environment name (e.g., production, staging, development)
Datadog Sites:
  - US1: datadoghq.com (default)
  - US3: us3.datadoghq.com
  - US5: us5.datadoghq.com
  - EU: datadoghq.eu
  - AP1: ap1.datadoghq.com
  - US1-FED: ddog-gov.com
