# Technology Stack

## Language & Runtime
- **Go 1.25.0** - Primary programming language
- Modern Go features and idioms expected

## Dependencies
- **github.com/google/uuid** - UUID generation for entity IDs
- **github.com/stretchr/testify** - Testing framework and assertions
- **go.uber.org/mock** - Mock generation for testing

## Architecture Patterns
- **Domain-Driven Design (DDD)** - Clear domain/application/infrastructure separation
- **Command Pattern** - Application layer uses command objects for operations
- **Event-Driven Architecture** - Domain events published after operations
- **Clean Architecture** - Dependencies point inward toward domain

## Observability
- **Structured Logging** - Context-aware logging with key-value fields
- **Distributed Tracing** - Span-based tracing for request flows
- **Metrics** - Performance and business metrics collection

## Common Commands

### Testing
```bash
# Run all tests with verbose output
go test -v ./...

# Run tests without cache (for debugging)
go test -v -count=1 ./...

# Run specific package tests
go test ./internal/modules/specialist/features/create/application -v -timeout=30s

# Run with race detector
go test -v -race ./...
```

### Coverage
```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

### Mock Generation
```bash
# Generate repository mocks
mockgen -source=internal/modules/specialist/features/create/application/interface.go -destination=internal/modules/specialist/features/create/application/mocks/repository_mock.go -package=mocks

# Generate event dispatcher mocks
mockgen -source=internal/commom/event/dipstacher.go -destination=internal/modules/specialist/features/create/application/mocks/event_dispatcher_mock.go -package=mocks
```