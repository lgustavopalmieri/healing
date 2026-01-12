---
inclusion: always
---

# Technology Stack & Build System

## Tech Stack

**Language**: Go 1.25.0

**Core Dependencies**:
- `google.golang.org/grpc` - gRPC server/client communication
- `google.golang.org/protobuf` - Protocol Buffers for API definitions
- `github.com/google/uuid` - UUID generation for entity IDs
- `github.com/stretchr/testify` - Testing assertions and test suites
- `go.uber.org/mock` - Mock generation for testing

## Build & Development Commands

### Protocol Buffer Generation
```bash
# Install required tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from .proto files
protoc --go_out=. --go-grpc_out=. proto/specialist.proto
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
```

### gRPC Testing with Evans
```bash
# Install Evans CLI tool
go install github.com/ktr0731/evans@latest

# Interactive gRPC testing
evans -r repl
# Then: show package -> package pb -> show service -> service SpecialistService -> call CreateSpecialist
```

## Code Generation

- **Mocks**: Use `go.uber.org/mock/gomock` with `//go:generate` directives
- **Protocol Buffers**: Manual generation using `protoc` commands
- Proto files located in `*/proto/` directories, generated code in `*/pb/` directories

## Testing Philosophy

- **Table-driven tests** are mandatory for all test cases
- **gomock** for all dependency mocking with explicit `.Times()` calls
- **Factory functions** for test data creation with override patterns
- Comprehensive coverage: happy path, validation errors, external failures, timeouts