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
For list responses, use `response.Paginated` — `data` stays a plain array and the page
counters go in `meta`:

```json
{
  "success": true,
  "message": "...",
  "data": [...],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "totalPages": 10,
    "hasNext": true,
    "hasPrev": false
  }
}
```

```go
page := response.NewPagination(total, filter.Pagination.Page, filter.Pagination.Limit)
return response.Paginated(ctx, items, page, response.WithMessage(domain.MsgListFetched))
```

`data` is always an array, `[]` rather than `null` on an empty page, so a client can iterate it
without a nil check. `pkg/pagination` parses `page` and `limit` off the request; it does not
describe the response.

All keys are camelCase — see [docs/api/router.md](../../docs/api/router.md#-response-envelope).
`TestPaginated_ExactShape` and `TestBarList_ReturnsTheDocumentedShape` assert this shape byte
for byte, so a change here that is not made in both places fails the build.

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
- Pagination: `?page=1&limit=10`. Parse with `pagination.ParsePagination(ctx)`, which clamps
  out-of-range values rather than rejecting them: `limit` is capped at 100, and a page past the
  end returns an empty page. Asking for a million rows should not make the database pay for it
- Filtering: use descriptive names matching the field (e.g. `?status=active&role_id=1`)
- Sorting: `?sort_by=created_at&order=desc`

## Fiber-Specific
- Parse request body with `ctx.BodyParser(&req)`
- Return errors using `ctx.Status(code).JSON(...)` — don't use `c.Next()` for error propagation
- Use `ctx.Locals("user")` to access authenticated user data set by auth middleware
