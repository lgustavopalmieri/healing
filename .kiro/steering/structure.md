---
inclusion: always
---

# Project Structure & Organization

## Directory Layout

```
internal/
├── commom/                    # Shared utilities across modules
│   ├── event/                 # Event system (dispatcher, listeners)
│   ├── observability/         # Logging, metrics, tracing interfaces
│   ├── tests/                 # Test utilities and helpers
│   │   ├── database/          # Database testing utilities (testcontainers)
│   │   └── event/             # Event testing utilities (Kafka mocks)
│   └── utils/                 # Common utilities (sanitization, etc.)
├── modules/                   # Business modules
│   └── specialist/            # Specialist domain module
│       ├── domain/            # Domain layer
│       │   ├── create/        # Create specialist domain logic (package: create)
│       │   │   ├── create.go          # Factory function and input struct
│       │   │   ├── validators.go      # Domain validation rules
│       │   │   ├── errors.go          # Domain-specific errors
│       │   │   ├── create_test.go     # Domain logic tests
│       │   │   └── validators_test.go # Validator tests
│       │   └── entity.go      # Specialist entity definition (package: domain)
│       └── features/          # Feature-based organization
│           ├── create/        # Create specialist feature
│           │   ├── application/       # Application layer
│           │   │   ├── mocks/         # Generated gomock mocks
│           │   │   ├── command.go     # Command execution logic
│           │   │   ├── new_command.go # Command constructor
│           │   │   ├── interface.go   # Repository and gateway interfaces
│           │   │   ├── dto.go         # Data transfer objects
│           │   │   ├── constants.go   # Application constants and errors
│           │   │   └── command_test.go # Command tests
│           │   └── infra/             # Infrastructure layer
│           │       ├── database/      # Database implementation
│           │       │   ├── repository.go      # PostgreSQL repository
│           │       │   ├── new.go             # Repository constructor
│           │       │   ├── errors.go          # Database-specific errors
│           │       │   └── repository_test.go # Repository tests
│           │       ├── external/      # External service integrations
│           │       │   └── license_gateway.go # License validation gateway
│           │       └── grpc_service/  # gRPC transport layer
│           │           ├── mocks/     # Generated gomock mocks
│           │           ├── pb/        # Generated protobuf code
│           │           ├── proto/     # Protocol buffer definitions
│           │           ├── service.go # gRPC service implementation
│           │           ├── dto.go     # gRPC DTO conversions
│           │           ├── di.go      # Dependency injection setup
│           │           └── service_test.go # gRPC service tests
│           └── get_by_id/     # Get specialist by ID feature (placeholder)
└── platform/                  # Platform infrastructure
    ├── database/              # Database connection management
    │   └── postgresql/        # PostgreSQL implementation
    │       ├── connection.go  # Connection pooling
    │       ├── migrations.go  # Migration runner
    │       └── migrations/    # SQL migration files
    ├── kafka/                 # Kafka producer/consumer
    │   ├── producer.go        # Event producer
    │   └── consumer.go        # Event consumer
    ├── server/                # Server implementations
    │   ├── grpcserver.go      # gRPC server setup
    │   └── metrics_server.go  # Prometheus metrics server
    └── telemetry/             # Observability implementations
        ├── otel.go            # OpenTelemetry setup
        ├── tracing_provider.go # Tracing configuration
        ├── prometheus.go      # Prometheus metrics
        ├── slog.go            # Structured logging
        ├── grpc_metrics.go    # gRPC metrics interceptors
        └── provider.go        # Telemetry provider
```

## Architectural Layers

### Domain Layer (`domain/`)
The domain layer is split into two packages:

**Package `domain`** (`internal/modules/specialist/domain/`):
- **entity.go**: Core business entity (`Specialist` struct)
- Pure Go struct with no business logic, only data structure

**Package `create`** (`internal/modules/specialist/domain/create/`):
- **create.go**: Factory function `CreateSpecialist()` and `CreateSpecialistInput` struct
- **validators.go**: Private validation functions (validateName, validateEmail, etc.)
- **errors.go**: Domain-specific errors (ErrInvalidName, ErrInvalidEmail, ErrInvalidLicense, etc.)
- **Tests**: Comprehensive unit tests for all domain logic
- **Pure Go**: No external dependencies except standard library and `internal/commom/utils`
- **Imports**: Must import parent `domain` package to access `Specialist` entity

**Key principle**: Domain logic is isolated from infrastructure concerns. The `create` package handles all business rules for specialist creation, while the `domain` package defines the entity structure.

