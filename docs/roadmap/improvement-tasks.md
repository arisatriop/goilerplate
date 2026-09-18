# Goilerplate Improvement Task List

Improvement plan for the authentication flow and optional-component architecture of this
boilerplate. Based on a review of the current implementation (`user_tokens`, `user_sessions`,
auth middleware, bootstrap, and wiring).

**Sizing:** S = ≤ ½ day · M = 1–2 days · L = 3–5 days
**Status:** in progress — Phases 1–3 done (T1.1–T3.7); Phase 4: all but T4.4 (cleanup job) done

---

## Decisions

| # | Decision | Status |
|---|---|---|
| D1 | Keep server-side sessions (`user_sessions`, one row per login/device). Sessions always exist; how often they are checked is configurable. | ✅ Decided |
| D2 | `auth.revocation` defaults to `strict` (secure by default). Simple projects may opt into `refresh_only`. | ✅ Decided |
| D3 | PostgreSQL only; MySQL support removed (T0.1, T2.1) | ✅ Decided |
| D4 | Fresh baseline: no backward compatibility with existing data, deployments, or other repositories. Migrations, config keys, and API contracts may be rewritten freely. | ✅ Decided |
| D5 | `/internal` stays on the same port as public routes and is intended for pod-to-pod traffic only. Exposure is controlled at the API gateway / Ingress, not by a separate listener. | ✅ Decided |

**Rationale for D1/D2:** revocable login (real logout, per-device logout, password reset
kicking out other devices, admin deactivation) requires server-side state, and nearly every
production project eventually needs at least one of these. Adding sessions to an existing
project later is far harder than relaxing the check via config. With `refresh_only`, sessions
are touched only on login, refresh, and logout, so simple projects pay almost nothing.

| Mode | Session checked | Logout takes effect | Use for |
|---|---|---|---|
| `strict` (default) | Every request (cached if Redis/memory) | Immediately¹ | Sensitive data, most production apps |
| `refresh_only` | Refresh only | Within access token TTL (~15m) | Simple/internal projects, prototypes |

¹ With `session_cache: memory` and multiple instances, other instances may lag by up to
`session_cache_ttl`.

**Rationale for D3:** existing migrations already run only on PostgreSQL, so MySQL support is
effectively broken today. Maintaining two dialects doubles the work of every schema change,
and PostgreSQL features (`TIMESTAMPTZ`, partial indexes, `CHECK`, `RETURNING`) keep the design simpler.

**Rationale for D5:** a second listener, extra Service, and NetworkPolicy add operational
complexity most projects do not need. The trade-off is that isolation depends on gateway
configuration, so the gateway must never forward `/internal`, and an optional shared secret
(T4.1) provides a safety net against misconfiguration.

**Rationale for D4:** the boilerplate is the starting point for future projects, so the cleanest
design wins over migration paths. Existing migration files are edited in place to form a clean
baseline instead of stacking corrective migrations on top.

---

## Context

### Guiding principles

1. **Correct and secure with zero optional components.** Redis, gRPC, OTel, S3, etc. add
   performance or scale — never correctness. No security gap may exist only when Redis is off.
2. **No `if cache.IsEnabled()` in the domain.** Use interfaces with `noop` / `memory` / `redis`
   implementations selected at wiring time.
3. **Fail-fast config validation.** An enabled feature with incomplete config, or placeholder
   secrets in production, stops the app at startup with a clear message.
4. **Minimal defaults.** Example config ships with Redis, gRPC, OTel, and email features off,
   and `local` storage.
5. **Atomic state changes.** Use conditional `UPDATE ... WHERE` and check `RowsAffected`
   (or `RETURNING`) instead of read-then-write.

### Deployment profiles

| | **Minimal** | **Standard** | **Full** |
|---|---|---|---|
| Example | Internal CRUD, MVP | Production app, 1–N instances | Microservice |
| DB | PostgreSQL | PostgreSQL | PostgreSQL |
| Redis | ❌ | ✅ | ✅ |
| gRPC / OTel | ❌ | ❌ | ✅ |
| Storage | local | local / S3 | S3 + CDN |
| Email/OTP, Google login | ❌ | optional | optional |

### Target token design

```
Access token  → Stateless JWT (TTL 10–15m), not stored anywhere.
Refresh token → JWT + jti, rotated on every refresh. user_sessions stores the active refresh_jti.
Session       → one row per login (per device/browser). Every token carries its session_id.
Revocation    → user_sessions.is_active = false (DB is always the source of truth).
                Logout revokes only the caller's session; other devices are untouched.
                Only LogoutAll, password change/reset, and user deactivation revoke all sessions.
Session check → via the SessionStore interface:
                  - none   : one primary-key query per request (Minimal)
                  - memory : in-process cache with a short TTL (Minimal, single instance)
                  - redis  : shared cache (Standard / Full)
Blacklist     → none; session-level revocation makes per-token blacklisting unnecessary.
One-time tokens (email verification, password reset, email change) → one_time_tokens table.
```

Proposed config:

```yaml
auth:
  revocation: strict        # strict | refresh_only
  session_cache: auto       # auto | none | memory | redis  (auto = redis if enabled, else none)
  session_cache_ttl: 30s    # memory: max cross-instance lag; redis: bounds staleness if the DB is changed directly
  session_expiry: 168h      # absolute session lifetime
  remember_me_expiry: 720h  # absolute session lifetime when remember_me = true
  refresh_reuse_grace: 10s  # tolerate concurrent refreshes from multiple tabs
  permission_cache_ttl: 15m # safety net; permission changes invalidate explicitly

jwt:
  access_secret: ...
  refresh_secret: ...
  access_token_expiry: 15m
  issuer: goilerplate
  audience: goilerplate-api
```

- `strict` — session checked on every request; logout takes effect immediately.
- `refresh_only` — no per-request lookup; revocation takes effect when the access token expires.
- `memory` with multiple instances — revocation may lag by up to `session_cache_ttl`; log a
  startup warning.
- Session lifetime is **absolute**: rotation never extends `expires_at`. The user must log in
  again once the session expires, regardless of activity.

### Client contract

- `/auth/refresh` returns a **new refresh token on every call**; clients must always store the
  latest one.
- Refreshing again with the immediately previous refresh token within `auth.refresh_reuse_grace`
  is safe and returns the current token pair (covers concurrent tabs and lost responses).
- Clients should still serialize refresh calls (single-flight) to avoid needless requests.
- Documented in `docs/guides/auth.md` (T6.2).

### Issues found in the current implementation

| Issue | Addressed by |
|---|---|
| **HTTP request logger writes secrets to logs unfiltered**: `Authorization` / `x-api-key` headers, login passwords (request payload), and issued tokens (response body) | T1.7 |
| ~~Client IP taken from `X-Forwarded-For` without a trusted-proxy check (spoofable); behind a load balancer `c.IP()` is the proxy, so per-IP rate limits are shared by all users~~ | T4.6 ✅ |
| ~~No security response headers~~ | T4.6 ✅ |
| ~~No change-password endpoint; no maximum password length (bcrypt rejects > 72 bytes, surfacing as a 500)~~ | T3.6 ✅, T3.7 ✅ |
| Registration reveals whether an email is already registered | T5.1 |
| ~~No security audit events (login failures, lockouts, token reuse, password resets)~~ (reset/email-change flows land with T5.1) | T4.7 ✅ |
| ~~Logout does not revoke access tokens issued by earlier refreshes~~ (fixed by the session check) | T1.2 ✅ |
| ~~Refresh tokens are reusable until expiry; no rotation or reuse detection~~ | T3.3 ✅ |
| ~~`user_sessions.is_active` / `expires_at` never checked; `RememberMe` has no effect~~ | T3.2 ✅, T3.3 ✅ |
| ~~`user_tokens` grows unbounded (row per login and per refresh, no cleanup)~~ (table dropped) | T3.2 ✅ |
| ~~DB `UPDATE used_at` on every authenticated request~~ | T3.2 ✅ |
| ~~`user_tokens` mixes session tokens with one-time tokens; `LogoutAll` would wipe pending reset/OTP tokens~~ (`one_time_tokens` added in T2.4; `user_tokens` dropped in T3.2, and `LogoutAll` no longer touches tokens at all) | T2.4 ✅, T3.2 ✅ |
| ~~Login writes (session, tokens, cache) are not transactional~~ | T3.4 ✅ |
| ~~`LogoutAll` scans every `token:*` / `session:*` key in Redis~~ | T1.2 ✅ |
| ~~Redis-enabled validation reads cache only; expiry check skipped~~ | T1.2 ✅ |
| ~~Logout errors printed with `fmt.Printf` and swallowed~~ (that code is gone) | T3.2 ✅, T3.5 ✅ |
| ~~Redis cache implementation lives in `domain/auth`; domain imports GORM and Fiber~~ | T1.2 ✅, T1.6 ✅ |
| ~~`/internal` (intended for pod-to-pod only) relies solely on gateway path rules; no safety net if the gateway is misconfigured, and the deployment docs do not state the rule~~ | T4.1 ✅ |
| ~~Partner API key compared with `==` (not constant time); raw key stored in context~~ | T4.2 ✅ |
| ~~gRPC server has no auth interceptor; reflection toggled by `app.env`~~ | T4.3 ✅ |
| ~~Idempotency middleware becomes a no-op without Redis~~ | T1.4 ✅ |
| ~~Migrations are PostgreSQL-only although MySQL is a supported driver~~ | T2.1 ✅ |
| ~~Time columns use `TIMESTAMP` without timezone~~ | T2.2 ✅ |
| ~~GORM models use MySQL column types; redundant/unused indexes~~ | T2.3 ✅, T2.4 ✅ |
| ~~Lockout off-by-one (uses pre-increment attempt count)~~ | T3.7 ✅ |
| ~~Login reveals account existence (disabled status returned before password check)~~ | T3.7 ✅ |
| ~~JWT validation uses `ParseUnverified` to pick a secret, with a generic fallback secret; no `aud`/`iss` checks~~ | T3.1 ✅ |
| CORS config exists but the middleware is commented out; 100MB body limit hardcoded | T5.4 |
| S3 driver: no path-style option, public-read only, extra `HeadObject` per upload | T5.3 |
| Permission cache TTL is 7 days and is only refreshed on login/refresh; role or permission changes are not invalidated (TTL fixed; invalidation hooks await role-management endpoints) | T1.2 |
| ~~Idempotency middleware lets concurrent duplicates both execute and does not detect a reused key with a different payload~~ | T1.4 ✅ |
| ~~`/auth/refresh` and `/auth/logout` share the per-IP login rate limit, so users behind one NAT throttle each other~~ | T4.6 ✅ |
| No tests for `domain/auth` or auth middleware | T6.1 |
| ~~Redis timeouts are multiplied by `time.Second` twice (`5s` config → ~158 years), so dial/read/write/pool timeouts never fire~~ | T1.2 ✅ |
| ~~Wrong email or password returns `400`; API conventions require `401`~~ | T3.7 ✅ |

