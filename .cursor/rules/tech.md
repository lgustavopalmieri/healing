---
inclusion: always
---

# Technology Stack & Build System

## Language

Go 1.25.0

## Dependencies (direct requires from go.mod)

```yaml
gRPC & Protobuf:
  google.golang.org/grpc: v1.78.0
  google.golang.org/protobuf: v1.36.11
  github.com/golang/protobuf: v1.5.4       # legacy, used by generated code

HTTP:
  github.com/gin-gonic/gin: v1.12.0

Database:
  github.com/lib/pq: v1.10.9               # PostgreSQL driver (database/sql)
  github.com/pressly/goose/v3: v3.26.0     # SQL migrations

Elasticsearch:
  github.com/elastic/go-elasticsearch/v8: v8.19.1

Kafka:
  github.com/twmb/franz-go: v1.20.7
  github.com/twmb/franz-go/pkg/kadm: v1.17.2  # admin operations

Observability:
  go.opentelemetry.io/otel: v1.39.0
  go.opentelemetry.io/otel/trace: v1.39.0
  go.opentelemetry.io/otel/sdk: v1.39.0
  go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp: v1.39.0
  github.com/prometheus/client_golang: v1.17.0

Config:
  github.com/spf13/viper: v1.21.0          # .env, yaml, env vars

Utils:
  github.com/google/uuid: v1.6.0

Testing:
  github.com/stretchr/testify: v1.11.1     # assert/require
  go.uber.org/mock: v0.6.0                 # gomock
  github.com/testcontainers/testcontainers-go: v0.40.0
  github.com/testcontainers/testcontainers-go/modules/postgres: v0.40.0
  github.com/testcontainers/testcontainers-go/modules/elasticsearch: v0.40.0
  github.com/testcontainers/testcontainers-go/modules/kafka: v0.40.0
```

## Infrastructure Services

```yaml
PostgreSQL: primary database (write store)
Elasticsearch: search/read store (specialist search)
Kafka: event bus (event-driven communication between features)
Prometheus: metrics (exposed via metrics_server on port 9090)
OpenTelemetry (OTLP): distributed tracing (export via HTTP)
```

## Ports

```yaml
gRPC: 50051
HTTP: 8080 (mapped to 4000 in docker-compose)
Metrics: 9090 (mapped to 4001 in docker-compose)
```

## Build & Run

```bash
# Run locally (without Docker)
make run
# equivalent to: APP_ENV=development cd cmd/server && go run main.go

# Run with Docker
make up

# Stop
make down
```

## Docker

- Multi-stage build: golang:1.25.0-alpine -> scratch
- Binary: CGO_ENABLED=0, static, trimpath
- Runs as non-root (user 65534)
- Entrypoint: /healing-specialist (built from ./cmd)

## CI (GitHub Actions)

```yaml
trigger: push/PR on main and develop
go version: 1.25.0
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

Integration tests use testcontainers (PostgreSQL, Elasticsearch, Kafka) and run together with go test ./...

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