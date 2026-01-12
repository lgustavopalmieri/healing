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
│   └── utils/                 # Common utilities (sanitization, etc.)
└── modules/                   # Business modules
    └── specialist/            # Specialist domain module
        ├── domain/            # Domain entities, validators, errors
        └── features/          # Feature-based organization
            └── create/        # Create specialist feature
                ├── application/   # Use cases, commands, DTOs
                │   └── mocks/     # Generated mocks for testing
                └── infra/         # Infrastructure layer
                    └── grpc_handler/  # gRPC transport layer
                        ├── pb/        # Generated protobuf code
                        └── proto/     # Protocol buffer definitions
```

## Architectural Layers

### Domain Layer (`domain/`)
- **Entities**: Core business objects (e.g., `Specialist`)
- **Validators**: Business rule validation logic
- **Errors**: Domain-specific error definitions
- **Pure Go**: No external dependencies, only standard library

### Application Layer (`application/`)
- **Commands**: Use case implementations (e.g., `CreateSpecialistCommand`)
- **DTOs**: Data transfer objects for input/output
- **Interfaces**: Contracts for external dependencies
- **Constants**: Application-level constants and messages

### Infrastructure Layer (`infra/`)
- **gRPC Handlers**: Transport layer for gRPC services
- **Protocol Buffers**: API definitions and generated code
- **External integrations**: Database, external APIs, message queues

## Naming Conventions

### Files & Packages
- Package names: lowercase, single word when possible
- File names: snake_case (e.g., `grpc_handler`, `create_test.go`)
- Test files: `*_test.go` suffix
- Mock files: `*_mock.go` suffix in dedicated `mocks/` directories

### Code Structure
- **Interfaces**: End with `Interface` suffix (e.g., `RepositoryInterface`)
- **Commands**: End with `Command` suffix (e.g., `CreateSpecialistCommand`)
- **DTOs**: End with `DTO` suffix (e.g., `CreateSpecialistDTO`)
- **Errors**: Start with `Err` prefix (e.g., `ErrInvalidLicense`)
- **Constants**: Descriptive names with context (e.g., `SpecialistCreatedEventName`)

## Feature Organization

Each feature follows the same structure:
- `application/` - Business logic and use cases
- `infra/` - External integrations and transport layers
- Features are self-contained with their own dependencies and tests

## Common Patterns

- **Dependency Injection**: Constructor functions with interface parameters
- **Context Propagation**: All operations accept `context.Context` as first parameter
- **Error Handling**: Domain errors bubble up through layers unchanged
- **Observability**: Structured logging with fields, distributed tracing spans