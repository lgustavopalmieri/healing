---
inclusion: always
---

# Project Structure & Organization

## Root Layout

```
.
├── cmd/
│   ├── main.go                        # Application entrypoint
│   └── server/
│       ├── bootstrap/                 # Initialization orchestration
│       │   ├── database.go
│       │   ├── opensearch.go
│       │   ├── sqs.go
│       │   ├── sqs_consumers.go
│       │   ├── grpc_services.go
│       │   ├── http_services.go
│       │   ├── otel.go
│       │   └── shutdown.go
│       └── config/
│           ├── config.go
│           ├── helpers.go
│           ├── load.go
│           └── validate.go
├── internal/
│   ├── commom/                        # Shared cross-module utilities (note: "commom" typo is intentional)
│   ├── modules/                       # Business domain modules
│   └── platform/                      # Infrastructure adapters
├── tests/
│   └── k6/                            # Load/stress tests (k6 framework)
├── adr/                               # Architecture Decision Records
├── docs/                              # Swagger generated docs
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod / go.sum
└── .env / .env.example
```

## internal/commom/ — Shared Utilities

```
commom/
├── event/
│   ├── dipstacher.go                  # EventDispatcher interface
│   ├── event.go                       # Event struct (Name, Payload, Timestamp)
│   ├── listener.go                    # Listener interface
│   └── retry.go                       # Retry with exponential backoff
├── observability/
│   ├── logging.go                     # Logger interface (Debug, Info, Warn, Error)
│   ├── metrics.go                     # Metrics interface (Counter, Histogram, Gauge)
│   ├── tracing.go                     # Tracer interface (Start, Span, RecordError)
│   └── mocks/
│       └── logger_mock.go             # Shared logger mock (gomock)
├── tests/                             # Test infrastructure helpers
│   ├── database/postgresql/           # Testcontainers PostgreSQL setup
│   ├── opensearch/                    # Testcontainers OpenSearch setup + factory
│   └── event/sqs/                     # Testcontainers LocalStack SQS setup
├── utils/
│   ├── sanitize.go
│   └── sanitize_test.go
└── value-objects/
    └── pagination/cursor/             # Cursor-based pagination value object
        ├── input.go / output.go       # Input/Output structs
        ├── encode.go / decode.go      # Cursor encoding/decoding
        ├── validate_input.go          # Input validation
        ├── errors.go / utils.go
        ├── encode_test.go / decode_test.go
        └── README.md
```

## internal/modules/specialist/ — Domain Module

### Domain Layer

```
domain/
├── entity.go                          # Specialist entity struct (package: domain)
├── errors.go                          # Shared domain errors
├── status.go                          # Status enum/constants (pending, active, etc.)
├── validate.go                        # Shared validation logic
├── validate_test.go
├── create/                            # Create domain logic (package: create)
│   ├── create.go                      # Factory function + CreateSpecialistInput
│   ├── errors.go                      # Create-specific errors
│   ├── uniqueness_errors.go           # Uniqueness constraint errors (email, license)
│   └── create_test.go
├── authorize_license/                 # License authorization domain logic
│   ├── authorize_license.go
│   ├── errors.go
│   └── authorize_license_test.go
├── update/                            # Update domain logic
│   ├── update.go
│   ├── validators.go
│   └── errors.go
└── search/                            # Search domain logic
    ├── errors.go
    ├── search_input/                  # Search input value object (sub-package)
    │   ├── input.go
    │   ├── field.go                   # Searchable field definitions
    │   ├── sort.go                    # Sort options
    │   ├── validate_input.go
    │   └── validate_input_test.go
    └── search_output/                 # Search output value object (sub-package)
        ├── output.go
        └── validate_output.go
```

Domain layer specifics:
- `domain/` (package `domain`) contains entity, status, errors, and shared validations
- Each sub-domain (`create/`, `update/`, `authorize_license/`) is a separate package that imports `domain`
- `search/` has extra depth: sub-packages `search_input/` and `search_output/` as value objects
- `create/` has `uniqueness_errors.go` separate from generic `errors.go` — pattern for constraint errors

### Features Layer

Each feature follows the `application/ + adapters/` structure (Clean Architecture / Hexagonal), with variations per feature.

The `adapters/` directory is split into:
- `adapters/inbound/` — driving adapters (HTTP handlers, gRPC services)
- `adapters/outbound/` — driven adapters (database repositories, OpenSearch repositories)

Note: `event_listeners/` inside features also use the `adapters/inbound/outbound` structure, matching the feature-level convention.

#### Feature: create

