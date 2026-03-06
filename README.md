# Healing

Healing is a digital healthcare platform that connects patients with health specialists across all disciplines — human and veterinary medicine, traditional and non-traditional practices (Chinese medicine, therapies, holistic treatments, etc.). Patients own their complete medical history built from every consultation on the platform. Specialists can create detailed profiles, collaborate on cases, and leverage AI agents to extend their reach.

## Current Scope

This repository contains the **Specialist Service** — the first microservice of the platform, responsible for specialist onboarding, credential validation, profile management, and discovery.

What's implemented so far:
- Specialist registration with external license validation
- Full-text search over specialist profiles with filters, sorting, and cursor-based pagination
- gRPC API for both features
- Event publishing for specialist lifecycle events
- Observability stack (tracing, metrics, structured logging)
- Stress testing infrastructure

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25 |
| API | gRPC + Protocol Buffers |
| Database | PostgreSQL (persistence) |
| Search Engine | Elasticsearch (specialist search index) |
| Event Streaming | Apache Kafka (event publishing/consuming) |
| Tracing | OpenTelemetry (OTLP export via HTTP and gRPC) |
| Metrics | Prometheus (custom metrics + gRPC interceptor metrics) |
| Logging | Go slog (structured JSON, trace-correlated) |
| Testing | testify + gomock + testcontainers-go |
| Stress Testing | Grafana k6 (gRPC load tests via Docker) |
| Migrations | Goose (SQL migrations) |
| Configuration | Viper (env-based config) |
| Containerization | Docker + Docker Compose |

## Running

```bash
# Start infrastructure
docker compose up -d

# Run the gRPC server
make run

# Run tests
go test ./...
```

## Swagger

http://localhost:8080/swagger/index.html (local)

http://localhost:4000/swagger/index.html (Docker)
