# Project Structure

## Directory Organization

```
internal/
├── commom/                    # Shared infrastructure components
│   ├── event/                 # Event system (dispatcher, listeners)
│   ├── observability/         # Logging, metrics, tracing interfaces
│   └── utils/                 # Common utilities (sanitization, etc.)
└── modules/                   # Business modules
    └── specialist/            # Specialist domain module
        ├── domain/            # Domain entities, value objects, business rules
        └── features/          # Feature-based organization
            ├── create/        # Create specialist feature
            │   ├── application/   # Commands, DTOs, interfaces
            │   │   └── mocks/     # Generated mocks for testing
            │   └── infra/         # Infrastructure implementations
            └── list/          # List specialists feature
```

## Architectural Layers

### Domain Layer (`domain/`)
- **Entities**: Core business objects (e.g., `Specialist`)
- **Value Objects**: Immutable objects representing concepts
- **Business Rules**: Domain validation and business logic
- **Errors**: Domain-specific error definitions

### Application Layer (`features/*/application/`)
- **Commands**: Use case implementations following command pattern
- **DTOs**: Data transfer objects for input/output
- **Interfaces**: Contracts for external dependencies
- **Constants**: Application-level constants and messages

### Infrastructure Layer (`features/*/infra/`)
- **Repositories**: Data persistence implementations
- **External Services**: Third-party integrations
- **Adapters**: Interface implementations

## Naming Conventions

### Files
- `entity.go` - Domain entities
- `create.go` - Domain creation logic
- `validators.go` - Domain validation rules
- `command.go` - Application command implementation
- `interface.go` - Dependency contracts
- `*_test.go` - Test files alongside source
- `*_mock.go` - Generated mocks in `/mocks` subdirectory

### Packages
- Use singular nouns for package names
- Feature packages named by action (e.g., `create`, `list`)
- Common utilities in `commom` (note: typo in original, maintain consistency)

## Testing Structure
- Tests live alongside source files with `_test.go` suffix
- Mocks generated in dedicated `/mocks` subdirectories
- Use testify for assertions and test structure
- Integration tests may span multiple layers

## Event System
- Events defined in `internal/commom/event/`
- Domain events published after successful operations
- Event names use descriptive constants (e.g., `SpecialistCreatedEventName`)

## Error Handling
- Domain errors defined in domain layer
- Application errors for use case failures
- Structured error messages with context