# Code Style Rules

## General
- Follow standard Go conventions (gofmt, golangci-lint clean)
- Prefer early returns to reduce nesting
- Max function length: ~60 lines; split if longer
- No unused imports or variables
- Use meaningful variable names — avoid single-letter vars outside loop counters

## Naming
- Exported types/functions: PascalCase
- Unexported: camelCase
- Constants: PascalCase or ALL_CAPS for config keys
- Files: snake_case (e.g. `user_repository.go`)

## Error Handling
- Always wrap errors with context: `fmt.Errorf("doing X: %w", err)` or `pkg/errors`
- Never silently swallow errors
- Return errors up to the handler layer; don't log and return
- Use domain-defined errors (`domain/<pkg>/error.go`) for expected error cases

## Financial Values
- **Always use `github.com/shopspring/decimal`** for money, rates, fees, and any financial calculation
- Never use `float64` for financial values — precision errors will cause bugs in production
- Use `decimal.NewFromString()` when parsing user input

## Clean Architecture Boundaries
- `domain/` has zero external dependencies (no GORM, no Fiber, no infrastructure imports)
- `application/` depends only on `domain/` interfaces — never on `infrastructure/` directly
- `infrastructure/` implements `domain/` interfaces — never imported by `application/` or `delivery/`
- `delivery/` depends on `domain/` interfaces (Usecase) only
- `wire/` is the only place that knows about all concrete implementations

## Structs & Models
- Domain entities live in `domain/<name>/entity.go` — no GORM tags
- GORM models live in `infrastructure/model/` — separate from domain entities
- Use `json` struct tags for API DTOs
- Use `validate` struct tags for request validation
- Separate request DTOs (`delivery/http/dto/request/`) from response DTOs (`delivery/http/dto/response/`) — never expose DB models directly

## Dependencies
- Don't add new dependencies without discussion
- Prefer stdlib or already-imported packages when possible
