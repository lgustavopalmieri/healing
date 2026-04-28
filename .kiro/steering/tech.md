---
inclusion: always
---

# Technology Stack & Build System

## Language

Go 1.25.0

## Dependencies (direct requires from go.mod)

```yaml
gRPC & Protobuf:
  google.golang.org/grpc: v1.79.2
  google.golang.org/protobuf: v1.36.11
  github.com/golang/protobuf: v1.5.4       # legacy, used by generated code

HTTP:
  github.com/gin-gonic/gin: v1.12.0
  github.com/gin-contrib/cors: v1.7.6

API Documentation:
  github.com/swaggo/swag: v1.16.6
  github.com/swaggo/gin-swagger: v1.6.1
  github.com/swaggo/files: v1.0.1

Database:
  github.com/lib/pq: v1.10.9               # PostgreSQL driver (database/sql)
  github.com/pressly/goose/v3: v3.26.0     # SQL migrations

Search (OpenSearch):
  github.com/opensearch-project/opensearch-go/v4: v4.6.0

Message Queue (AWS SQS):
  github.com/aws/aws-sdk-go-v2: v1.41.5
  github.com/aws/aws-sdk-go-v2/config: v1.32.13
  github.com/aws/aws-sdk-go-v2/credentials: v1.19.13
  github.com/aws/aws-sdk-go-v2/service/sqs: v1.42.25

Observability:
  go.opentelemetry.io/otel: v1.42.0
  go.opentelemetry.io/otel/trace: v1.42.0
  go.opentelemetry.io/otel/metric: v1.42.0
  go.opentelemetry.io/otel/sdk: v1.42.0
  go.opentelemetry.io/otel/sdk/metric: v1.42.0
  go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.42.0
  go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp: v1.42.0

Config:
  github.com/spf13/viper: v1.21.0          # .env, yaml, env vars

Utils:
  github.com/google/uuid: v1.6.0

Testing:
  github.com/stretchr/testify: v1.11.1     # assert/require
  go.uber.org/mock: v0.6.0                 # gomock
  github.com/testcontainers/testcontainers-go: v0.41.0
  github.com/testcontainers/testcontainers-go/modules/postgres: v0.40.0
  github.com/testcontainers/testcontainers-go/modules/opensearch: v0.41.0
  github.com/testcontainers/testcontainers-go/modules/localstack: v0.41.0
```

## Infrastructure Services

```yaml
PostgreSQL: primary database (write store)
AWS OpenSearch: search/read store (specialist search, multi-tenant via FGAC)
AWS SQS (FIFO): event bus (event-driven communication between features, native DLQ)
OpenTelemetry (OTLP): distributed tracing (export via gRPC) and metrics (export via HTTP)
```

## Observability Stack

The service uses OpenTelemetry as the unified observability framework:

- **Tracing** — OTel SDK with OTLP gRPC exporter, ParentBased sampler (100% ratio)
- **Metrics** — OTel SDK with OTLP HTTP exporter, periodic reader. Custom metrics via observability.Metrics interface (Counter, Histogram, Gauge)
- **gRPC Metrics** — Interceptor-based: request count, latency histogram, per-method attributes
- **HTTP Metrics** — Gin middleware: request count, response time histogram, per-endpoint attributes
- **Logging** — Go slog with structured JSON output, trace-correlated

Observability interfaces (`internal/commom/observability/`):
- `Logger` — Debug, Info, Warn, Error with context and structured fields
- `Tracer` — Start span with context propagation, RecordError, SetAttribute
- `Metrics` — Counter, Histogram, Gauge with labels

## Ports

```yaml
gRPC: 50051
HTTP: 8080 (mapped to 4000 in docker-compose)
```

## Build & Run

```bash
# Run locally (without Docker)
make run
# equivalent to: APP_ENV=development go run ./cmd

# Run with Docker
make up

# Stop
make down
```

## Docker

- Multi-stage build: golang:1.25.0-alpine -> scratch
- Binary: CGO_ENABLED=0, static, trimpath, -ldflags="-s -w"
- Runs as non-root (user 65534:65534)
- Entrypoint: /healing-specialist (built from ./cmd)
- Exposes ports: 50051 (gRPC), 8080 (HTTP)

## CI (GitHub Actions)

```yaml
trigger: push/PR on develop
go version: 1.25.0
concurrency: cancel-in-progress per branch
steps:
  - go mod download + verify
  - go vet ./...
  - go test -race -coverprofile=coverage.out -covermode=atomic -timeout 600s ./...
env: TESTCONTAINERS_RYUK_DISABLED=false
```

## Code Generation

```yaml
Mocks:
  tool: go.uber.org/mock/gomock
  pattern: //go:generate directives in interface files
  output: mocks/ inside each package

Protocol Buffers:
  tools: protoc-gen-go, protoc-gen-go-grpc
  source: */proto/*.proto
  output: */pb/*.pb.go, */pb/*_grpc.pb.go
  command: protoc --go_out=. --go-grpc_out=. proto/<service>.proto
```

## Testing

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Same command as CI
go test -race -coverprofile=coverage.out -covermode=atomic -timeout 600s ./...
```

Integration tests use testcontainers (PostgreSQL, OpenSearch, LocalStack SQS) and run together with `go test ./...`

## gRPC Testing (Evans)

```bash
go install github.com/ktr0731/evans@latest
evans -r repl
# show package -> package pb -> show service -> service <ServiceName> -> call <Method>
```

## Load Testing (k6)

```bash
# Inside tests/k6/
make <target>
# Uses docker-compose.k6.yml to spin up the environment
# gRPC (stress/) and HTTP (stress-http/) tests are separate
```

## Environment Variables

```yaml
APP_ENV: application environment (development, staging, production)
ENV_DIR: directory for .env file

SERVER_GRPC_PORT: gRPC server port (default 50051)
SERVER_HTTP_PORT: HTTP server port (default 8080)
SERVER_SHUTDOWN_TIMEOUT: graceful shutdown timeout
SERVER_MAX_CONNECTIONS: max gRPC connections
SERVER_CONNECTION_TIMEOUT: connection timeout

POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, POSTGRES_SSLMODE: database connection
POSTGRES_MAX_OPEN_CONNS, POSTGRES_MAX_IDLE_CONNS, POSTGRES_CONN_MAX_LIFETIME, POSTGRES_CONN_MAX_IDLE_TIME: pool tuning

SQS_REGION: AWS region for SQS
SQS_QUEUE_PREFIX: queue name prefix (e.g. "specialist")
SQS_ENDPOINT: optional, for LocalStack in development

OPENSEARCH_ADDRESSES: comma-separated OpenSearch endpoints
OPENSEARCH_REGION: AWS region (enables SigV4 auth when set)
OPENSEARCH_INDEX_PREFIX: index name prefix (e.g. "healing")

LICENSE_VALIDATION_BASE_URL: external license validation API

OTEL_EXPORTER_OTLP_ENDPOINT: OTLP collector endpoint
OTEL_EXPORTER_OTLP_PROTOCOL: export protocol (http/protobuf)
OTEL_SERVICE_NAME: service name for telemetry
OTEL_RESOURCE_ATTRIBUTES: additional resource attributes (key=value,key=value)
```
