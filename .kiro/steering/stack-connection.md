---
inclusion: always
---

# Stack Environment Variables - Agent Rules

**Purpose:** Standard environment variables for connecting applications to the Healing Specialist infrastructure stack.

**Context:** These variables are environment-agnostic. Variable names remain constant across all environments (dev, staging, prod). Only values change per environment.

---

## PostgreSQL Database

```env
# Service: PostgreSQL 16
POSTGRES_HOST=<hostname>           # postgres (docker) | localhost (host) | RDS endpoint (prod)
POSTGRES_PORT=5432                 # Always 5432
POSTGRES_USER=<username>           # dev_user (dev) | app_user (prod)
POSTGRES_PASSWORD=<password>       # dev_password (dev) | secrets manager (prod)
POSTGRES_DB=healing_specialist_db  # Database name (constant)
DATABASE_URL=postgresql://<user>:<pass>@<host>:5432/healing_specialist_db
```

---

## Kafka Message Broker

```env
# Service: Apache Kafka (KRaft mode)
KAFKA_BROKER=<hostname>:9092              # broker:9092 (docker) | localhost:9092 (host) | MSK endpoint (prod)
KAFKA_BOOTSTRAP_SERVERS=<hostname>:9092   # Same as KAFKA_BROKER
```

---

## OpenTelemetry Collector

```env
# Service: OTel Collector (traces, metrics, logs)
OTEL_EXPORTER_OTLP_ENDPOINT=http://<hostname>:4318        # HTTP endpoint
OTEL_EXPORTER_OTLP_GRPC_ENDPOINT=http://<hostname>:4317   # gRPC endpoint (alternative)
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf                 # Protocol (constant)
OTEL_SERVICE_NAME=<service-name>                          # Your service identifier
```

**Hostnames by environment:**
- Docker: `otel-collector`
- Host: `localhost`
- Prod: Load balancer or service mesh endpoint

---

## Observability Stack (Read-Only)

```env
# Service: Grafana (dashboards)
GRAFANA_URL=http://<hostname>:3000

# Service: Prometheus (metrics storage)
PROMETHEUS_URL=http://<hostname>:9090

# Service: Tempo (traces storage)
TEMPO_URL=http://<hostname>:3200

# Service: Loki (logs storage)
LOKI_URL=http://<hostname>:3100
```

**Note:** These are typically for internal monitoring tools, not application code.

---

## Environment-Specific Hostname Patterns

### Development (Docker Network)
```
postgres, broker, otel-collector, prometheus, tempo, loki, grafana
```

### Development (Host Machine)
```
localhost
```

### Production
```
RDS endpoints, MSK clusters, ALB/NLB endpoints, service mesh names
```

---

## Agent Instructions

**When generating application code:**

1. **Always use these exact variable names** - they are standardized across environments
2. **Never hardcode values** - always read from environment variables
3. **Provide sensible defaults only for development** - use `os.getenv('VAR', 'default')`
4. **For Docker deployments** - use service names as hostnames (e.g., `postgres`, `broker`)
5. **For host deployments** - use `localhost` as hostname
6. **For production** - expect fully qualified endpoints from DevOps team

**Connection string patterns:**
```go
import (
    "fmt"
    "os"
)

// PostgreSQL
databaseURL := os.Getenv("DATABASE_URL")
// or construct from parts
dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
    os.Getenv("POSTGRES_USER"),
    os.Getenv("POSTGRES_PASSWORD"),
    os.Getenv("POSTGRES_HOST"),
    os.Getenv("POSTGRES_PORT"),
    os.Getenv("POSTGRES_DB"),
)

// Kafka
kafkaBroker := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")

// OpenTelemetry
otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
serviceName := os.Getenv("OTEL_SERVICE_NAME")
if serviceName == "" {
    serviceName = "unknown-service"
}
```

**Docker Compose integration:**
```yaml
services:
  your-app:
    environment:
      - POSTGRES_HOST=postgres
      - POSTGRES_PORT=5432
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=healing_specialist_db
      - KAFKA_BOOTSTRAP_SERVERS=broker:9092
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
      - OTEL_SERVICE_NAME=your-service-name
    networks:
      - healing-network

networks:
  healing-network:
    external: true
```

---

## Quick Reference Table

| Variable | Service | Port | Docker Host | Local Host | Notes |
|----------|---------|------|-------------|------------|-------|
| `POSTGRES_HOST` | PostgreSQL | 5432 | `postgres` | `localhost` | Database server |
| `KAFKA_BOOTSTRAP_SERVERS` | Kafka | 9092 | `broker:9092` | `localhost:9092` | Message broker |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTel Collector | 4318 | `http://otel-collector:4318` | `http://localhost:4318` | Telemetry (HTTP) |
| `OTEL_EXPORTER_OTLP_GRPC_ENDPOINT` | OTel Collector | 4317 | `http://otel-collector:4317` | `http://localhost:4317` | Telemetry (gRPC) |
| `GRAFANA_URL` | Grafana | 3000 | `http://grafana:3000` | `http://localhost:3000` | Dashboards |
| `PROMETHEUS_URL` | Prometheus | 9090 | `http://prometheus:9090` | `http://localhost:9090` | Metrics |
| `TEMPO_URL` | Tempo | 3200 | `http://tempo:3200` | `http://localhost:3200` | Traces |
| `LOKI_URL` | Loki | 3100 | `http://loki:3100` | `http://localhost:3100` | Logs |

---

## Validation Checklist

When implementing service connections, ensure:

- ✅ All connection strings use environment variables
- ✅ No hardcoded credentials or endpoints
- ✅ Graceful fallback for missing optional variables
- ✅ Clear error messages for missing required variables
- ✅ Connection retry logic with exponential backoff
- ✅ Health check endpoints for each external dependency
