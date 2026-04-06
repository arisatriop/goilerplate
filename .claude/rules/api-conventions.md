# API Conventions

## Routing
- Use RESTful naming: plural nouns (`/users`, `/roles`, `/bars`)
- Nested resources for owned entities: `/users/{id}/roles`
- Routes are registered in `internal/delivery/http/router/` — `public.go`, `internal.go`, or `partner.go` depending on auth requirements

## Request Validation
- Use `go-playground/validator` struct tags on request DTOs in `delivery/http/dto/request/`
- Validate at the handler layer before passing to use case
- Return 400 Bad Request with validation error details on invalid input

## Response Envelope
All responses must use a consistent envelope via the pkg response helper:
```json
{
  "success": true,
  "message": "...",
  "data": { ... }
}
```
For paginated responses:
```json
{
  "success": true,
  "message": "...",
  "data": [...],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 100
  }
}
```

## HTTP Status Codes
| Situation | Status |
|---|---|
| Success GET/PUT/PATCH | 200 |
| Success POST (created) | 201 |
| Bad input / validation error | 400 |
| Unauthorized (no/invalid token) | 401 |
| Forbidden (valid token, no permission) | 403 |
| Resource not found | 404 |
| Server/logic error | 500 |

## Authentication
- Public routes: no auth middleware
- Internal routes (`router/internal.go`): require Bearer JWT via `middleware.Auth`
- Partner routes (`router/partner.go`): require API key via `middleware.APIKey`
- Auth data is stored in `ctx.Locals` — retrieve via the auth middleware helper, not raw header parsing

## Query Parameters
- Pagination: `?page=1&limit=10`
- Filtering: use descriptive names matching the field (e.g. `?status=active&role_id=1`)
- Sorting: `?sort_by=created_at&order=desc`

## Fiber-Specific
- Parse request body with `ctx.BodyParser(&req)`
- Return errors using `ctx.Status(code).JSON(...)` — don't use `c.Next()` for error propagation
- Use `ctx.Locals("user")` to access authenticated user data set by auth middleware
