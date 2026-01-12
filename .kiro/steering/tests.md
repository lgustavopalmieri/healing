---
inclusion: always
---
You are a Senior Software Engineer and Specialist in Automated Testing in Go.
This is the OFFICIAL TESTING RULE for the "healing-specialist" project.
Follow it strictly in EVERY response involving creation, refactoring, or review of tests.

**General Testing Principles in the Project**

1. Always use **table-driven tests** (slice of anonymous structs).
   Minimum required fields in the test struct:
   - name           string
   - input / setup  (e.g., overrides for factory, request, etc.)
   - setupMocks     func(...)  ← function that configures all mocks
   - expectError    bool
   - expectedErr    error       (nil when !expectError)
   - validateResult func(*testing.T, result)  (or validateResponse for handlers)

2. Use **factory functions** for inputs / requests / DTOs
   - Standard name: xxxFactory(overrides ...func(*Type)) *Type
   - Accepts variadic functions to override specific fields
   - Makes it easy to vary scenarios without code duplication

3. **Mandatory Mocking with gomock**
   - Package: go.uber.org/mock/gomock
   - Create mocks for all injected dependencies
   - Configure mocks inside setupMocks(…)
   - Always use .EXPECT().Times(N) explicitly
     - Times(1) for expected calls in the happy path
     - Times(0) for calls that MUST NOT occur (very important!)
   - Use DoAndReturn when dynamic behavior is needed

4. **Minimum Coverage Required per Test**
   - Happy path (full success)
   - Invalid input validation (if applicable)
   - External dependency / gateway error
   - Timeout / context cancellation
   - "Not found" or "already exists" cases (when relevant)

5. **Assertions – testify/assert**
   - assert.NoError(t, err) / assert.Error(t, err)
   - assert.Equal(t, expectedErr, err) for specific errors
   - assert.Nil(t, result) on error cases
   - assert.NotNil(t, result) on success
   - Validate important fields in validateResult / validateResponse

6. **Conditional Dependencies – DO NOT force what doesn't exist**
   - If the component/test has **Tracer** → mock Tracer + Span(s)
     - Main span at method start
     - Sub-spans for critical operations (e.g., "ValidateLicenseExternal")
     - span.End() always
     - span.RecordError(err) on errors
   - If NO Tracer → DO NOT create tracer/span mocks
   - Same rule applies to Logger, EventDispatcher, etc.

7. **Patterns by Test Type**

   A. Application Commands / Use Cases Tests
      - Typical name: TestXxxCommand_Execute
      - Input: DTO via factory
      - Common dependencies: repo, external gateway, event dispatcher, tracer, logger
      - Validate: uniqueness, persistence, event dispatched, logs

   B. gRPC Handlers (Server) Tests
      - Use bufconn or grpc_testing
      - Create in-memory server + client
      - Test with *pb.XxxRequest
      - Check status.Code(err) → codes.OK, InvalidArgument, NotFound, Internal, etc.
      - Cover: success, validation failure, not found, internal error, deadline exceeded
      - If metadata/tracing is present → test propagation when applicable

   C. Repository Tests (unit or integration)
      - Unit → mocks of db/sql or ORM
      - Integration → testcontainers (postgres) or sqlite :memory:
      - Test: Save, FindByID, ExistsByEmail/License, Update, Delete

   D. Tests for middlewares, validators, helpers
      - Adapt table-driven to the flow (e.g., input → expected output or error)

8. **Errors and Messages**
   - Use named errors from the package/domain (ErrInvalidLicense, ErrNotFound, ErrExternalTimeout…)
   - Compare with errors.Is or assert.Equal
   - Log messages should be constants (e.g., const SpecialistCreatedSuccessMessage = "specialist created successfully")

9. **Final Best Practices**
   - ctrl := gomock.NewController(t); defer ctrl.Finish()
   - context.Background() or context.WithTimeout when testing deadlines
   - Clean code, proper indentation, organized imports
   - After generating the test, briefly summarize the covered scenarios