---

## Summary

| Phase | Tasks | Estimate |
|---|---|---|
| 0. Decision | T0.1 | ✅ decided |
| 1. Optional-component foundation | T1.1 – T1.7 | ~8–10 days |
| 2. Schema & database | T2.1 – T2.4 | ✅ done |
| 3. Token security | T3.1 – T3.7 | ✅ done |
| 4. Other auth surfaces | T4.1 – T4.7 | ~6–8 days |
| 5. Optional features | T5.1 – T5.6 | ~10–13 days |
| 6. Tests & documentation | T6.1 – T6.2 | ~3–4 days |

---

## Phase 0 — Decision

### T0.1 Decide database support
- [x] Choose database support → **PostgreSQL only** (see D3)
- [x] Record the decision in the README (done together with T2.1)

**Done when:** the decision is documented in the README.

---

## Phase 1 — Optional-component foundation

### T1.1 Startup config validation · M — ✅ done
- [x] Add `config.Validate()`, called in `bootstrap.Init()` before any connection is opened
- [x] Per-feature rules:
  - `redis.enabled` → `host` required
  - `filesystem.driver=s3` → `bucket` and `region` required
  - `grpc.enabled` → `port` required (and different from `server.port`)
  - `otel.enabled` → `endpoint` required; `filesystem.driver` must be `local | s3 | drive`
- [x] JWT rules:
  - secrets at least 32 bytes
  - no `<...>` placeholders
  - `access_secret` ≠ `refresh_secret`
- [x] Production-only rules (`app.env=production`): reject weak secrets (example markers such as
      `changeme` / `your_`, or fewer than 10 distinct characters) in JWT secrets, API keys, and
      the S3 secret key
- [x] Log a warning when `internal_auth.mode=none` → done in T4.1, via the new `config.Warnings()`
- [x] Collect all errors and report them together; secret values never appear in messages

`jwt.secret_key` was only checked for presence and placeholders; T3.1 removed the key.

**Done when:** invalid config stops the app with a clear list of errors; each rule has a unit test.

### T1.2 Cache abstraction: interfaces + Null Object · L — ✅ done
- [x] Domain interfaces:
  - `SessionStore` (Get / Set / Delete / DeleteByUser) in `domain/auth/cache.go`
  - `PermissionCache` (Get / Set / Invalidate / InvalidateAll) in `domain/auth/cache.go`
  - `LockProvider` (`TryLock(key, ttl)` returning a release func) as `lock.Provider` in `domain/lock`
- [x] Implementations in `infrastructure/cache/`:
  - `noop` — pass-through, no caching (sessions and permissions; no noop lock)
  - `memory` — TTL map with per-user index, concurrency-safe, no new dependency; real in-process lock
  - `redis` — per-user index set `auth:user_sessions:{uid}` and `auth:permission_users` (no SCAN);
    `SET NX` + owner token for locks, released by a compare-and-delete script
- [x] Shared contract tests run against memory and a real Redis (`REDIS_TEST_ADDR`, database 15);
      CI starts a Redis service for them
- [x] Move the Redis implementation out of `domain/auth/cache_service.go` (file deleted)
- [x] Config `auth.session_cache: auto|none|memory|redis` (also selects the permission cache),
      `auth.session_cache_ttl` (default 30s), validated at startup; `redis` requires `redis.enabled`
- [x] Remove every `IsEnabled()` call from `internal/domain/**`
- [x] Add `auth.ErrNotFound` so the domain no longer imports GORM
- [x] Log a startup warning when `memory` is used (revocation lag across instances)
- [x] `PermissionCache` TTL from `auth.permission_cache_ttl` (default 15m, not the session lifetime)
- [ ] Invalidate cached permissions on every change: role permissions, role menus, user roles,
      user permission overrides (per user, or all users for role-level changes)
      → no endpoint changes roles or permissions yet; `PermissionService.InvalidateUserPermissions` /
      `InvalidateAllPermissions` exist and must be called by those endpoints when they are added.
      Login and refresh already invalidate the user's cache.
- [x] Access tokens are validated against the database instead of a Redis-only lookup; the token
      cache and blacklist are removed (they only existed because the cache was trusted)
- [x] `SessionService.GetActive` checks the session through `SessionStore` on every authenticated
      request; cache failures fall back to the database
- [x] Fix Redis client timeouts in `internal/bootstrap/redis.go`: `dial_timeout`, `read_timeout`,
      `write_timeout`, and `pool_timeout` are already `time.Duration`; pass them directly instead of
      `time.Second * time.Duration(...)`

**Done when:** `grep -r "IsEnabled\|gorm" internal/domain` returns nothing, the app works in
all three cache modes, and a revoked permission is denied on the next request.

Verified end to end in `none`, `memory`, and `redis` against an isolated database: permission
revocation is denied once the cache is invalidated, logout rejects older access tokens of the same
session, other devices keep working, and logout-all rejects every device.

### T1.3 Conditional wiring · M — ✅ done
**Depends on:** T1.2
- [x] `grpc.enabled=false` → gRPC server not constructed, services not registered (`App.GrpcServer` is nil)
- [x] `otel.enabled=false` → no OTel handlers attached: no tracer/meter provider, no Fiber, gRPC,
      or GORM instrumentation, and `/metrics` is not registered
- [x] Partner routes registered only when `api_key` is non-empty
- [ ] Background jobs start only when `jobs.*.enabled` → no background jobs exist yet; T4.4 adds
      the first one (`jobs.cleanup.enabled`) and must follow this rule
- [x] Startup log prints a summary of enabled components (`"msg":"components"`)

**Done when:** with the minimal config, no gRPC port is opened and no Redis/OTel connection is attempted.

Verified by running the server with the minimal and a full configuration and inspecting its sockets:
minimal listens only on the HTTP port and connects only to PostgreSQL; full also opens the gRPC
port, connects to Redis, serves `/metrics`, and registers partner routes.

### T1.4 Idempotency: memory fallback and correctness · M — ✅ done
**Depends on:** T1.2
- [x] Use memory storage when Redis is disabled instead of becoming a no-op (`pkg/cache.MemoryStorage`)
- [x] Log a startup warning that memory storage only deduplicates within one instance
      (the app cannot detect how many instances run)
- [x] In-flight lock per key (`LockProvider`): a concurrent request with the same key gets
      `409 Conflict` instead of executing twice
- [x] Store a fingerprint (method + URL with query + body hash) with the key: same key with a different
      payload gets `422 Unprocessable Entity`
- [x] Replay the stored `Content-Type`; keys scoped per user and limited to 255 characters
- [x] Storage or lock failures return `500` without executing the request (fail closed)
- [x] Example route applies idempotency after the permission check, so rejected requests are
      neither stored nor locked

**Done when:** without Redis, a replayed request returns the stored response, two simultaneous
requests with the same key execute once, and a reused key with a different body is rejected.

### T1.5 Minimal example config · S — ✅ done
- [x] `config.example.yaml`: Redis, gRPC, OTel off; storage `local`; email/Google features off
      (email/Google config does not exist yet; T5.1/T5.2 add it disabled by default)
