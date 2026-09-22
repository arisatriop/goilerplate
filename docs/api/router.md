# Router & API Routes Guide

Guide to Goilerplate's routing structure. Router is organized into 3 different groups based on audience and authentication strategy.

---

## 🌐 Global Routes

Applies to all requests:

| Method | Path       | Description                                                          |
| ------ | ---------- | -------------------------------------------------------------------- |
| `GET`  | `/`        | Welcome message                                                      |
| `GET`  | `/livez`   | Liveness. Checks nothing — answers whether the process should restart |
| `GET`  | `/readyz`  | Readiness. Pings PostgreSQL and Redis; `503` when one is down         |

`/health` and `/healthcheck` are kept as aliases of `/livez` and `/readyz` respectively.

Point a Kubernetes `livenessProbe` at `/livez` and a `readinessProbe` at `/readyz`. A liveness
probe that checks dependencies turns a database outage into a restart loop across every pod.

`/readyz` reports each dependency as `healthy` or `unhealthy` and nothing more. The endpoint is
unauthenticated, and driver errors carry host names, ports, database names and user names; the
detail is written to the application log instead.

---

## 🔒 Internal Routes (`internal.go`)

**For service-to-service communication within your own infrastructure.**

- **Prefix:** `/internal`
- **Auth:** `InternalAuthenticate()` — validates that caller is a trusted internal service
- **Versioning:** ❌ None — internal contracts are tightly coupled and evolve together
- **Permissions:** ❌ None — full access after authentication

### Use Cases
- Internal service calls (micro-services)
- Admin tools & scripts
- Monitoring & health checks

### Example Routes
```
POST   /internal/users/bulk-import
POST   /internal/cache/clear
GET    /internal/stats
```

---

## 🤝 Partner Routes (`partner.go`)

**For external partner / third-party integrations.**

- **Prefix:** `/partner/v1`
- **Auth:** `PartnerAuthenticate()` — validates partner credentials (API key, OAuth, etc)
- **Versioning:** ✅ Yes (`v1`) — allows evolving partner API without breaking integrations
- **Permissions:** ❌ None — after authentication, partners have full access to scoped routes
- **Registered only when** at least one key is configured under `api_key`; otherwise `/partner/*` returns `404`

### Use Cases
- Third-party integrations (payment processors, shipping providers)
- External SaaS integrations
- Partner APIs

### Example Routes
```
GET    /partner/v1/products
POST   /partner/v1/orders
GET    /partner/v1/shipments/{id}
```

---

## 👥 Public Routes (`public.go`)

**For end-user facing APIs (web/mobile clients).**

- **Prefix:** `/api/v1`
- **Auth:** `Authenticate()` — JWT-based user authentication
- **Versioning:** ✅ Yes (`v1`) — supports API evolution for end users
- **Permissions:** ✅ Yes — each resource endpoint requires specific permission via `RequiredPermission()`, enabling fine-grained RBAC

### Use Cases
- User-facing APIs
- Web/mobile app endpoints
- Public endpoints (authentication-based)

### Example Routes
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
GET    /api/v1/users/profile
PATCH  /api/v1/users/{id}
DELETE /api/v1/users/{id}
```

---

## 📊 Comparison Table

| Aspek          | Internal           | Partner             | Public                      |
| -------------- | ------------------ | ------------------- | --------------------------- |
| **Audience**   | Own services       | External partners   | End users                   |
| **Auth**       | `InternalAuthenticate` | `PartnerAuthenticate` | `Authenticate` (JWT)      |
| **Versioned**  | ❌                 | ✅ (`v1`)           | ✅ (`v1`)                   |
| **Permissions**| ❌                 | ❌                  | ✅ (RBAC per endpoint)      |
| **Prefix**     | `/internal`        | `/partner/v1`       | `/api/v1`                   |

---

## 🔐 Authentication Methods

### Public Routes (JWT)

```go
router.Post("/auth/login", handler.Auth.Login)
router.Get("/users/profile", middleware.Authenticate(), handler.User.GetProfile)
router.Delete("/users/:id",
    middleware.Authenticate(),
    middleware.RequiredPermission("users.delete"),
    handler.User.Delete)
