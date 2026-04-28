# Healing

Healing is a digital healthcare platform that connects patients with health specialists across all disciplines — human and veterinary medicine, traditional and non-traditional practices (Chinese medicine, therapies, holistic treatments, etc.). Patients own their complete medical history built from every consultation on the platform. Specialists can create detailed profiles, collaborate on cases, and leverage AI agents to extend their reach.

## Current Scope

This repository contains the **Specialist Service** — the first microservice of the platform, responsible for specialist onboarding, credential validation, profile management, and discovery.

What's implemented so far:

- Specialist registration with external license validation
- Specialist profile update with event-driven data synchronization
- Full-text search over specialist profiles with filters, sorting, and cursor-based pagination
- Dual transport API: gRPC + REST (Gin) with Swagger documentation
- Event-driven architecture via AWS SQS FIFO queues
- Search indexing via AWS OpenSearch with multi-tenant isolation support
- Full observability stack: OpenTelemetry tracing (gRPC export), OTel metrics (HTTP export), structured logging (slog)
- Integration tests with testcontainers (PostgreSQL, OpenSearch, LocalStack SQS)
- Stress testing infrastructure (Grafana k6 for gRPC and HTTP)

## Tech Stack

- **Language** — Go 1.25.0
- **API** — gRPC + Protocol Buffers, REST via Gin, Swagger (swaggo)
- **Database** — PostgreSQL (write store, lib/pq driver, Goose migrations)
- **Search Engine** — AWS OpenSearch (opensearch-go/v4, AWS SigV4 auth, FGAC multi-tenant)
- **Message Queue** — AWS SQS FIFO (aws-sdk-go-v2, idempotent queue creation, native DLQ)
- **Tracing** — OpenTelemetry SDK v1.42, OTLP export via gRPC
- **Metrics** — OpenTelemetry SDK v1.42, OTLP export via HTTP (gRPC + HTTP interceptor metrics)
- **Logging** — Go slog (structured JSON, trace-correlated)
- **Testing** — testify + gomock + testcontainers-go (PostgreSQL, OpenSearch, LocalStack)
- **Stress Testing** — Grafana k6 (gRPC and HTTP load tests via Docker)
- **Configuration** — Viper (env-based config)
- **Containerization** — Docker (multi-stage, scratch, non-root) + Docker Compose

## Architecture

The service follows Clean Architecture / Hexagonal with DDD principles:

- **Domain** — Entities, value objects, business rules, and domain errors
- **Application** — Use cases orchestrating domain logic, repository interfaces, and event publishing
- **Adapters/Inbound** — gRPC services and HTTP handlers (driving adapters)
- **Adapters/Outbound** — PostgreSQL repositories, OpenSearch repositories, external gateways (driven adapters)
- **Event Listeners** — SQS consumers as independent mini-modules with their own listener + adapters layers

Event flow: write operations publish events to SQS FIFO queues → consumers handle async side effects (license validation, search index sync, future email notifications).

## Running

```bash
# Run locally
make run

# Run with Docker
make up

# Stop
make down

# Run tests
go test ./...

# Run tests (same as CI)
go test -race -coverprofile=coverage.out -covermode=atomic -timeout 600s ./...
```

## Swagger

- Local: http://localhost:8080/swagger/index.html
- Docker: http://localhost:4000/swagger/index.html

## Ports

- **gRPC** — 50051
- **HTTP** — 8080 (mapped to 4000 in Docker)