- [x] Remove the `service.xendit/midtrans/doku` block
- [x] Add `config.full.example.yaml` documenting every option (a test fails when a config key is missing from it)
- [x] Development-only JWT secrets in the minimal example, rejected by T1.1 validation in production
- [x] Tests: both examples decode strictly (no unknown keys) and the minimal example passes validation

**Done when:** copying `config.example.yaml` plus a database is enough to run the app.

### T1.6 Framework-free domain · M — ✅ done
- [x] Introduce a `DeviceRequest` struct (user agent, forwarded IPs, accept headers, remote IP);
      `DeviceService` no longer accepts `*fiber.Ctx`
- [x] Move audit context helpers to `pkg/auditctx`, replacing `infrastructure/context`
      (this also removes the `application/register` → `infrastructure` import)
- [x] HTTP (and gRPC) handlers build `DeviceRequest` themselves (`handler.newDeviceRequest`; no gRPC
      handler needs device info yet)

**Done when:** `grep -r "gofiber" internal/domain` returns nothing.

### T1.7 Redact secrets from logs · S — ✅ done
- [x] HTTP request logger: redact `Authorization`, `Cookie`, `Set-Cookie`, `x-api-key`,
      `X-Internal-Secret` headers
- [x] Redact sensitive JSON fields in request and response bodies at any depth
      (`password`, `current_password`, `new_password`, `access_token`, `refresh_token`,
      `token`, `otp`, `secret`, `api_key`), configurable via `log.redact_fields`
- [x] Skip body logging entirely for `/auth/*` routes (`log.omit_body_paths`, default `/api/v1/auth`)
- [x] Redact query parameters with sensitive names (e.g. reset `token`)
- [x] Apply the same rules to the gRPC request logger (request/response payloads; metadata is not logged)
- [x] OTel HTTP spans: strip sensitive query parameters from recorded URLs
- [x] Same rules for outgoing `pkg/httpclient` and `pkg/grpcclient` logs

Implemented in `pkg/redact`. Field names match case-insensitively ignoring `_` / `-`, so
protojson `accessToken` is covered. Unparseable JSON bodies are not logged.

**Done when:** a login → refresh → logout run produces logs without any password, token, or API key
(covered by a test that inspects captured log output).

---

## Phase 2 — Schema & database

Migrations are edited in place to form a clean baseline (D4).

### T2.1 Remove MySQL support · S — ✅ done
**Depends on:** T0.1
- [x] Delete `internal/bootstrap/database/mysql.go` and the MySQL branch in `gorm.go` / `app.go`
- [x] Remove `db.driver` from config (removed; an existing `driver: postgres` line is ignored)
- [x] Remove the `go-sql-driver/mysql` and `gorm.io/driver/mysql` dependencies (`go mod tidy`) — no longer
      direct; they remain indirect because `gorm.io/plugin/opentelemetry` imports the MySQL driver
- [x] Document "PostgreSQL only" in the README, `CLAUDE.md`, and `docs/`
- [x] GORM honors `db.sslmode` (it was hardcoded to `disable`); pgx and GORM share one quoted DSN
      (`PostgresDSN`), so passwords with spaces or quotes work

**Done when:** the project builds without MySQL dependencies and `make migrate-up` succeeds on PostgreSQL.

### T2.2 Migration baseline + UTC · S — ✅ done
- [x] Every time column is `TIMESTAMPTZ`
- [x] Merge `add_auth_fields_to_users` into `create_users_table`
- [x] Rename `001_create_bars_table` to the same timestamp naming as the other migrations
- [x] Primary keys: UUIDv7 generated in the application (`uuid.NewV7()`, already available in
      `google/uuid`) for index locality; no reliance on DB-side `gen_random_uuid()` defaults
- [x] `time.Local = time.UTC` in `cmd/server` and `cmd/migrate`
- [x] GORM `NowFunc: utils.Now`; DSN `TimeZone=UTC` (`timezone=UTC`, unquoted — `gorm.io/driver/postgres`
      rejects a quoted value)
- [x] The migrator's own `migrations.executed_at` is `TIMESTAMPTZ`; the unused `utils.TimeJakarta` is removed
- [x] Migration and CRUD conventions (`.claude/skills`, `/add-domain`, `docs/guides/crud-operations.md`)
      updated: `TIMESTAMPTZ`, no DB-side UUID default

Existing databases must be reset (drop, or migrate down fully) and migrated again, because the
baseline migrations were edited in place (D4).

**Done when:** `make migrate-up` then `make migrate-down` (all steps) runs cleanly on an empty
database, and API timestamps always end in `Z` regardless of server timezone.

Verified on an empty database: up → down ×12 (no tables left) → up again; no `timestamp without
time zone` columns and no `id` defaults; with `TZ=Asia/Jakarta`, every timestamp in the login
response ends in `Z`, and user, session, and token IDs are UUIDv7.

### T2.3 `user_sessions` schema · M — ✅ done
- [x] Columns: `id`, `user_id`, `refresh_jti`, `previous_refresh_jti`, `rotated_at`,
      `device_name`, `device_type`, `device_id`, `ip_address`, `user_agent`, `is_active`,
      `expires_at`, `last_used_at`, `revoked_at`, `revoked_reason`, `created_at`
- [x] No `refresh_token_hash` (replaced by `refresh_jti`)
- [x] `ip_address` as `INET`; empty values are stored as NULL
- [x] Indexes: `(user_id, is_active)` and `expires_at` only; `location` is dropped from the
      table, the entity, and the session response DTO
- [x] Align the GORM model with the schema (no `location`, no MySQL-style tags)
- [x] Refresh tokens carry a `jti` (`jwt.TokenPair.RefreshTokenID`), stored as the session's
      `refresh_jti` at login. Rotation into `previous_refresh_jti` / `rotated_at` is T3.3.
- [x] `DeleteTokensBySession` (which looked the session up by `refresh_token_hash`) is replaced
      by `RevokeSession(userID, sessionID, reason)`: one conditional `UPDATE ... WHERE
      id = ? AND user_id = ? AND is_active` that also writes `revoked_at` / `revoked_reason`.
      Revoking the session is what invalidates its refresh token — the refresh middleware
      already requires an active session. `DeactivateUserSessions` takes a reason the same way.
- [x] Revocation reasons are constants in `domain/auth`: `logout`, `logout_all`,
      `password_change`, `reuse_detected`, `admin`

**Done when:** migrations apply and roll back cleanly and the model matches the schema.

Verified on an empty database: up → down ×13 (only `migrations` left) → up again; the live
schema matches the column list above with exactly the two indexes. Covered by PostgreSQL
integration tests (`POSTGRES_TEST_DSN`): session round-trip including NULL `ip_address`,
revoking one device leaving the user's other sessions active, a second revoke reporting
`ErrNotFound`, another user's ID failing to revoke the session, and `DeactivateUserSessions`
stamping the reason on every active row.

### T2.4 `one_time_tokens` table · M — ✅ done
- [x] Add `create_one_time_tokens_table`. The `user_tokens` migration stayed until T3.2 removed
      the access/refresh token storage that used it; T3.2 dropped that table and its migration.
- [x] `token_type` restricted by `CHECK` to `email_verification | password_reset | email_change`
- [x] Indexes: `token_hash` UNIQUE, `(user_id, token_type)`, `expires_at` — nothing else
- [x] `attempts` column (failed verification attempts, used by OTP limits in T5.1)
- [x] Accurate column/table comments; GORM model without MySQL-style `char(36)` / `datetime(3)` tags
- [x] Repository `ConsumeOneTimeToken(hash, type)`:
      `UPDATE ... SET used_at = now WHERE token_hash = ? AND token_type = ? AND used_at IS NULL AND expires_at > now`,
      then require `RowsAffected == 1`
- [x] Repository also has `CreateOneTimeToken`, `GetLatestActiveOneTimeToken` (newest unused,
      unexpired token per user and type, for OTP verification), and `IncrementOneTimeTokenAttempts`
- [x] `ip_address` is `INET`; empty values are stored as NULL

**Done when:** two concurrent requests with the same token yield exactly one success (covered by a test).

Covered by PostgreSQL integration tests (`POSTGRES_TEST_DSN`, skipped when unset; CI starts a
PostgreSQL service): 20 concurrent consumers of one token yield exactly one success, and reuse,
wrong type, unknown hash, and expired tokens are rejected.

---

## Phase 3 — Token security

### T3.1 JWT hardening · M — ✅ done
- [x] Remove `ValidateToken` (`ParseUnverified` + `secretKey` fallback); use explicit
      `ValidateAccessToken` / `ValidateRefreshToken`. Each offers only the secret of the type
      it expects, so a refresh token cannot authenticate a request even though the middleware
      no longer checks `claims.Type` itself — the jwt package owns that check now.
- [x] Parser options: `WithValidMethods` (HS256 only), `WithIssuer`, `WithAudience`,
      `WithLeeway` (`jwt.leeway`, default 30s), `WithExpirationRequired`