```
features/create/
├── application/
│   ├── usecase.go / usecase_test.go
│   ├── new_usecase.go
│   ├── interface.go
│   ├── dto.go
│   ├── constants.go
│   └── mocks/                         # 4 mocks: event_dispatcher, logger, repository, tracer
├── event_listeners/                   # Event-driven side effects
│   ├── send_credentials_email/        # (empty — placeholder)
│   └── validate_license/              # SQS consumer listener (fully implemented)
│       ├── listener/                  # Handler logic (mirrors application/ pattern)
│       │   ├── handler.go / handler_test.go
│       │   ├── new_handler.go
│       │   ├── interface.go
│       │   ├── dto.go
│       │   ├── constants.go
│       │   └── mocks/                 # 2 mocks: event_dispatcher, repository
│       └── adapters/
│           ├── inbound/
│           │   └── sqs/               # SQS consumer management
│           └── outbound/
│               ├── database/          # repository.go, new.go, errors.go, repository_test.go
│               └── external/          # gateway.go, new.go
└── adapters/
    ├── inbound/
    │   ├── http_handler/              # handler.go, handler_test.go, dto.go, di.go, mocks/
    │   ├── grpc_service/              # service.go, service_test.go, dto.go, di.go, mocks/, pb/, proto/
    │   └── handler_integration_test.go  # Integration test (at inbound/ root, not inside a sub-folder)
    └── outbound/
        └── database/                  # repository.go, new.go, errors.go, repository_test.go
```

Create specifics:
- Only feature with `event_listeners/` containing fully implemented SQS listeners
- `validate_license/` is a mini-module with its own `listener/` + `adapters/` (inbound/sqs, outbound/database, outbound/external)
- `send_credentials_email/` exists as an empty placeholder
- `handler_integration_test.go` lives at the `adapters/inbound/` root, not inside `database/` or `http_handler/`
- `adapters/inbound/` has both `http_handler/` and `grpc_service/` (dual transport)

#### Feature: search

```
features/search/
├── application/
│   ├── usecase.go / usecase_test.go
│   ├── new_usecase.go
│   ├── interface.go
│   ├── dto.go
│   ├── constants.go
│   └── mocks/                         # 1 mock: repository_mock (no logger/tracer mocks)
└── adapters/
    ├── inbound/
    │   ├── http_handler/              # handler.go, handler_test.go, dto.go, di.go, mocks/
    │   └── grpc_service/              # service.go, service_test.go, dto.go, di.go, mocks/, pb/, proto/
    └── outbound/
        └── opensearch/                # OpenSearch repository (read store)
            ├── repository.go / repository_test.go
            ├── new.go / errors.go
            ├── builders.go            # Query builders
            ├── mappers.go             # Response mappers
            └── dto.go                 # OpenSearch-specific DTOs
```

Search specifics:
- No `event_listeners/`
- `adapters/outbound/` uses `opensearch/` instead of `database/` — read repository
- `opensearch/` has extra files: `builders.go`, `mappers.go` (query construction + response mapping)
- `application/mocks/` has only `repository_mock.go` (fewer dependencies than create/update)

#### Feature: update

```
features/update/
├── application/
│   ├── usecase.go / usecase_test.go
│   ├── new_usecase.go
│   ├── interface.go
│   ├── dto.go
│   ├── constants.go
│   └── mocks/                         # 4 mocks: event_dispatcher, logger, repository, tracer
├── event_listeners/
│   ├── send_status_email/             # (empty — placeholder)
│   └── update_data_repositories/      # Sync data to read stores
│       ├── listener/                  # Handler logic (mirrors application/ pattern)
│       │   ├── handler.go / handler_test.go
│       │   ├── new_handler.go
│       │   ├── interface.go
│       │   ├── dto.go
│       │   ├── constants.go
│       │   └── mocks/
│       └── adapters/
│           ├── inbound/               # SQS consumer
│           └── outbound/              # OpenSearch repository for index sync
└── adapters/
    ├── inbound/
    │   ├── http_handler/              # handler.go, handler_test.go, dto.go, di.go, mocks/
    │   └── grpc_service/              # service.go, service_test.go, dto.go, di.go, mocks/, pb/, proto/
    └── outbound/
        └── database/                  # repository.go, new.go, errors.go, repository_test.go
```

Update specifics:
- `event_listeners/update_data_repositories/` syncs specialist data to OpenSearch after updates
- `send_status_email/` is empty (placeholder)
- No `handler_integration_test.go` at the `adapters/inbound/` root

## internal/platform/ — Infrastructure Adapters