```

**Flow:**
1. Client sends JWT token in `Authorization: Bearer <token>`
2. Middleware validates token
3. Middleware checks permissions
4. Handler executes

### Partner Routes (API Key)

```go
router.Post("/partner/v1/orders",
    middleware.PartnerAuthenticate(),
    handler.Partner.CreateOrder)
```

**Flow:**
1. Client sends API key in header or query param
2. Middleware validates API key
3. Handler executes

### Internal Routes (Service Auth)

```go
router.Post("/internal/users/bulk-import",
    middleware.InternalAuthenticate(),
    handler.Admin.BulkImport)
```

**Flow:**
1. Internal service sends secret/token
2. Middleware validates internal auth
3. Handler executes

---

## 🎯 Route Registration Example

### Public Route with Permission

```go
// File: router/public.go
func registerUserRoutes(router fiber.Router, handler *handler.User, mw *middleware.Middleware) {
    // Create user
    router.Post("/users",
        mw.RequiredPermission("users.create"),
        handler.Create)

    // List users
    router.Get("/users",
        mw.RequiredPermission("users.read"),
        handler.GetList)

    // Get user by ID
    router.Get("/users/:id",
        mw.RequiredPermission("users.read"),
        handler.GetByID)

    // Update user
    router.Patch("/users/:id",
        mw.RequiredPermission("users.update"),
        handler.Update)

    // Delete user
    router.Delete("/users/:id",
        mw.RequiredPermission("users.delete"),
        handler.Delete)
}
```

---

## 🛡️ Rate Limiting

Rate limiting is applied per route group via `RateLimit` middleware. Storage is Redis-backed when Redis is enabled — counters are shared across all instances. Falls back to in-memory when Redis is disabled (dev/local only).

| Scope | Key | Config key |
|---|---|---|
| Auth routes | IP address | `rate_limit.auth` |
| Public routes | Authenticated user ID | `rate_limit.user` |
| Partner routes | API key | `rate_limit.partner` |

```go
// Applied at the group level in router
auth := route.Group("api/v1/auth").Use(r.Wired.Middleware.RateLimit.Auth)
api  := route.Group("api").Use(r.Wired.Middleware.RateLimit.User)
```

---

## 🔁 Idempotency

For sensitive `POST` endpoints (create order, payment, etc.), apply idempotency middleware to prevent duplicate processing on retries.

**How it works:**
1. Client sends `Idempotency-Key: <uuid>` header (at most 255 characters); keys are scoped per user
2. If key is new → process request, store the 2xx response for 24h
3. If key is seen with the same method, URL, and body → return the stored response + `Idempotency-Replayed: true` header
4. If key is seen with a different method, URL, or body → `422`
5. If a request with the same key is still being processed → `409` (the client retries later)
6. Non-2xx responses are not stored, so a failed request can be retried with the same key
7. If key is missing and `RequireIdempotencyKey` is applied → `400`

With Redis enabled, stored responses and in-flight locks are shared by every instance. Without
Redis they are kept in memory, so deduplication only works within one instance (a startup warning
is logged).

Place it **after** authentication and permission checks, so rejected requests are neither stored
nor locked:

```go
// Optional key — skip if not provided
foo.Post("",
    r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionFooCreate),
    r.Wired.Middleware.Idempotency,
    handler.Create)

// Mandatory key — 400 if missing
foo.Post("",
    r.Wired.Middleware.Auth.RequiredPermission(constants.PermissionFooCreate),
    middleware.RequireIdempotencyKey(),
    r.Wired.Middleware.Idempotency,
    handler.Create)
