# Project Structure

## Root Level
```
├── go.mod, go.sum          # Go module definition and dependencies
├── coverage.out            # Test coverage output file
├── cmds.md                 # Common commands documentation
└── internal/               # Private application code
```

## Internal Architecture

### Common Shared Components (`internal/commom/`)
```
internal/commom/
├── event/                  # Event-driven architecture components
│   ├── dipstacher.go      # Event dispatcher implementation
│   ├── event.go           # Event entity and factory
│   └── listener.go        # Event listener interface
├── observability/         # Cross-cutting observability concerns
│   ├── logging.go         # Structured logging interface
│   ├── metrics.go         # Metrics collection
│   └── tracing.go         # Distributed tracing
└── utils/                 # Shared utility functions
    ├── sanitize.go        # String sanitization utilities
    └── sanitize_test.go   # Utility tests
```

### Module Structure (`internal/modules/`)
Each business module follows a consistent structure:

```
internal/modules/{module}/
├── domain/                # Pure domain logic
│   ├── entity.go         # Domain entities
│   ├── create.go         # Entity creation logic
│   ├── validators.go     # Domain validation rules
│   ├── errors.go         # Domain-specific errors
│   └── *_test.go         # Domain tests
└── features/             # Feature-based organization
    └── {feature}/        # Individual feature (e.g., create, list)
        ├── application/  # Application services
        │   ├── command.go      # Command implementation
        │   ├── dto.go          # Data transfer objects
        │   ├── interface.go    # Port definitions
        │   ├── constants.go    # Feature constants
        │   └── new_command.go  # Command factory
        └── infra/        # Infrastructure adapters
```

## Naming Conventions

### Packages
- `domain` - Pure business logic, no external dependencies
- `application` - Use case orchestration, depends on domain
- `infra` - External integrations, implements application interfaces

### Files
- `entity.go` - Domain entities and value objects
- `create.go` - Entity creation and factory methods
- `validators.go` - Domain validation logic
- `errors.go` - Domain-specific error definitions
- `command.go` - Application command implementation
- `dto.go` - Data transfer objects for application layer
- `interface.go` - Port definitions (interfaces)
- `constants.go` - Application constants and messages

### Testing
- Test files follow `*_test.go` naming convention
- Tests are co-located with the code they test
- Use table-driven tests where appropriate

## Architectural Boundaries

### Domain Layer
- Contains pure business logic
- No external dependencies except standard library
- Defines domain entities, value objects, and business rules
- Validates business invariants

### Application Layer  
- Orchestrates use cases
- Depends on domain layer
- Defines ports (interfaces) for external dependencies
- Handles cross-cutting concerns (logging, tracing, events)

### Infrastructure Layer
- Implements application interfaces
- Handles external integrations
- Contains adapters for databases, APIs, message queues

## Module Independence
- Each module in `internal/modules/` represents a bounded context
- Modules should minimize dependencies on each other
- Shared concerns are extracted to `internal/commom/`
- Communication between modules should be through events when possible