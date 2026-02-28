# Router

The router package organizes API routes into three distinct groups based on the intended consumer and authentication strategy. All routes share a common set of global middleware (`RequestLogger`, `Recover`) applied in `router.go`.

## Global Routes

| Method | Path      | Description                                         |
| ------ | --------- | --------------------------------------------------- |
| `GET`  | `/`       | Welcome message                                     |
| `GET`  | `/health` | Health check (PostgreSQL, GORM, Redis connectivity) |

---

## Internal Routes (`internal.go`)

**For service-to-service communication within your own infrastructure.**

- **Prefix:** `/internal`
- **Auth:** `InternalAuthenticate()` — validates that the caller is a trusted internal service.
- **Versioning:** None — internal contracts are tightly coupled and evolve together.
- **Permissions:** None — full access is granted once authenticated.

---

## Partner Routes (`partner.go`)

**For external partner / third-party integrations.**

- **Prefix:** `/partner/v1`
- **Auth:** `PartnerAuthenticate()` — validates partner credentials (e.g., API key).
- **Versioning:** Yes (`v1`) — allows evolving the partner API without breaking existing integrations.
- **Permissions:** None — once authenticated, partners have full access to their scoped routes.

---

## Public Routes (`public.go`)

**For end-user facing APIs (web/mobile clients).**

- **Prefix:** `/api/v1`
- **Auth:** `Authenticate()` — JWT-based user authentication.
- **Versioning:** Yes (`v1`) — supports API evolution for end users.
- **Permissions:** Yes — each resource endpoint requires a specific permission via `RequiredPermission()`, enabling fine-grained RBAC.

---

## Summary

| Aspect          | Internal               | Partner                 | Public                      |
| --------------- | ---------------------- | ----------------------- | --------------------------- |
| **Audience**    | Own services           | External partners       | End users                   |
| **Auth**        | `InternalAuthenticate` | `PartnerAuthenticate`   | `Authenticate` (JWT)        |
| **Versioned**   | ❌                     | ✅ (`v1`)               | ✅ (`v1`)                   |
| **Permissions** | ❌                     | ❌                      | ✅ (RBAC per endpoint)      |
| **Prefix**      | `/internal`            | `/partner/v1`           | `/api/v1`                   |