```

> Only `POST` needs idempotency. `PUT` and `DELETE` are naturally idempotent by HTTP spec.

---

## 📝 Best Practices

### ✅ DO

1. **Use appropriate route group** based on audience
2. **Version public & partner APIs** for backward compatibility
3. **Set granular permissions** for public routes (RBAC)
4. **Document auth requirements** in each route
5. **Keep internal routes simple** without versioning
6. **Apply idempotency** on sensitive POST endpoints (payments, orders)

### ❌ DON'T

1. Mix internal & public routes in one group
2. Change auth method of existing route (breaking change)
3. Deploy breaking API changes without new version
4. Skip permission checks in public routes
5. Expose internal routes publicly
6. Apply idempotency on PUT/DELETE — they are already idempotent by HTTP spec

---

## 🔗 Related Docs

- [CRUD Operations](../guides/crud-operations.md) - How to create new endpoints
- [Authentication & Authorization](../guides/architecture.md) - JWT & RBAC details
- [Main README - API Endpoints](../../README.md#-api-endpoints) - Request examples

---

## 🚀 Adding New Routes

### Step 1: Define Handler

```go
// handler/product.go
type Product struct {
    validator *validator.Validate
    usecase   product.Usecase
}

func (h *Product) Create(ctx *fiber.Ctx) error {
    // implementation
}
```

### Step 2: Register Route

```go
// router/public.go
func (r *Router) setupRoutes() {
    // ...
    productHandler := r.handler.Product

    products := r.Public.Group("/products")

    // Sensitive POST — require idempotency key
    products.Post("",
        middleware.RequireIdempotencyKey(),
        r.Wired.Middleware.Idempotency,
        middleware.RequiredPermission("products.create"),
        productHandler.Create)

    products.Get("", middleware.RequiredPermission("products.read"), productHandler.GetList)
    // ...
}
```

### Step 3: Add Permission Constant

```go
// pkg/constants/permission.go
const (
    PermissionProductCreate = "products.create"
    PermissionProductRead   = "products.read"
    PermissionProductUpdate = "products.update"
    PermissionProductDelete = "products.delete"
)
```

---

## 💡 Route Organization

```
router/
├── router.go          # Main router setup, global middleware
├── public.go          # /api/v1 routes (user-facing)
├── partner.go         # /partner/v1 routes (third-party)
└── internal.go        # /internal routes (service-to-service)
```

Each file handles a separate group with its own authentication strategy.

---

## 📦 Response Envelope

Every response goes through `pkg/response`. Handlers do not build the envelope themselves.

```json
{
  "success": true,
  "message": "Success",
  "data": { "...": "..." },
  "meta": {
    "requestId": "01J...",
    "timestamp": "2026-01-01T00:00:00Z"
  }
}
```

On failure `success` is `false`, `data` is omitted, and `errors` carries the detail:

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": [{ "field": "email", "message": "email is required" }]
}
```

### Key casing

**Every key the API emits is camelCase** — `requestId`, `accessToken`, `totalPages`, `hasNext`.

`Meta` used to emit `request_id` while the DTOs around it emitted `accessToken` and
`refreshTokenExpiresAt`, so a single response body carried both conventions and a client had to
know which applied where. `TestEnvelope_EveryKeyIsCamelCase` in `pkg/response` walks the whole
marshalled envelope and fails if a snake_case key reappears at any depth.

Two places keep snake_case deliberately and are not part of this rule:

- **JWT claims** (`pkg/jwt`) — `user_id` and `session_id` follow the JWT convention, and the
  registered claims (`sub`, `exp`, `jti`) are fixed by RFC 7519.
- **Log attributes** — `request_id` is a log field, not a response field.

### Status codes

| Situation | Status |
|---|---|
| Success `GET`/`PUT`/`PATCH` | 200 |
| Success `POST` (created) | 201 |
| Bad input / validation error | 400 |
| No or invalid token | 401 |
| Valid token, no permission | 403 |
| Resource not found | 404 |
| Duplicate / state conflict | 409 |
| Rate limit exceeded | 429 |
| Server or logic error | 500 |
| A dependency is down (`/readyz`) | 503 |
