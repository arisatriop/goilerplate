# API Conventions

These follow common REST practice (Microsoft REST API Guidelines, Google AIP, Zalando RESTful
API Guidelines, RFC 9110) where those agree, and state this project's choice where they don't.
Gaps between these rules and today's code are tracked in
`docs/roadmap/improvement-tasks.md` (phase R).

## Resources and routing
- Versioned under `/api/v1`. A breaking change gets a new version. Before the first release the
  contract may still change freely (Appendix A, decision D4)
- Collections are plural nouns. Multi-word path segments are kebab-case (`/logout-all`). Ownership
  nests: `/users/me/sessions/{id}`, with `me` standing for the authenticated caller
- Non-CRUD operations are a `POST` to a verb sub-resource (`/auth/logout`). Use them sparingly;
  most operations are CRUD on a resource
- Routes are registered in `internal/delivery/http/router/` by audience:

  | File | Prefix | Who calls it | Guard |
  |---|---|---|---|
  | `public.go` | `/api/v1` | End users. Login and registration are unauthenticated; everything else requires `Auth.Authenticate()` | JWT, per-route `RequiredPermission`, rate limit per IP / session / user |
  | `internal.go` | `/internal` | Other services in the cluster | Gateway allowlist, plus `InternalAuthenticate()` when `internal_auth.mode=shared_secret` |
  | `partner.go` | `/partner/v1` | External partners | `PartnerAuthenticate()` (API key) and a per-key rate limit |

- Every authenticated `/api/v1` route checks permission with `RequiredPermission`. The exception is
  a `/users/me/...` route: its subject comes from the token, so it can only reach the caller's
  own data. `/internal` and `/partner` authorise the calling service or partner, not a user

## Methods and status codes

| Situation | Status |
|---|---|
| `GET`, `PUT`, `PATCH`, action `POST` succeeded | 200 |
| `POST` created a resource | 201, with a `Location` header naming it (gap R4) |
| `DELETE` succeeded | 200 with the envelope. Repeating it on a resource already gone is 404 |
| Malformed body, query or path; failed validation | 400 |
| Missing, invalid, expired or revoked credential | 401 |
| Authenticated but not permitted | 403 |
| Not found — **or exists but belongs to someone else** | 404. Never let the status reveal that another user's resource exists |
| Conflict with current state (duplicate unique key, stale version) | 409 |
| Body larger than `server.body_limit` | 413 (returned by the server before any handler runs) |
| `Idempotency-Key` reused with a different payload | 422 |
| Rate limit exceeded | 429, with `Retry-After` (gap R4) |
| Unexpected failure | 500. Generic message; the detail goes only to the log |
| A dependency is down (`/readyz`) | 503 |

`PUT` replaces a resource and `PATCH` partially updates it. Both are idempotent here, as is
`DELETE`. A `POST` that creates something with side effects accepts an `Idempotency-Key`
(`middleware.RequireIdempotencyKey()` + `Middleware.Idempotency`) so a client can safely retry it.

## Requests
- JSON bodies with camelCase keys, the same casing as responses
- Query parameters are snake_case: `page`, `limit`, `sort_by`, `keep_current`
- Parse, then validate, at the handler: `BodyParser` / `QueryParser` / `ParamsParser`, then
  `validator.Struct`. A parse failure or a validation failure is a 400 before anything reaches
  a use case. Validate path IDs as UUIDs so a malformed one never reaches PostgreSQL
- Validation errors never echo the submitted value. A rejected password must not come back in
  the response body, or in a logged one. Build the validator with `response.NewValidator()` so
  field names match the JSON keys
- **The caller comes from the token, never from the request.** Read it with
  `ctx.Locals(string(constants.ContextKeyUserID))` and `ContextKeySessionID`, which the auth
  middleware sets. A user ID in a body or path is an attacker-chosen value
- Pass `ctx.UserContext()` to use cases, not the `*fiber.Ctx`

## Response envelope
Every response is built by `pkg/response`. Handlers never call `ctx.Status(...).JSON(...)`
themselves, so the envelope exists in exactly one place.

```json
{ "success": true, "message": "...", "data": { } }
```

Errors use the same envelope. `errors` carries details when they help the client fix the request:

```json
{ "success": false, "message": "Validation failed",
  "errors": [ { "field": "email", "tag": "email", "message": "email must be a valid email address" } ] }
```

- Return use-case errors through `response.HandleError(ctx, err)`. It maps a `utils.ClientError`
  to its status and turns anything else into a logged, generic 500
- Clients branch on the **status code**. `message` is for humans and may change. A stable,
  machine-readable error code — or a move to RFC 9457 Problem Details — is gap R3
- All keys are camelCase (`TestEnvelope_EveryKeyIsCamelCase` enforces it). Timestamps are RFC 3339
  in UTC. IDs are strings. Decimal amounts are JSON strings
- A collection is never `null`. An empty one is `[]`

## Collections and pagination
Use `response.Paginated` for any collection that can grow without bound. `data` stays a plain
array, and the counters go in `meta`:

```json
{
  "success": true,
  "message": "...",
  "data": [...],
  "meta": { "page": 1, "limit": 10, "total": 100, "totalPages": 10, "hasNext": true, "hasPrev": false }
}
```

```go
page := response.NewPagination(total, filter.Pagination.Page, filter.Pagination.Limit)
return response.Paginated(ctx, items, page, response.WithMessage(domain.MsgListFetched))
```

- `pagination.ParsePagination(ctx)` parses `page` and `limit`. It **clamps** out-of-range values
  rather than rejecting them: `limit` is capped at 100, and a page past the end returns an empty
  page
- A small collection bounded per owner (a user's own sessions) may be returned whole with
  `response.Success`. Paging it would make it harder to use and protect nothing
- Offset pagination degrades on deep pages and shifts under concurrent inserts. If a collection
  is large or append-heavy, use cursor (keyset) pagination instead. None exists yet
- Sorting is `?sort_by=<field>&order=asc|desc`. Map `sort_by` through an allowlist to a column,
  and never interpolate it into SQL
- Filters are named for the field they filter: `?status=active&role_id=...`

`TestPaginated_ExactShape` and `TestBarList_ReturnsTheDocumentedShape` assert this shape byte
for byte, so changing it here without changing the code fails the build.

## Headers and cross-cutting behaviour
- `Authorization: Bearer <accessToken>`. `/auth/refresh` takes the refresh token the same way
  (F3 adds an httpOnly cookie mode)
- `X-Request-ID` is echoed on every response and attached to every log line of the request. A
  caller's own ID is kept only when `requestid.Valid` accepts it; otherwise a new one is generated
- CORS, security headers, body limits and timeouts are configured in `internal/bootstrap`.
  Handlers don't set them

## Documentation
- Every handler carries swag annotations that describe what it actually returns: the status
  codes, and `data` typed as the real DTO. Run `make swag` and commit `.swagger/` in the same
  change as the handler