```
platform/
├── database/postgresql/
│   ├── connection.go                  # DSN construction, connection pool setup
│   ├── migrations.go                  # Goose migration runner
│   └── migrations/                    # SQL migration files (ordered: 001_, 002_, 003_)
├── opensearch/
│   ├── client.go                      # OpenSearch client with optional AWS SigV4 auth
│   ├── factory.go                     # Factory: client init + index creation + prefix management
│   └── indexes/
│       ├── registry.go                # Index registry with prefix support (multi-tenant)
│       └── specialists.go             # Specialist index mapping with custom analyzers
├── sqs/
│   ├── client.go                      # AWS SDK v2 SQS client initialization
│   ├── producer.go                    # SQSProducer implements EventDispatcher interface
│   ├── consumer.go                    # SQS consumer with Long Polling + delete after success
│   ├── ensure.go                      # Idempotent queue creation with DLQ setup
│   └── health.go                      # Health check for SQS queues
├── server/
│   ├── grpcserver.go                  # gRPC server with keepalive, health check, reflection
│   └── httpserver.go                  # HTTP server with CORS, health endpoint, Swagger
└── telemetry/
    ├── otel.go                        # OTel tracer wrapper
    ├── provider.go                    # OTel resource + MeterProvider initialization
    ├── tracing_provider.go            # TracerProvider setup (OTLP gRPC export)
    ├── metrics_provider.go            # MeterProvider setup (OTLP HTTP export)
    ├── metrics.go                     # OtelMetrics: Counter, Histogram, Gauge via OTel SDK
    ├── grpc_metrics.go                # gRPC unary interceptor for request metrics
    ├── http_metrics.go                # Gin middleware for HTTP request metrics
    ├── slog.go                        # Structured logging with slog (JSON output)
    └── factory.go                     # Telemetry factory utilities
```

## tests/k6/ — Load & Stress Tests

```
tests/k6/
├── Makefile
├── docker-compose.k6.yml
├── commom/
│   ├── factories/specialist.js        # Test data factory
│   ├── grpc-client.js                 # gRPC client helper
│   ├── http-client.js                 # HTTP client helper
│   ├── http-stress-runner.js          # HTTP stress test runner
│   └── stress-runner.js              # gRPC stress test runner
└── modules/specialist/create/
    ├── stress/                        # gRPC stress tests
    │   ├── config.js
    │   ├── create-specialist-test.js
    │   ├── validations.js
    │   └── proto/specialist.proto     # Proto file duplicated for k6
    └── stress-http/                   # HTTP stress tests
        ├── config.js
        ├── create-specialist-test.js
        └── validations.js
```

## Structural Variations Between Features

```yaml
create:
  event_listeners: yes (validate_license fully implemented via SQS, send_credentials_email empty)
  adapters/outbound/database: yes
  adapters/outbound/opensearch: no
  adapters/inbound/http_handler: yes
  adapters/inbound/grpc_service: yes
  handler_integration_test: yes (at adapters/inbound/ root)
  application/mocks: 4 (event_dispatcher, logger, repository, tracer)
  domain sub-packages: create/ (with separate uniqueness_errors)

search:
  event_listeners: no
  adapters/outbound/database: no
  adapters/outbound/opensearch: yes (with builders, mappers, own dto)
  adapters/inbound/http_handler: yes
  adapters/inbound/grpc_service: yes
  handler_integration_test: no
  application/mocks: 1 (repository)
  domain sub-packages: search_input/ + search_output/ (value objects as sub-packages)

update:
  event_listeners: yes (update_data_repositories with SQS consumer + OpenSearch sync, send_status_email empty)
  adapters/outbound/database: yes
  adapters/outbound/opensearch: no
  adapters/inbound/http_handler: yes
  adapters/inbound/grpc_service: yes
  handler_integration_test: no
  application/mocks: 4 (event_dispatcher, logger, repository, tracer)
  domain sub-packages: update/ (with validators separate from update.go)
```

## File Naming Patterns by Layer

- `usecase.go` + `new_usecase.go` — application layer (logic and constructor separated)
- `handler.go` + `new_handler.go` — listener layer (same pattern)
- `repository.go` + `new.go` — adapters/outbound/database and adapters/outbound/opensearch
- `gateway.go` + `new.go` — event_listeners/*/adapters/outbound/external
- `service.go` + `di.go` — adapters/inbound/grpc_service (DI separated from service)
- `handler.go` + `di.go` — adapters/inbound/http_handler
- `dto.go` — present in all layers (application, listener, adapters/inbound/grpc_service, adapters/inbound/http_handler, adapters/outbound/opensearch)
- `constants.go` — application and listener (span names, event names, error messages)
- `errors.go` — domain, application (via constants.go), adapters/outbound/database, adapters/outbound/opensearch
- `interface.go` — application and listener (dependency contracts)

## Package Dependencies Flow

```
domain (entity, status, errors, validate)
    ↑
domain/<feature>/ (factory, validation, feature-specific errors)
    ↑
features/<feature>/application/ (use case orchestration, interfaces, DTOs)
    ↑
features/<feature>/adapters/inbound/ (grpc_service, http_handler)
features/<feature>/adapters/outbound/ (database, opensearch)

features/<feature>/event_listeners/<listener>/listener/ (handler, interfaces, DTOs)
    ↑
features/<feature>/event_listeners/<listener>/adapters/ (inbound/sqs, outbound/database, outbound/external)
```

Event listeners are independent mini-modules within a feature, with their own listener/ + adapters/ separation.
