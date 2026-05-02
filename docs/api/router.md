# Router & API Routes Guide

Guide to Goilerplate's routing structure. Router is organized into 3 different groups based on audience and authentication strategy.

---

## 🌐 Global Routes

Applies to all requests:

| Method | Path      | Description                                         |
| ------ | --------- | --------------------------------------------------- |
| `GET`  | `/`       | Welcome message                                     |
| `GET`  | `/health` | Health check (PostgreSQL, GORM, Redis connectivity) |

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
1. Client sends `Idempotency-Key: <uuid>` header
2. If key is new → process request, cache 2xx response for 24h
3. If key is seen → return cached response + `Idempotency-Replayed: true` header
4. If key is missing and `RequireIdempotencyKey` is applied → return `400`

```go
// Optional key — skip if not provided
foo.Post("", r.Wired.Middleware.Idempotency, handler.Create)

// Mandatory key — 400 if missing
foo.Post("",
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