- [x] Key rotation: `kid` header on every token; `jwt.key_id` names the active key and
      `jwt.previous_keys` lists retired ones, accepted for verification only. An unknown or
      missing `kid` is rejected rather than falling back to the active key.
- [ ] Deferred: `jwt.algorithm: HS256 | EdDSA`. Nothing outside this service verifies tokens
      yet, and `SigningMethod` in `pkg/jwt` is the single place to widen when that changes.
- [x] `aud` claim from new `jwt.audience` config
- [x] `jti` on both access and refresh tokens (`TokenPair.AccessTokenID` / `RefreshTokenID`)
- [x] Remove `jwt.secret_key` and `jwt.refresh_token_expiry` from config. The refresh token
      now expires with its session: `auth.session_expiry` (default 168h) and
      `auth.remember_me_expiry` (default 720h) replace the hardcoded `SessionDuration` /
      `rememberMeDuration` constants and are passed to the signer per login.
- [x] Single `pkg/hash.Token()` helper replacing the `hashToken` copies in the auth
      middleware, `TokenStorage`, and `pkg/utils` (the use case's copy went in T2.3)
- [x] `wire` builds one JWT service shared by the use case and the middleware; a second
      instance could have drifted to a different key or issuer

**Done when:** tokens with the wrong issuer, audience, algorithm, or secret are rejected (one test each),
and a token signed with a previous key still validates after rotating the active key.

Covered by `pkg/jwt/jwt_test.go`: wrong issuer, wrong audience, `alg=none` and `alg=HS512`,
wrong secret, unknown and missing `kid`, swapped token types, expired token, and rotation —
a token signed with the previous key validates after the active key moves on, and stops
validating once that key is dropped. Config rules are covered in `config/validate_test.go`.

Verified live against a scratch database and a real server: login issues `kid=v1`, HS256,
`iss=goilerplate`, `aud=goilerplate-api`, `type` and `jti` on both tokens; an access token is
rejected at `/auth/refresh` (401) and a refresh token at `/auth/logout` (401). After restarting
with `key_id: v2` and the old key moved to `previous_keys`, new logins sign with `kid=v2`
while the v1 access and refresh tokens issued before the rotation still work (200).

### T3.2 Stateless access token + session check · L — ✅ done
**Depends on:** T1.2, T2.3, T3.1
- [x] `Authenticate` middleware: verify signature → `SessionStore.Get(claims.SessionID)` →
      require active and unexpired. The access-token lookup alongside it is now gone: nothing
      about an access token is stored, so the session is the only state consulted.
- [x] Cache miss → read DB → populate cache (T1.2)
- [x] Config `auth.revocation: strict | refresh_only`, defaulting to `strict`. Wiring resolves
      it once into `SessionService.EnsureActiveForRequest`, so no request path branches on
      configuration. `/auth/refresh` always calls `GetActive` regardless of the mode — that is
      the point where a revoked login must stop minting access tokens.
- [x] Remove `TokenStorage`, `TokenService`, access-token persistence, and `MarkTokenAsUsed`
- [x] Drop the `user_tokens` table and its migration, the `UserToken` entity and GORM model,
      and the six repository methods that served them
- [x] Put `session_id` in context instead of `token_hash`; `constants.ContextTokenHash` removed
- [x] `Logout` revokes the session directly (`RevokeSession`) instead of deleting token rows;
      `LogoutAll` no longer deletes tokens either

**Done when:**
- in `strict` mode, logout rejects **every** access token of that session immediately, with
  cache `none`, `memory` (single instance), and `redis`
- no DB writes happen per request

Verified live against a scratch database, one run per cache mode. In each, logging in, calling
`/auth/logout`, then replaying **the same** access token returns 401:

| `auth.session_cache` | valid token | logout | same token after logout |
|---|---|---|---|
| `none` | 200 | 200 | **401** |
| `memory` | 200 | 200 | **401** |
| `redis` | 200 | 200 | **401** |

For the write claim, `pg_stat_user_tables` reported `n_tup_upd=12 n_tup_ins=21` both before and
after 30 authenticated requests on one access token — unchanged, where previously every request
issued an `UPDATE used_at`. Unit tests cover both revocation modes, including that
`refresh_only` performs no session lookup per request while refresh still rejects the revoked
session.

### T3.3 Refresh token rotation + reuse detection · L — ✅ done
**Depends on:** T3.2
- [x] Refresh issues a new access token **and** a new refresh token
- [x] Load the session by `claims.SessionID` first: missing, inactive, or expired → plain 401
      (**not** a reuse event — e.g. a refresh after logout). The refresh middleware already does
      this, so such a request never reaches the rotation logic.
- [x] Atomic rotation via `RotateRefreshJTI`:
      `UPDATE user_sessions SET previous_refresh_jti = refresh_jti, refresh_jti = ?, rotated_at = now, last_used_at = now WHERE id = ? AND refresh_jti = ? AND is_active AND expires_at > now`
- [x] `RowsAffected = 0` re-reads the session **from the repository, never the cache**, then:
  - jti equals `previous_refresh_jti` and `rotated_at` is within `auth.refresh_reuse_grace` →
    idempotent success: a new access token plus a refresh token carrying the session's current
    `refresh_jti`, with no further rotation
  - otherwise → reuse detected: revoke only that session (`revoked_reason = reuse_detected`,
    matching the constant and column comment shipped in T2.3) and log a security event
- [x] Refresh JWT `exp` = `session.expires_at` via the new `jwt.SignRefreshToken`, which takes an
      explicit jti and absolute expiry; rotation never extends the session
- [x] Session lifetime from `auth.session_expiry` / `auth.remember_me_expiry` (T3.1)
- [x] Rotation is always on (no config switch)
- [x] New config `auth.refresh_reuse_grace` (default 10s), validated to stay within
      `jwt.access_token_expiry`
- [x] Login and refresh now share `buildMenuAndPermissions`, which they previously duplicated

**Done when:**
- reusing an old refresh token after the grace period revokes that session only
- refreshing after logout returns 401 without a security event
- two concurrent refreshes within the grace period both succeed and end up holding the same
  current refresh token
- a client that lost the refresh response can refresh again with the old token within the grace period
- `remember_me` yields a session that expires after `auth.remember_me_expiry`

Covered by PostgreSQL integration tests (20 concurrent rotations of one token yield exactly one
winner; revoked, expired, and wrong-jti rotations are rejected) and by use-case tests for the
grace decision: within grace returns the session's current token without revoking, past grace
and unknown-jti revoke with `reuse_detected`, and an already-dead session is a plain 401 with no
revocation recorded.

Verified live with `auth.refresh_reuse_grace: 5s`:

| Case | Result |
|---|---|
| Refresh | 200, new refresh token with a different jti |
| Replay the replaced token within grace | 200, returns the session's **current** token |
| Replay it after grace | 401, session revoked `reuse_detected`, security event logged |
| That session's current token afterwards | 401 |
| Another device of the same user | 200, untouched |
| Plain vs `remember_me` refresh `exp` | 168.0h vs 720.0h |
| `exp` after a further rotation | unchanged |

**Defect found and fixed along the way:** `user_sessions.revoked_reason` is documented and
defined as "NULL while active", but the GORM model used a plain `string`, so active rows stored
`''` and a `WHERE revoked_reason IS NULL` filter silently matched nothing. The model field is now
`*string` written through `nullableString`, with a test asserting an active row is found by that
filter. Introduced in T2.3.

### T3.4 Transactional login · S — ✅ done
- [x] Wrap login-info update and session creation in one DB transaction. Token signing touches
      nothing outside the function, so it stays before the transaction opens.
- [x] Write to cache only after commit
- [x] `auth.Repository` gains `WithTx(ctx)`, following the pattern the user and user-role
      repositories already use

**Done when:** a failure midway leaves no orphaned session.

Covered by a PostgreSQL integration test: a transaction that creates the session and then fails
leaves no session row and no `last_login_at` stamp, while the session is visible inside the
transaction before the rollback. A second test pins `WithTx` falling back to the plain
repository when the context carries no transaction.

### T3.5 Logout and LogoutAll · M — ✅ done
**Depends on:** T3.2
- [x] Logout: deactivate the session (`revoked_reason = logout`) and evict it from cache;
      **idempotent** — an already-revoked session is the state the caller asked for, so
      `RevokeSession` returning `ErrNotFound` no longer becomes a 401. The cache is evicted
      either way, so a stale entry cannot outlive the call.
- [x] LogoutAll: `DeactivateUserSessions` (one `UPDATE ... WHERE user_id = ? AND is_active`,
      stamping `logout_all`) then `SessionStore.DeleteByUser` (T3.2)
- [x] Remove `fmt.Printf`, empty `if err {}` blocks, and `//nolint` markers — these lived in
      `TokenService`, deleted in T3.2. No `//nolint` remains anywhere; the surviving
      `fmt.Printf` calls are in `pkg/filesystem` and the migrator CLI, outside the auth path.
- [x] Logout only touches the session identified by `claims.SessionID`; it never queries or
      revokes by `user_id` or `device_id`. `user_id` appears only as a guard in the
      `WHERE id = ? AND user_id = ?` clause, so one user cannot revoke another's session.

**Done when:**
- logout on device A leaves device B's access and refresh tokens fully working
- after LogoutAll every token on every device is rejected
- both are tested in all three cache modes

Verified live, one server run per cache mode:

| `auth.session_cache` | logout A | A access | A refresh | B refresh | logout-all | then access | then refresh | other device |
|---|---|---|---|---|---|---|---|---|
| `none` | 200 | 401 | 401 | **200** | 200 | 401 | 401 | 401 |
| `memory` | 200 | 401 | 401 | **200** | 200 | 401 | 401 | 401 |
| `redis` | 200 | 401 | 401 | **200** | 200 | 401 | 401 | 401 |

`revoked_reason` afterwards: 4 rows `logout`, 8 rows `logout_all`, and no active row carrying an
empty reason.

**Noted while testing (not fixed here):** `/auth/refresh` and `/auth/logout` share the per-IP
`rate_limit.auth` bucket (10/min), so a burst of these calls returns 429 before auth is even
consulted. Already tracked for T4.6; the runs above raised the limit to isolate the behaviour
under test.

### T3.6 Revoke sessions on user state changes · S — ✅ done
**Depends on:** T3.5
- [x] User deactivated → all sessions revoked. `DeactivateUser` flips `is_active` and revokes
      every session (`revoked_reason = admin`) **in one transaction**, then evicts the user from
      the session cache. Doing only the first would leave API access alive until the sessions
      expired, because the per-request check reads the session, not the account.
- [x] `PUT /users/me/password`: requires the current password; revokes every other session and
      keeps the current one. Both the user ID and the kept session ID come from the middleware,
      so a caller can only change their own password and only keep the session they are using.
      The new password goes through `pkg/password` (T3.7), and the update clears any lockout —
      the person proved they own the account, so making them wait out someone else's failed
      guesses would punish the wrong party.
- [ ] Password reset → LogoutAll including the current session — belongs to T5.1, which owns the
      reset flow. `RevokeOtherUserSessions(userID, "", reason)` already revokes all of them.
- [x] Refresh keeps rejecting inactive users (`ValidateUserForRefresh`)

**Done when:** a deactivated user loses API access immediately, and changing the password
logs out every other device.

Verified live. Password change on device A, with B and C signed in:

| Case | Result |
|---|---|
| Wrong current password | `401 "Invalid credential"`, nothing revoked |
| New password too short / 73 bytes | `400` / `400 "at most 72 bytes"`, nothing revoked |
| Valid change | `200` |
| A (made the change) | refresh `200` — stays signed in |
| B and C | refresh `401`, access token `401` |
| Login with the old password | `401` |
| Login with the new password | `200` |

`revoked_reason` afterwards: 4 rows `password_change`, other users untouched. Repository tests
cover that another user's sessions are never included, and that an empty `keepSessionID` revokes
all of them. Domain tests cover the deactivation transaction and the three refusal paths.

**Caveat, verified rather than assumed:** "immediately" holds when deactivation goes through
`DeactivateUser`. Flipping `users.is_active` directly in the database bypasses the revocation:
`/auth/refresh` does reject the user (403, confirmed), but an already-issued **access token keeps
working until it expires** — measured at 200 on an authenticated route after a direct DB update.
This is the same class of caveat as `auth.session_cache_ttl` ("bounds staleness if the DB is
changed directly"). No admin endpoint calls `DeactivateUser` yet; that waits on the
user-management endpoints, alongside the permission-invalidation hooks noted in T1.2.

### T3.7 Atomic lockout + anti-enumeration · M — ✅ done (one bullet deferred)
- [x] Single statement via `RegisterFailedLogin`, which returns whether the account is now locked:
      `UPDATE users SET failed_login_attempts = failed_login_attempts + 1, locked_until = CASE WHEN failed_login_attempts + 1 >= ? THEN ? ELSE locked_until END WHERE id = ? RETURNING ...`
      The `CASE` compares the **post-increment** value, fixing the off-by-one that locked on N+1.
- [x] Check order:
  1. Unknown user → dummy bcrypt comparison → "invalid credentials"
  2. Locked → rejected without evaluating the password, with a dummy bcrypt for equal timing
  3. Wrong password → atomic increment → "invalid credentials"
  4. Disabled → "account disabled", only after the correct password
- [x] Locked response message is generic: `MsgAccountLocked` is now
      "Too many failed attempts, please try again later"
- [x] Invalid credentials return `401`; locked returns `401` too, so neither can be told from the
      other by status. `400` is left to request validation.
- [x] `auth.lockout.max_attempts` (default 5) and `auth.lockout.duration` (default 10m)
- [x] Password policy in the new `pkg/password`, applied at registration: minimum 8 characters,
      maximum 72 **bytes** (bcrypt's truncation point, checked in bytes so multi-byte characters
      count correctly), returned as 400. No composition rules, per NIST 800-63B.
- [ ] **Deferred by decision:** no common-password list ships with the boilerplate. The
      `password.CommonChecker` interface and the policy hook are in place, defaulting to
      `NoCommonList`, so a deployment supplies its own list; wiring points at
      `wire/application.go`.

**Done when:**
- the account locks exactly on attempt N
- while locked, the correct password gets the same response as a wrong one
- the response for an unregistered email is identical to a wrong password (both `401`)
- a concurrency test covers the increment

PostgreSQL integration tests: attempts 1–4 do not lock and the 5th does, with the counter and
`locked_until` checked after each; 20 concurrent failures are all counted exactly once (counter
reaches 20, and every attempt from the 5th on reports the account locked). Domain tests cover the
check order, including that a locked account never counts or evaluates the password.

Verified live with `max_attempts: 5`:

| Case | Result |
|---|---|
| Unknown email | `401 "Invalid credential"` |
| Wrong password | `401 "Invalid credential"` — byte-identical to the above |
| Attempts 1–5 | all `401 "Invalid credential"`; DB shows `attempts=5, locked=true` after the 5th |
| Wrong password while locked | `401 "Too many failed attempts, please try again later"` |
| **Correct** password while locked | identical response |
| Counter during locked attempts | stays at 5 — the password is never evaluated |
| Register with 5-char / 73-byte / valid password | `400` / `400 "at most 72 bytes"` / `201` |

Response-time medians over 8 requests each: unknown email 52.6 ms, existing account 53.3 ms
(ratio 0.99), so the dummy comparison keeps timing from revealing which emails are registered.

---

## Phase 4 — Other auth surfaces

### T4.1 Gateway-controlled `/internal` routes · S — ✅ done
**Depends on:** T1.1

`/internal` stays on the public port and is reachable only through in-cluster traffic (D5).

- [x] Gateway / Ingress rule documented and enforced: forward only an explicit allowlist
      (`/api`, `/partner`, `/health`); never a catch-all `/`. Where the gateway supports it, add
      an explicit deny for `/internal`
- [x] Optional safety net: `internal_auth.mode: none | shared_secret` (default `none`);
      `X-Internal-Secret` compared in constant time. Recommended when the gateway is managed by
      another team or its config changes often
- [x] Startup warning in production when `internal_auth.mode=none` (moved from T1.1)
- [x] Optional caller identity: record `X-Service-Name` in context and logs instead of a fixed
      `system` user
- [x] `docs/deployment/kubernetes.md`: state the rule and a verification command
      (`curl -i https://<public-host>/internal/...` must not reach the app)

The default stays `none`. Requiring a secret out of the box would break every in-cluster caller
that was never given one, and D5 already puts the gateway in front. What was missing was saying
so: `config.Warnings()` is a new non-fatal advisory channel, logged at startup, whose first entry
is exactly this. Errors stop the app; warnings state an assumption that would otherwise go
unexamined.

The comparison reuses `hash.Equal`, so it is over two digests and does not reveal the secret's
length. The secret is reduced to its digest at construction, and `X-Internal-Secret` was already
in `redact.DefaultHeaders`.

`X-Service-Name` replaces the hardcoded `system` caller. It is an **unverified claim** — in
`shared_secret` mode every caller holds the same secret, so nothing distinguishes one from
another — and it is for attribution in logs, never authorization. Because it lands in every log
line the request produces, it is stripped to `[A-Za-z0-9._-]` and capped at 40 characters so a
header cannot pad the logs.

**Done when:**
- a request to `/internal/*` through the public gateway does not reach the app — enforced by the
  Ingress allowlist now documented in `docs/deployment/kubernetes.md`, with a verification
  command and the `X-Request-Id` tell (only the app emits it, so an ingress 404 is
  distinguishable from an app 404)
- with `shared_secret` enabled, requests without the secret get 401 ✅

Verified live:

| Request | `mode: none` | `mode: shared_secret` |
|---|---|---|
| `GET /internal/bars`, no header | 200 | **401** |
| with the correct `X-Internal-Secret` | 200 | 200 |
| wrong secret | — | 401 |
| secret + one extra character | — | 401 |
| secret minus the last character | — | 401 |
| `/health` | 200 | 200 |
| `/api/v1/auth/login` | 400 (no creds) | 400 |

In production with `mode: none`, startup logs:

```json
{"level":"WARN","msg":"configuration","warning":"internal_auth.mode=none: /internal is protected
only by the gateway's routing rules. Set internal_auth.mode=shared_secret for a second lock if
that config is not yours to control."}
```

With `shared_secret` configured: zero warnings. Startup refuses bad config outright —
`internal_auth.mode must be none or shared_secret, got "mtls"`, `internal_auth.secret is
required`, `internal_auth.secret must be at least 32 bytes, got 5` — and no message contains the
value.

Caller attribution, read back from `application-log`:

| `X-Service-Name` sent | recorded `user_id` |
|---|---|
| *(none)* | `system` |
| `billing-worker` | `billing-worker` |
| `bad"name{}` | `badname` |
| 80 × `z` | 40 × `z` |

The secret appeared 0 times in the server log; the header logs as `[REDACTED]`.

### T4.2 Harden partner API keys · S — ✅ done
- [x] `subtle.ConstantTimeCompare`, over two SHA-256 digests rather than the raw strings.
      Comparing the raw values returns early on a length mismatch, which reveals how long the
      real key is. The lookup also visits every configured partner instead of stopping at the
      first match, so the time taken does not depend on which partner presented the key.
- [x] Store keys in config as SHA-256 hashes: a value may be written `sha256:<64 hex chars>`
      instead of the key itself. The server only ever compares digests, so both forms behave
      identically and a config file need never hold a usable credential.
- [x] Put the partner name in context, never the raw key

New `pkg/apikey` holds the `Registry`. Entries are a slice, not a map, because the lookup
deliberately visits all of them. Only digests are kept, so the plaintext keys are reduced at
startup rather than held in memory for the life of the process.

Config validation rejects a malformed digest at startup — a typo would otherwise become a partner
that can simply never authenticate. A digest is checked for shape only: the production entropy
heuristics cannot see through SHA-256 (a hashed `changeme` looks exactly as strong as a hashed
random key), so running them on a digest would claim a confidence the value has not earned. The
plaintext form is still checked as before.

**Done when:** keys never appear in context or logs. ✅

Verified live, against the previous implementation for comparison. Two partners were configured,
`acme` with a plaintext key and `globex` with `sha256:...`:

| Presented | Before (HEAD) | After |
|---|---|---|
| `acme`'s key | 200 | 200 |
| `globex`'s key (configured as a digest) | **401** — the form did not exist | 200 |
| correct key + one extra character | 401 | 401 |
| correct key minus the last character | 401 | 401 |
| the digest itself, used as a key | 401 | 401 |
| no header | 401 | 401 |

Forcing one domain log on a partner route, the `application-log` entry read:

```
before:  "user_id": "partner-key-plaintext-7f3a9c2e5b1d"     ← a live credential in the log
after:   "user_id": "acme"
```

`logger.baseAttrs` puts `user_id` on every application log entry, so under the old code every log
line a partner request produced carried its API key. The `x-api-key` *header* was already redacted
by T1.7 — the leak was the context value, which redaction never saw. Grepping the new server's
whole log for either raw key: 0 occurrences.

Note that a digest is not a credential: presenting `globex`'s digest as the key is rejected, which
is the property that makes storing it in config worth anything.

Tests: name resolution, rejection of a key that is one character too long or too short, both
configured forms authenticating the same partner, a digest pasted in upper case, a malformed
digest matching nothing, an empty registry rejecting everyone including a request with no header,
the registry holding only digests, and the middleware putting the name — not the key — into both
the Go context and Fiber locals.

**Not addressed here:** partner routes apply the rate limiter *after* authentication, so guesses
against the API key are not rate limited at all. Keying the limiter by API key would not help, as
an attacker gets a fresh bucket per guess; this needs a per-IP limit on partner routes.

### T4.3 gRPC auth interceptor · M — ✅ done
**Depends on:** T1.3, T3.2
- [x] Config `grpc.auth.mode: token | shared_secret | none` (default `token`)
  - [x] `token`: unary **and stream** interceptors read `authorization` from metadata and reuse
        the same validator and `SessionStore` as HTTP, so handlers get the same user context
  - [x] `shared_secret`: service-to-service calls without user context, via the
        `x-internal-secret` metadata key and the same constant-time digest compare as T4.1/T4.2
  - [x] `none`: acceptable only when the gRPC port is reachable in-cluster only (same reasoning
        as D5), and logged as a startup warning in production
- [x] Config allowlist for public methods (e.g. health) — full method names, or a whole service
      with a trailing `/*`
- [x] Reflection controlled by `grpc.reflection` (default off), not by `app.env`
- [x] `grpc.tls` for deployments where the port leaves the cluster

**The gRPC port had no authentication at all.** Anything that could open a connection could call
every method.

**Structural change:** the gRPC server is now built in the wire layer instead of
`bootstrap.Init()`. Interceptors can only be handed to `grpc.NewServer`, and the auth interceptor
needs the token validator and session store, which do not exist until infrastructure is wired.
`app.GrpcServer` is assigned there; `app.Config.GRPC.Enabled` replaces `app.GrpcServer != nil` as
the "is gRPC on" test during wiring.

`token` mode reuses `ValidateAccessToken` and `EnsureActiveForRequest` rather than reimplementing
either. Two auth paths that disagree about what a token means is how a revoked session keeps
working on one of them.

**Streams are covered too.** An interceptor that only handled unary calls would leave every
streaming method open while looking like it protected the server. A stream is checked once, when
it opens — there is no later point at which the client can be asked again, so a long-lived stream
outlives a revocation until it closes.

**Reflection is off by default in every environment**, not just production. It publishes the whole
service surface to anyone who can reach the port, and "not production" is not the same question as
"safe to enumerate". This matches what `CLAUDE.md` already claimed the behaviour was.

`X-Service-Name` sanitisation is now shared with HTTP as `utils.ServiceName`: the gRPC request
logger had the same unbounded, unsanitised header that T4.1 fixed on the HTTP side.

**Done when:** in `token` mode, gRPC calls without a valid token get `codes.Unauthenticated`,
and allowlisted methods still work. ✅

Verified live against a running server, with a client that imports the proto module.

| `grpc.auth.mode: token` (the default) | Result |
|---|---|
| no metadata | `Unauthenticated` |
| valid access token | **OK** |
| token without the `Bearer` scheme | `Unauthenticated` |
| tampered token | `Unauthenticated` |
| `Basic` instead of `Bearer` | `Unauthenticated` |

Cross-transport revocation, which is the point of sharing the session store:

```
gRPC before logout                 -> OK
POST /auth/logout (over HTTP)      -> 200
gRPC with the same token           -> Unauthenticated
HTTP with the same token           -> 401
```

| `grpc.auth.mode: shared_secret` | Result |
|---|---|
| no metadata | `Unauthenticated` |
| correct secret | **OK** |
| wrong secret | `Unauthenticated` |
| secret + one extra character | `Unauthenticated` |
| secret minus the last character | `Unauthenticated` |
| a bearer token instead | `Unauthenticated` |

With `public_methods: ["/hello.v1.HelloService/SayHello"]`, that method answered **OK** with no
metadata at all while the secret was still required everywhere else.

Reflection, both with `app.env: local` — the setting the old code keyed off:

```
grpc.reflection unset  ->  "server does not support the reflection API"
grpc.reflection: true  ->  bar.v1.BarService, foo.v1.FooService, hello.v1.HelloService, …
```

TLS, with a generated certificate:

| Client | Result |
|---|---|
| plaintext against the TLS port | `Unavailable` |
| TLS, trusting the certificate | **OK** |
| TLS, not trusting it | `x509: certificate signed by unknown authority` |

Startup validation refuses bad config, and no message contains a secret:

```
grpc.auth.mode must be token, shared_secret, or none, got "mtls"
grpc.auth.secret is required
grpc.auth.public_methods[0] "SayHello" must be a full method name like /pkg.Service/Method or /pkg.Service/*
grpc.tls.cert_file is required
grpc.tls.key_file is required
```

In production with `mode: none`, startup warns rather than refusing, alongside the `internal_auth`
warning from T4.1. A named certificate that cannot be read stops startup instead of quietly
serving plaintext on a port the operator believes is encrypted.

Tests: token mode accepting a valid token and setting the same context keys as HTTP; eight
rejection cases; a case-insensitive scheme; a **revoked session** rejected; a **refresh token**
rejected; shared-secret accept and five rejects; `none` checking nothing; the allowlist matching
exactly, by wildcard, and *not* matching a sibling method, a different service, or a service that
merely shares a prefix; streams rejected without a token and carrying the authenticated context
when accepted; reflection registered only when configured; an unreadable certificate failing.

### T4.4 Cleanup job · M
**Depends on:** T1.2, T2.3, T2.4
- [ ] Delete expired or used `one_time_tokens` older than the retention period
- [ ] Delete inactive or expired `user_sessions` older than the retention period
- [ ] Batched deletes
- [ ] Config `jobs.cleanup.enabled`, `interval`, `retention`
- [ ] Guard with `LockProvider.TryLock` (Redis: `SET NX`; memory/noop: always acquires)
- [ ] Stop gracefully on shutdown via context

**Done when:** the job runs in the minimal profile, and with Redis only one instance executes it.

### T4.5 Auth code cleanup · M — ✅ done
**Depends on:** Phase 3
- [x] Remove duplicated `extractBearerToken` / validation between middleware and `TokenService`.
      `TokenService` was already gone (T1.2), but the two remaining callers each checked for an
      empty header and then called a parser that checked again. Now one `bearerToken(ctx)` reads
      and parses, and every failure answers identically — telling a caller whether the header was
      missing, not Bearer, or empty after the scheme describes our parser, and the one thing it
      reliably tells an attacker is which guess got further.
- [x] Extract the shared "menus + permissions" builder used by Login and Refresh — done in T3.x
      as `buildMenuAndPermissions`, but it still **reimplemented permission resolution**: the
      same three queries and its own byte-identical copy of `mergePermissions`. See below.
- [x] Remove dead code (list below)
- [x] All logging through the structured logger

**The find worth naming.** `authUseCase.mergePermissions` and `PermissionService.mergePermissions`
were character-for-character identical, and `buildMenuAndPermissions` duplicated the whole
roles → role permissions → overrides → merge sequence that `GetUserFinalPermissions` already did.
So the permission list handed to the client at login and the list `RequiredPermission` enforces on
every request afterwards were produced by two separate implementations reading the same tables. A
change to grant/revoke semantics in one would have left the token saying a user has a permission
the enforcement denies, or the reverse. `buildMenuAndPermissions` now calls the same resolver the
per-request check uses, and the second copy is gone.

**Removed as dead** (each verified to have no caller):

| Removed | Note |
|---|---|
| `User.Password` | plaintext password field on the entity, never set or read |
| `User.RememberMe` | never set or read; the live one is on `LoginCredentials` |
| `User.IsAdmin()` | returned a hardcoded `false` with a TODO |
| `User.ShouldLockAccount()` | lockout is decided inside `RegisterFailedLogin`'s `UPDATE` (T3.7), so a read-then-compare helper is the race that change removed |
| `auth.Login` struct | nothing constructed it |
| `TokenType*` constants | a second definition of the same three strings as `OneTimeToken*`, which is the live set matching the DB `CHECK` |
| commented-out `TokenPair`, `Login`, `ChangePassword`, `ForgotPassword`, `ResetPassword`, `RefreshToken`, `AuthResponse` | ~50 lines of commented-out entities |
| `Repository.DeleteUserSessions` + impl | sessions are deactivated, never deleted, to keep the audit trail |
| `Repository.LockUser` + impl | superseded by `RegisterFailedLogin` |
| `MaxFailedLoginAttempts`, `AccountLockDuration`, `AccessTokenExpiry`, `RefreshTokenExpiry`, `VerificationTokenExpiry`, `ResetTokenExpiry` | every one is now a config key; leaving hardcoded twins invites reading the wrong value |
| `TokenManager.TokenInfo()` | its only caller was a commented-out log line, removed with it |

**Kept deliberately**, despite having no caller yet: the `one_time_tokens` repository methods and
`PermissionService.InvalidateAllPermissions`. Both are tested foundations waiting on T5.1 and
role-management endpoints respectively, not leftovers.

**Logging:** `domain/auth`, the middleware, the handlers, the repository and `config` contain no
`fmt.Print`/`log.Print` at all. Two genuine ones remained in `pkg/filesystem` (Google Drive),
writing `Warning: ...` to stdout with no level. They now go through `slog` — not `pkg/logger`,
because that package imports `config`, which imports `pkg/filesystem`.

**Done when:** `golangci-lint` is clean and no unused functions remain. ✅

`unused` is now enabled in `.golangci.yml`, so "no unused functions" is enforced rather than
being a sweep that decays. It only sees unexported symbols, so exported dead code still needs a
human — which is how everything in the table above was found.

Net: **233 deletions against 37 insertions**, `golangci-lint` 0 issues, all 20 packages passing.

Verified live that consolidating permission resolution changed no behaviour. A user with role
`editor` granted `bar.list`, `bar.create`, `bar.delete`, plus a user-level **revoke** of
`bar.delete` and a user-level **grant** of `foo.list`:

| | Result |
|---|---|
| permissions in the login response | `[bar.create, bar.list, foo.list]` |
| permissions in the refresh response | `[bar.create, bar.list, foo.list]` |
| `GET /api/v1/bars` (`bar.list`, granted) | 200 |
| `DELETE /api/v1/bars/:id` (`bar.delete`, revoked by override) | 403 |

So the override semantics survive, and the list the client is handed agrees with what the
enforcement path allows — which is the property the two implementations put at risk.

### T4.6 HTTP hardening: client IP and security headers · S — ✅ done
- [x] Config `server.trusted_proxies` (IPs or CIDRs, validated at startup) and
      `server.proxy_header` (default `X-Forwarded-For`), with Fiber `EnableTrustedProxyCheck`.
      Empty means trust nobody, so the default is safe: the client IP is always the peer that
      connected.
- [x] `DeviceService` takes the already-resolved `ClientIP` and no longer parses
      `X-Forwarded-For` / `X-Real-IP` itself. Deciding whether a forwarding header may be
      believed needs to know which hop the request arrived from, which only the transport does.
- [x] Security headers via Fiber `helmet` (already part of fiber v2 — no new dependency):
      `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`,
      `Referrer-Policy: strict-origin-when-cross-origin`, and `X-XSS-Protection: 0` because the
      legacy filter is itself a vulnerability. `Strict-Transport-Security` when `server.hsts`
      is on — and only on HTTPS, since telling a browser to refuse plain HTTP from a service
      only reachable over plain HTTP would lock everyone out.
- [x] Partner rate-limit key is `hash.Token(apiKey)`, so the raw key never becomes a storage key
- [x] The per-IP auth limiter now covers only unauthenticated endpoints (login, register).
      `/auth/refresh`, `/auth/logout`, and `/auth/logout-all` use a new per-**session** limiter
      that runs after the token is verified.

**Done when:** a spoofed `X-Forwarded-For` from an untrusted source does not change the client IP,
and security headers are present on every response.

Verified live. With no trusted proxies configured, three logins — one plain, two sending
`X-Forwarded-For: 203.0.113.7` and `198.51.100.42` plus matching `X-Real-IP` — all recorded
`127.0.0.1` in `user_sessions`. After adding `127.0.0.1` to `trusted_proxies`, a login carrying
`X-Forwarded-For: 203.0.113.99` recorded that address instead. Response headers on every request:

```
X-Xss-Protection: 0
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: strict-origin-when-cross-origin
```

Rate-limit separation, with `rate_limit.auth.max: 10`: after 14 logins (`200`×9 then `429`×5),
`/auth/refresh` and `/auth/logout` both still returned `200` — the shared-bucket problem that
interfered with the T3.5 and T3.7 test runs is gone.

Tests: a spoofed header from an untrusted peer leaves `ClientIP` unchanged while a trusted peer's
is honoured; HSTS is absent when off, absent over plain HTTP, present on HTTPS, and **absent when
an untrusted peer claims `X-Forwarded-Proto: https`**; two sessions from one IP get separate
budgets; a recording `fiber.Storage` confirms the raw API key never appears as a storage key;
malformed `trusted_proxies` entries fail startup validation.

### T4.7 Security event logging · S — ✅ done
**Depends on:** T1.7
- [x] Structured `security` log events with `user_id`, `session_id`, IP, and user agent (never secrets):
  - [x] login success / failure — failure carries a `reason`: `unknown_email`, `bad_password`,
        `account_locked`, `account_disabled`
  - [x] account locked
  - [x] refresh token reuse detected
  - [x] logout-all and session revoked
  - [x] password changed
  - [x] account deactivated (not in the original list; `DeactivateUser` revokes sessions, so it
        belongs in the same trail)
  - [ ] password reset requested / completed, email changed — the flows do not exist yet.
        `ActionPasswordResetRequested`, `ActionPasswordResetCompleted` and `ActionEmailChanged`
        are defined so T5.1 adopts this vocabulary instead of inventing a second one.
- [x] Distinct log label so events can be routed to a SIEM or alerting: `label: "security-log"`,
      separate from `application-log` and `incoming-request-log`

`logger.Security(ctx, logger.SecurityEvent{...})` writes one entry. Actions, outcomes, and
reasons are exported constants: alerting rules are written against these strings, so they are an
API, not prose.

Identity comes from the context (`user_id`, `session_id`, `request_id`, `client_ip`,
`user_agent`) and may be overridden per event. Login is why the override exists — it establishes
an identity the context does not have yet, and a failed login has no session at all. The request
logger now also puts `client_ip` and `user_agent` on the context, so the auth domain can name the
caller without taking a `*fiber.Ctx`; the IP is the one `c.IP()` resolved against the
trusted-proxy list from T4.6, not a header the client chose.

**What is deliberately not logged:** the attempted email on an unknown-email failure, and the
replayed `jti` on reuse detection. There is no account to name in the first case, and the second
is part of a bearer token while the audit trail outlives it by far. Every `reason` is a fixed
vocabulary word, never anything derived from a credential.

Failures log at `WARN` and successes at `INFO`, so the events an operator most wants still appear
at the default level.

**Done when:** each event above emits exactly one structured log entry (covered by tests). ✅

Verified live against a running server. Every event fired exactly once, in order:

| # | Request | Event | Level | Reason |
|---|---|---|---|---|
| 1 | login, unknown email | `login_failed` | WARN | `unknown_email` |
| 2–4 | login ×3, wrong password | `login_failed` ×3 | WARN | `bad_password` |
| 4 | (3rd failure hits the limit) | `account_locked` | WARN | `bad_password` |
| 5 | login, correct password while locked | `login_failed` | WARN | `account_locked` |
| 6 | login, correct password | `login_succeeded` | INFO | — |
| 7 | `POST /auth/logout` | `session_revoked` | INFO | `logout` |
| 8 | `POST /auth/logout-all` | `all_sessions_revoked` | INFO | `logout_all` |
| 9 | `PUT /users/me/password` | `password_changed` | INFO | — |
| 10 | replay a superseded refresh token past the grace window | `refresh_token_reuse_detected` | WARN | `reuse_detected` |

`user_id` was populated on all 19 events except the unknown-email one, which by definition has no
subject. Grepping the 19 events for JWTs, either password, and the attempted email address found
nothing.

Tests: one event per rejection reason and no more; locking emits its own event separate from the
failure that caused it; a failure that does *not* lock emits no lock event; validation stays
silent on success; explicit identity beats context identity; failures are WARN; a call with no
HTTP context still records. In the middleware: the client IP and user agent reach the context
before any handler runs, and a spoofed `X-Forwarded-For` does not change the recorded IP.

---

## Phase 5 — Optional features

### T5.1 Email module: verification, forgot/reset password, email change · L
**Depends on:** T2.4, T3.6
- [ ] `EmailSender` interface with `log` (default, prints to log) and `smtp` drivers
- [ ] Config `auth.email.enabled`, `email.driver`, `email.smtp.*`, `frontend.base_url`
- [ ] Flows:
  - email verification via 6-digit OTP
  - forgot password → emailed reset link → reset password (revokes all sessions via T3.6)
  - email change: re-authenticate with current password → OTP to the **new** address →
    confirm; the email is swapped only after confirmation
- [ ] Invalidate previous tokens of the same type before issuing a new one
- [ ] Store OTP hashes as HMAC-SHA256 with a server secret (a plain SHA-256 of a 6-digit code
      is trivially brute-forced if the database leaks); high-entropy reset tokens may use SHA-256
- [ ] OTP verification looks up the latest active token by `(user_id, token_type)`, compares the
      HMAC in constant time, and increments `attempts` on mismatch; the token is invalidated after
      `auth.otp.max_attempts` (default 5). Looking up by hash alone cannot count wrong guesses.
- [ ] Registration without enumeration: when the email is already registered, return the same
      success response and send a "you already have an account" email instead. Without the email
      module this is not possible; rely on the per-IP limit and document the trade-off.
- [ ] All token consumption via atomic `ConsumeToken` (T2.4)
- [ ] `ForgotPassword` always returns 200 (no enumeration)
- [ ] Config `auth.require_email_verification` (default `false`); when on, login is rejected
      until verified, checked after the password
- [ ] Routes registered only when `auth.email.enabled`

**Done when:** all flows work end to end with the `log` driver and no SMTP server.

### T5.2 Google login · M
- [ ] Config `auth.google.enabled`, `google.client_id`
- [ ] Verify Google ID tokens; login creates or links the user
- [ ] Link / unlink endpoints for authenticated users; linking requires the Google email to
      match the account email
- [ ] `google_id` column on `users` (nullable, unique)

**Done when:** Google routes do not exist when the feature is disabled.

### T5.3 S3 storage improvements · M
- [ ] `use_path_style` for providers that need path-style addressing (MinIO, Cloudflare R2)
- [ ] Correct bucket parameter per provider (bucket-in-domain endpoints vs path style)
- [ ] `cdn_base_url`: serve files through a CDN; keep the bucket private and set
      `Cache-Control: public, max-age=31536000, immutable`
- [ ] Presigned GET URLs (`presigned_url_expiry`, default 1h) when no CDN is configured
- [ ] Pass the multipart file size to avoid an extra `HeadObject` call per upload
- [ ] Allow `.webp`; English validation messages with KB/MB formatting
- [ ] Every new field is zero-value safe

**Done when:** upload and URL generation work for AWS S3, MinIO/R2 (path style), and a CDN.

### T5.4 CORS and body limit · S
- [ ] Enable the CORS middleware from `server.enable_cors` (config already exists)
- [ ] Config validation rejects `allow_origin: *` together with `allow_credentials: true`
- [ ] `server.body_limit` configurable (default 10MB)

**Done when:** CORS headers appear only when enabled, and oversized requests get 413.

### T5.5 Session management endpoints · M
**Depends on:** T3.5
- [ ] `GET /users/me/sessions` — active sessions (device name/type, IP, `last_used_at`, current flag)
- [ ] `DELETE /users/me/sessions/:id` — revoke one of the user's own sessions (ownership enforced)
- [ ] `POST /auth/logout-all` keeps the option to exclude the current session

**Done when:** a user can see their devices and log out a specific other device.

### T5.6 Refresh token via httpOnly cookie (web clients) · M
**Depends on:** T3.3
- [ ] Config `auth.refresh_transport: body | cookie` (default `body`)
- [ ] `cookie`: `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/api/v1/auth`; refresh token
      omitted from JSON
- [ ] CSRF protection: `Origin` allowlist check on cookie-authenticated endpoints
- [ ] `SameSite=Strict` requires the frontend and API to be same-site (e.g. `app.example.com` and
      `api.example.com`); for cross-site frontends use `SameSite=None` plus a CSRF token
- [ ] CORS `allow_credentials: true` with explicit origins (T5.4)

**Done when:** in cookie mode, browser JavaScript cannot read the refresh token and refresh still works.

---

## Phase 6 — Tests & documentation

### T6.1 Test matrix · L
- [ ] Auth use-case unit tests with the memory `SessionStore` and a fake repository
- [ ] Integration tests (testcontainers or docker-compose) on PostgreSQL across cache modes
      `none`, `memory`, `redis`
- [ ] Required scenarios:
  - login → refresh → logout → old tokens rejected
  - login on devices A and B → logout on A → B still authenticates and refreshes
  - refresh token reuse on A revokes A only; B unaffected
  - refresh token reuse detection and grace period (idempotent re-issue)
  - logs contain no passwords, tokens, or API keys
  - spoofed `X-Forwarded-For` from an untrusted source is ignored
  - refresh after logout returns 401 without a reuse event
  - LogoutAll across devices
  - lockout and anti-enumeration
  - locked account: correct password gets the same response as a wrong one
  - concurrent one-time token consumption
  - OTP invalidated after max attempts
  - revoking a role permission is enforced on the next request
  - concurrent requests with the same `Idempotency-Key` execute once
  - gRPC call without a token
  - `/internal` with `internal_auth.mode=shared_secret` and no secret returns 401
  - startup config validation
- [ ] CI (GitHub Actions) runs the full matrix

**Done when:** CI is green for every combination.

### T6.2 Documentation · M
- [ ] `docs/guides/auth.md`: token flow, rotation, revocation modes, trade-offs of each cache
      mode, client contract
- [ ] `docs/deployment/configuration.md`: Minimal / Standard / Full profiles with example configs
- [ ] Update `CLAUDE.md` and `.claude/rules/` for changed conventions (PostgreSQL only,
      Null Object pattern, `one_time_tokens`)

**Done when:** a new developer can run the Minimal profile using only the docs.

---

## Dependency graph

```
T0.1 ─┬─ T1.1 ─ T1.5
      ├─ T1.2 ─┬─ T1.3
      │        └─ T1.4
      ├─ T1.6
      └─ T2.1 ─ T2.2 ─ T2.3 ─ T2.4

T3.2  needs T1.2, T2.3, T3.1
T3.3  needs T3.2
T3.5  needs T3.2  → T3.6 → T5.1 (also needs T2.4)
T4.3  needs T1.3, T3.2
T4.4  needs T1.2, T2.3, T2.4
T4.5  after Phase 3
T5.5  needs T3.5 · T5.6 needs T3.3

T4.7  needs T1.7

Start with: T1.7 (secrets currently leak into logs)
Parallel at any time: T3.4, T3.7, T4.1 (after T1.1), T4.2, T4.6, T5.2, T5.3, T5.4
T6.1 runs alongside every phase · T6.2 last
```