### Application Layer (`application/`)
- **Commands**: Use case implementations (e.g., `CreateSpecialistCommand`)
- **DTOs**: Data transfer objects for input/output
- **Interfaces**: Contracts for external dependencies (repositories, gateways)
- **Constants**: Application-level constants, error messages, and span names
- **Constructor**: Dependency injection via `NewCreateSpecialistCommand()`
- **Orchestration**: Coordinates domain logic, external validations, persistence, and events

### Infrastructure Layer (`infra/`)

**Database** (`infra/database/`):
- PostgreSQL repository implementation
- Uniqueness constraint validation
- Integration tests using testcontainers

**External** (`infra/external/`):
- External service integrations (e.g., license validation gateway)
- HTTP/REST client implementations

**gRPC Service** (`infra/grpc_service/`):
- gRPC transport layer
- Protocol buffer definitions (`.proto` files)
- Generated protobuf code (`pb/` directory)
- DTO conversions between protobuf and domain models
- Dependency injection setup

## Naming Conventions

### Files & Packages
- Package names: lowercase, single word when possible
- Nested domain packages: named after their purpose (e.g., `create`, `update`)
- File names: snake_case (e.g., `grpc_service`, `create_test.go`)
- Test files: `*_test.go` suffix
- Mock files: `*_mock.go` suffix in dedicated `mocks/` directories

### Code Structure
- **Interfaces**: End with `Interface` suffix (e.g., `RepositoryInterface`, `SpecialistCreateRepositoryInterface`)
- **Commands**: End with `Command` suffix (e.g., `CreateSpecialistCommand`)
- **DTOs**: End with `DTO` suffix (e.g., `CreateSpecialistDTO`)
- **Errors**: Start with `Err` prefix (e.g., `ErrInvalidLicense`, `ErrDuplicateEmail`)
- **Constants**: Descriptive names with context (e.g., `SpecialistCreatedEventName`, `CreateSpecialistSpanName`)
- **Constructors**: Start with `New` prefix (e.g., `NewCreateSpecialistCommand`, `NewRepository`)

### Import Conventions
- Domain packages import parent packages when needed (e.g., `domain/create` imports `domain`)
- Application layer imports both `domain` and `domain/create` packages
- Infrastructure imports application interfaces and domain entities
- Use full import paths: `github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/create`

## Feature Organization

Each feature follows a consistent three-layer structure:

**Domain Logic** (`domain/<feature>/`):
- Factory functions and business rules
- Validation logic
- Domain-specific errors
- Completely isolated from infrastructure

**Application Layer** (`features/<feature>/application/`):
- Use case orchestration (Commands)
- Interface definitions for dependencies
- DTOs for data transfer
- Application-level constants and errors
- Generated mocks for testing

**Infrastructure Layer** (`features/<feature>/infra/`):
- Concrete implementations of application interfaces
- External service integrations
- Database repositories
- gRPC/HTTP transport layers
- Protocol buffer definitions

Features are self-contained with their own dependencies and tests, enabling independent development and deployment.

## Common Patterns

### Dependency Injection
- Constructor functions with interface parameters (e.g., `NewCreateSpecialistCommand()`)
- Interfaces defined in application layer, implemented in infrastructure
- All dependencies injected at construction time

### Context Propagation
- All operations accept `context.Context` as first parameter
- Context used for cancellation, timeouts, and tracing
- Spans created from context for distributed tracing

### Error Handling
- Domain errors defined in `domain/*/errors.go`
- Application errors defined in `application/constants.go`
- Domain errors bubble up through layers unchanged
- Infrastructure errors wrapped with application-level errors

### Observability
- Structured logging with fields using `observability.Logger`
- Distributed tracing with spans using `observability.Tracer`
- Metrics collection via Prometheus
- All operations logged with relevant context (IDs, emails, etc.)

### Testing Strategy
- **Domain tests**: Pure unit tests with no mocks
- **Application tests**: Table-driven tests with gomock for dependencies
- **Infrastructure tests**: Integration tests with testcontainers (database) or real services
- **Factory functions**: Used in tests for flexible test data creation with override patterns

### Package Dependencies Flow
```
domain (entity) ← domain/create (factory + validation)
                       ↑
                  application (orchestration)
                       ↑
                infrastructure (implementations)
```

Domain packages never import application or infrastructure. Application never imports infrastructure. This ensures clean separation of concerns and testability.