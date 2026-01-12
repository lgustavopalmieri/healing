# Technology Stack

## Language & Runtime
- **Go 1.25.0** - Primary programming language
- Standard library focused with minimal external dependencies

## Key Dependencies
- `github.com/google/uuid` - UUID generation for entity IDs
- `github.com/stretchr/testify` - Testing framework and assertions

## Architecture Patterns

### Observability
- **Structured Logging**: Context-aware logging with key-value fields
- **Distributed Tracing**: OpenTelemetry-style span creation and error recording
- **Metrics**: Dedicated metrics collection (see `internal/commom/observability/`)

### Event-Driven Architecture
- **Event Dispatcher**: Asynchronous event publishing for domain events
- **Event Listeners**: Decoupled event handling
- Events use `any` type for flexible payloads with UTC timestamps

### Domain-Driven Design
- **Clean Architecture**: Clear separation of domain, application, and infrastructure layers
- **Command Pattern**: Application services implement command execution
- **Repository Pattern**: Data persistence abstraction
- **Domain Events**: Business events published after successful operations

## Common Commands

### Testing
```bash
# Run all tests with verbose output
go test -v ./...

# Run tests without cache (for debugging)
go test -v -count=1 ./...

# Run tests with race detection
go test -v -race ./...

# Run specific test
go test -v ./path/to/package -run ^TestName$
```

### Coverage
```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

### Build
```bash
# Build the application
go build ./...

# Run with Go modules
go mod tidy
go mod download
```

## Code Conventions
- Use `context.Context` for all operations requiring cancellation or timeouts
- Implement proper error handling with custom domain errors
- Use structured logging with contextual fields
- Follow Go naming conventions (PascalCase for exported, camelCase for unexported)
- Prefer composition over inheritance