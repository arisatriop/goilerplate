# Testing Rules

## General
- Run tests with: `go test ./...` or `make test`
- Tests live alongside the code they test (`*_test.go` in the same package)
- Test file naming: `<filename>_test.go`

## What to Test
- Business logic in the use-case (application) layer is the highest priority
- Repository layer: test with real DB or integration tests, not mocks
- Handler layer: test HTTP status codes and response shapes
- Financial calculations must have dedicated unit tests with edge cases

## Mocking
- Do NOT mock the database in tests — use real PostgreSQL (test DB)
- Mock external integrations (S3, payment gateways) using interfaces
- Keep mocks minimal — only mock what you actually need

## Test Structure
```go
func TestUsecaseName_MethodName(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

## Financial Test Cases
When testing financial calculations, always include:
- Zero values
- Negative values (if applicable)
- Large numbers (overflow risk)
- Decimal precision cases (e.g. 0.1 + 0.2)

## Coverage
- New business logic must have at least basic happy-path + one error-path test
- Don't aim for 100% coverage on boilerplate — focus on logic that can fail
