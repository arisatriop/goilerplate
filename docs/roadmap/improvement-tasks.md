# Goilerplate Improvement Task List

Improvement plan for the authentication flow and optional-component architecture of this
boilerplate. Based on a review of the current implementation (`user_tokens`, `user_sessions`,
auth middleware, bootstrap, and wiring).

**Sizing:** S = ≤ ½ day · M = 1–2 days · L = 3–5 days
**Status:** in progress — T1.7, T1.1, T1.5 done

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
| Client IP taken from `X-Forwarded-For` without a trusted-proxy check (spoofable); behind a load balancer `c.IP()` is the proxy, so per-IP rate limits are shared by all users | T4.6 |
| No security response headers | T4.6 |
| No change-password endpoint; no maximum password length (bcrypt rejects > 72 bytes, surfacing as a 500) | T3.6, T3.7 |
| Registration reveals whether an email is already registered | T5.1 |
| No security audit events (login failures, lockouts, token reuse, password resets) | T4.7 |
| Logout does not revoke access tokens issued by earlier refreshes | T3.2 |
| Refresh tokens are reusable until expiry; no rotation or reuse detection | T3.3 |
| `user_sessions.is_active` / `expires_at` never checked; `RememberMe` has no effect | T3.2, T3.3 |
| `user_tokens` grows unbounded (row per login and per refresh, no cleanup) | T3.2, T4.4 |
| DB `UPDATE used_at` on every authenticated request | T3.2 |
| `user_tokens` mixes session tokens with one-time tokens; `LogoutAll` would wipe pending reset/OTP tokens | T2.4, T3.5 |
| Login writes (session, tokens, cache) are not transactional | T3.4 |
| `LogoutAll` scans every `token:*` / `session:*` key in Redis | T1.2, T3.5 |
| Redis-enabled validation reads cache only; expiry check skipped | T3.2 |
| Logout errors printed with `fmt.Printf` and swallowed | T3.5, T4.5 |
| Redis cache implementation lives in `domain/auth`; domain imports GORM and Fiber | T1.2, T1.6 |
| `/internal` (intended for pod-to-pod only) relies solely on gateway path rules; no safety net if the gateway is misconfigured, and the deployment docs do not state the rule | T4.1 |
| Partner API key compared with `==` (not constant time); raw key stored in context | T4.2 |
| gRPC server has no auth interceptor; reflection toggled by `app.env` | T4.3 |
| Idempotency middleware becomes a no-op without Redis | T1.4 |
| Migrations are PostgreSQL-only although MySQL is a supported driver | T2.1 |
| Time columns use `TIMESTAMP` without timezone | T2.2 |
| GORM models use MySQL column types; redundant/unused indexes | T2.3, T2.4 |
| Lockout off-by-one (uses pre-increment attempt count) | T3.7 |
| Login reveals account existence (disabled status returned before password check) | T3.7 |
| JWT validation uses `ParseUnverified` to pick a secret, with a generic fallback secret; no `aud`/`iss` checks | T3.1 |
| CORS config exists but the middleware is commented out; 100MB body limit hardcoded | T5.4 |
| S3 driver: no path-style option, public-read only, extra `HeadObject` per upload | T5.3 |
| Permission cache TTL is 7 days and is only refreshed on login/refresh; role or permission changes are not invalidated | T1.2 |
| Idempotency middleware lets concurrent duplicates both execute and does not detect a reused key with a different payload | T1.4 |
| `/auth/refresh` and `/auth/logout` share the per-IP login rate limit, so users behind one NAT throttle each other | T4.6 |
| No tests for `domain/auth` or auth middleware | T6.1 |
| Redis timeouts are multiplied by `time.Second` twice (`5s` config → ~158 years), so dial/read/write/pool timeouts never fire | T1.2 |
| Wrong email or password returns `400` (`utils.ClientErr(http.StatusBadRequest, MsgInvalidCredential)` in `domain/auth/user_validator.go`); API conventions require `401` | T3.7 |

---

## Summary

| Phase | Tasks | Estimate |
|---|---|---|
| 0. Decision | T0.1 | ✅ decided |
| 1. Optional-component foundation | T1.1 – T1.7 | ~8–10 days |
| 2. Schema & database | T2.1 – T2.4 | ~3–5 days |
| 3. Token security | T3.1 – T3.7 | ~9–12 days |
| 4. Other auth surfaces | T4.1 – T4.7 | ~6–8 days |
| 5. Optional features | T5.1 – T5.6 | ~10–13 days |
| 6. Tests & documentation | T6.1 – T6.2 | ~3–4 days |

---

## Phase 0 — Decision

### T0.1 Decide database support
- [x] Choose database support → **PostgreSQL only** (see D3)
- [ ] Record the decision in the README (done together with T2.1)

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
- [ ] Log a warning when `internal_auth.mode=none` → moved to T4.1 (the config key does not exist yet)
- [x] Collect all errors and report them together; secret values never appear in messages

`jwt.secret_key` is only checked for presence and placeholders because T3.1 removes it.

**Done when:** invalid config stops the app with a clear list of errors; each rule has a unit test.

### T1.2 Cache abstraction: interfaces + Null Object · L — 🚧 in progress (foundation merged)
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
- [ ] Move the Redis implementation out of `domain/auth/cache_service.go`
- [ ] Config `auth.session_cache: auto|none|memory|redis`
- [ ] Remove every `IsEnabled()` call from `internal/domain/**`
- [ ] Add `auth.ErrNotFound` so the domain no longer imports GORM
- [ ] Log a startup warning when `memory` is used (revocation lag across instances)
- [ ] `PermissionCache` TTL from `auth.permission_cache_ttl` (default 15m, not the session lifetime)
- [ ] Invalidate cached permissions on every change: role permissions, role menus, user roles,
      user permission overrides (per user, or all users for role-level changes)
- [x] Fix Redis client timeouts in `internal/bootstrap/redis.go`: `dial_timeout`, `read_timeout`,
      `write_timeout`, and `pool_timeout` are already `time.Duration`; pass them directly instead of
      `time.Second * time.Duration(...)`

**Done when:** `grep -r "IsEnabled\|gorm" internal/domain` returns nothing, the app works in
all three cache modes, and a revoked permission is denied on the next request.

### T1.3 Conditional wiring · M
**Depends on:** T1.2
- [ ] `grpc.enabled=false` → gRPC server not constructed, services not registered
- [ ] `otel.enabled=false` → no OTel handlers attached
- [ ] Partner routes registered only when `api_key` is non-empty
- [ ] Background jobs start only when `jobs.*.enabled`
- [ ] Startup log prints a summary of enabled components

**Done when:** with the minimal config, no gRPC port is opened and no Redis/OTel connection is attempted.

### T1.4 Idempotency: memory fallback and correctness · M
**Depends on:** T1.2
- [ ] Use memory storage when Redis is disabled instead of becoming a no-op
- [ ] Log a startup warning that memory storage only deduplicates within one instance
      (the app cannot detect how many instances run)
- [ ] In-flight lock per key (`LockProvider`): a concurrent request with the same key gets
      `409 Conflict` instead of executing twice
- [ ] Store a fingerprint (method + path + body hash) with the key: same key with a different
      payload gets `422 Unprocessable Entity`

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

### T1.6 Framework-free domain · M
- [ ] Introduce a `DeviceRequest` struct (user agent, forwarded IPs, accept headers, remote IP);
      `DeviceService` no longer accepts `*fiber.Ctx`
- [ ] Move audit context helpers to `pkg/auditctx`, replacing `infrastructure/context`
- [ ] HTTP (and gRPC) handlers build `DeviceRequest` themselves

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

### T2.1 Remove MySQL support · S
**Depends on:** T0.1
- [ ] Delete `internal/bootstrap/database/mysql.go` and the MySQL branch in `gorm.go` / `app.go`
- [ ] Remove `db.driver` from config (or accept only `postgres` and fail fast otherwise)
- [ ] Remove the `go-sql-driver/mysql` and `gorm.io/driver/mysql` dependencies (`go mod tidy`)
- [ ] Document "PostgreSQL only" in the README, `CLAUDE.md`, and `docs/`

**Done when:** the project builds without MySQL dependencies and `make migrate-up` succeeds on PostgreSQL.

### T2.2 Migration baseline + UTC · S
- [ ] Every time column is `TIMESTAMPTZ`
- [ ] Merge `add_auth_fields_to_users` into `create_users_table`
- [ ] Rename `001_create_bars_table` to the same timestamp naming as the other migrations
- [ ] Primary keys: UUIDv7 generated in the application (`uuid.NewV7()`, already available in
      `google/uuid`) for index locality; no reliance on DB-side `gen_random_uuid()` defaults
- [ ] `time.Local = time.UTC` in `cmd/server` and `cmd/migrate`
- [ ] GORM `NowFunc: utils.Now`; DSN `TimeZone=UTC`

**Done when:** `make migrate-up` then `make migrate-down` (all steps) runs cleanly on an empty
database, and API timestamps always end in `Z` regardless of server timezone.

### T2.3 `user_sessions` schema · M
- [ ] Columns: `id`, `user_id`, `refresh_jti`, `previous_refresh_jti`, `rotated_at`,
      `device_name`, `device_type`, `device_id`, `ip_address`, `user_agent`, `is_active`,
      `expires_at`, `last_used_at`, `revoked_at`, `revoked_reason`, `created_at`
- [ ] No `refresh_token_hash` (replaced by `refresh_jti`)
- [ ] `ip_address` as `INET`
- [ ] Indexes: `(user_id, is_active)` and `expires_at` only
- [ ] Align the GORM model with the schema

**Done when:** migrations apply and roll back cleanly and the model matches the schema.

### T2.4 `one_time_tokens` table · M
- [ ] Replace the `user_tokens` migration with `create_one_time_tokens_table`
- [ ] `token_type` restricted by `CHECK` to `email_verification | password_reset | email_change`
- [ ] Indexes: `token_hash` UNIQUE, `(user_id, token_type)`, `expires_at` — nothing else
- [ ] `attempts` column (failed verification attempts, used by OTP limits in T5.1)
- [ ] Accurate column/table comments; GORM model without MySQL-style `char(36)` / `datetime(3)` tags
- [ ] Repository `ConsumeToken(hash, type)`:
      `UPDATE ... SET used_at = now WHERE token_hash = ? AND token_type = ? AND used_at IS NULL AND expires_at > now`,
      then require `RowsAffected == 1`

**Done when:** two concurrent requests with the same token yield exactly one success (covered by a test).

---

## Phase 3 — Token security

### T3.1 JWT hardening · M
- [ ] Remove `ValidateToken` (`ParseUnverified` + `secretKey` fallback); use explicit
      `ValidateAccessToken` / `ValidateRefreshToken`
- [ ] Parser options: `WithValidMethods` (configured algorithm only), `WithIssuer`,
      `WithAudience`, small leeway
- [ ] Key rotation: `kid` header on every token; config holds one active signing key plus
      optional previous verification keys, so secrets rotate without logging everyone out
- [ ] Optional, may be deferred until a second service needs to verify tokens:
      `jwt.algorithm: HS256 | EdDSA` (default `HS256`); `EdDSA` lets other services verify with
      only the public key
- [ ] `aud` claim from new `jwt.audience` config
- [ ] `jti` on both access and refresh tokens
- [ ] Remove `jwt.secret_key` and `jwt.refresh_token_expiry` from config (replaced by
      `auth.session_expiry` / `auth.remember_me_expiry`)
- [ ] Single `pkg/hash.Token()` helper replacing the four `hashToken` copies

**Done when:** tokens with the wrong issuer, audience, algorithm, or secret are rejected (one test each),
and a token signed with a previous key still validates after rotating the active key.

### T3.2 Stateless access token + session check · L
**Depends on:** T1.2, T2.3, T3.1
- [ ] `Authenticate` middleware: verify signature → `SessionStore.Get(claims.SessionID)` →
      require active and unexpired
- [ ] Cache miss → read DB → populate cache
- [ ] Config `auth.revocation: strict | refresh_only`
- [ ] Remove `TokenStorage`, access-token persistence, blacklist, `CacheToken` / `GetToken`,
      and `MarkTokenAsUsed`
- [ ] Put `session_id` in context instead of `token_hash`

**Done when:**
- in `strict` mode, logout rejects **every** access token of that session immediately, with
  cache `none`, `memory` (single instance), and `redis`
- no DB writes happen per request

### T3.3 Refresh token rotation + reuse detection · L
**Depends on:** T3.2
- [ ] Refresh issues a new access token **and** a new refresh token
- [ ] Load the session by `claims.SessionID` first: missing, inactive, or expired → plain 401
      (**not** a reuse event — e.g. a refresh after logout)
- [ ] Atomic rotation:
      `UPDATE user_sessions SET previous_refresh_jti = refresh_jti, refresh_jti = ?, rotated_at = now, last_used_at = now WHERE id = ? AND refresh_jti = ? AND is_active AND expires_at > now`
- [ ] When `RowsAffected = 0` on an active session, **re-read the session after the failed
      UPDATE** (the pre-loaded copy may predate a concurrent rotation; PostgreSQL row locking
      makes the UPDATE wait for the concurrent commit), then:
  - jti equals `previous_refresh_jti` and `rotated_at` is within `auth.refresh_reuse_grace` →
    **idempotent success**: return a new access token and re-issue a refresh token carrying the
    session's current `refresh_jti` (no further rotation). Handles concurrent tabs and a lost
    refresh response without client retry logic.
  - otherwise → **reuse detected**: revoke **only that session** (`revoked_reason = refresh_reuse`)
    and log a security event; other devices of the same user are unaffected
- [ ] Refresh JWT `exp` = `session.expires_at`; rotation never extends the session
- [ ] Session lifetime from `auth.session_expiry` / `auth.remember_me_expiry`
- [ ] Rotation is always on (no config switch): a non-rotating refresh token is the weaker
      option under RFC 9700 and not worth supporting

**Done when:**
- reusing an old refresh token after the grace period revokes that session only
- refreshing after logout returns 401 without a security event
- two concurrent refreshes within the grace period both succeed and end up holding the same
  current refresh token
- a client that lost the refresh response can refresh again with the old token within the grace period
- `remember_me` yields a session that expires after `auth.remember_me_expiry`

### T3.4 Transactional login · S
- [ ] Wrap login-info update and session creation in one DB transaction
- [ ] Write to cache only after commit

**Done when:** a failure midway leaves no orphaned session.

### T3.5 Logout and LogoutAll · M
**Depends on:** T3.2
- [ ] Logout: deactivate the session (`revoked_reason = logout`) and evict it from cache;
      **idempotent** (200 even if already logged out)
- [ ] LogoutAll: `UPDATE user_sessions SET is_active = false ... WHERE user_id = ?`, then
      `SessionStore.DeleteByUser`
- [ ] Remove `fmt.Printf`, empty `if err {}` blocks, and `//nolint` markers
- [ ] Logout must only touch the session identified by `claims.SessionID` — never query or
      revoke by `user_id` or `device_id` (the device fingerprint can collide across devices
      behind the same NAT with the same browser)

**Done when:**
- logout on device A leaves device B's access and refresh tokens fully working
- after LogoutAll every token on every device is rejected
- both are tested in all three cache modes

### T3.6 Revoke sessions on user state changes · S
**Depends on:** T3.5
- [ ] User deactivated → LogoutAll
- [ ] Add `PUT /users/me/password`: requires the current password; revokes every other session
      and keeps the current one
- [ ] Password reset (T5.1) → LogoutAll including the current session
- [ ] Refresh keeps rejecting inactive users

**Done when:** a deactivated user loses API access immediately, and changing the password
logs out every other device.

### T3.7 Atomic lockout + anti-enumeration · M
- [ ] Single statement:
      `UPDATE users SET failed_login_attempts = failed_login_attempts + 1, locked_until = CASE WHEN failed_login_attempts + 1 >= ? THEN ? ELSE locked_until END WHERE id = ?`
- [ ] Check order:
  1. Unknown user → dummy bcrypt comparison → "invalid credentials"
  2. **Locked → reject without evaluating the password** (still run a dummy bcrypt for equal
     timing). Evaluating the password while locked would let an attacker keep guessing and
     learn when a guess is correct, defeating the lockout.
  3. Wrong password → increment attempts → "invalid credentials"
  4. Disabled → "account disabled" (only revealed to someone who knows the password)
- [ ] Locked response message is generic ("too many attempts, try again later")
- [ ] Invalid credentials return `401 Unauthorized` instead of `400` (API conventions); `400` stays
      for request validation errors only
- [ ] Max attempts and lock duration configurable via `auth.lockout.*`
- [ ] Password policy on register, change, and reset (NIST 800-63B): minimum 8 characters,
      maximum 72 **bytes** (bcrypt limit) returned as a 400 validation error; no composition rules
- [ ] Reject common passwords using an embedded list (e.g. top 10k), as NIST 800-63B requires
      checking against commonly used or compromised passwords

**Done when:**
- the account locks exactly on attempt N
- while locked, the correct password gets the same response as a wrong one
- the response for an unregistered email is identical to a wrong password (both `401`)
- a concurrency test covers the increment

---

## Phase 4 — Other auth surfaces

### T4.1 Gateway-controlled `/internal` routes · S
**Depends on:** T1.1

`/internal` stays on the public port and is reachable only through in-cluster traffic (D5).

- [ ] Gateway / Ingress rule documented and enforced: forward only an explicit allowlist
      (`/api`, `/partner`, `/health`); never a catch-all `/`. Where the gateway supports it, add
      an explicit deny for `/internal`
- [ ] Optional safety net: `internal_auth.mode: none | shared_secret` (default `none`);
      `X-Internal-Secret` compared in constant time. Recommended when the gateway is managed by
      another team or its config changes often
- [ ] Startup warning in production when `internal_auth.mode=none` (moved from T1.1)
- [ ] Optional caller identity: record `X-Service-Name` in context and logs instead of a fixed
      `system` user
- [ ] `docs/deployment/kubernetes.md`: state the rule and a verification command
      (`curl -i https://<public-host>/internal/...` must not reach the app)

**Done when:**
- a request to `/internal/*` through the public gateway does not reach the app
- with `shared_secret` enabled, requests without the secret get 401

### T4.2 Harden partner API keys · S
- [ ] `subtle.ConstantTimeCompare`
- [ ] Optional: store keys in config as SHA-256 hashes
- [ ] Put the partner name in context, never the raw key

**Done when:** keys never appear in context or logs.

### T4.3 gRPC auth interceptor · M
**Depends on:** T1.3, T3.2
- [ ] Config `grpc.auth.mode: token | shared_secret | none` (default `token`)
  - `token`: unary and stream interceptors read `authorization` from metadata and reuse the same
    validator and `SessionStore` as HTTP, so handlers get the same user context
  - `shared_secret`: service-to-service calls without user context
  - `none`: acceptable only when the gRPC port is reachable in-cluster only (same reasoning as D5)
- [ ] Config allowlist for public methods (e.g. health)
- [ ] Reflection controlled by `grpc.reflection` (default off), not by `app.env`
- [ ] Optional `grpc.tls` for deployments where the port leaves the cluster

**Done when:** in `token` mode, gRPC calls without a valid token get `codes.Unauthenticated`,
and allowlisted methods still work.

### T4.4 Cleanup job · M
**Depends on:** T1.2, T2.3, T2.4
- [ ] Delete expired or used `one_time_tokens` older than the retention period
- [ ] Delete inactive or expired `user_sessions` older than the retention period
- [ ] Batched deletes
- [ ] Config `jobs.cleanup.enabled`, `interval`, `retention`
- [ ] Guard with `LockProvider.TryLock` (Redis: `SET NX`; memory/noop: always acquires)
- [ ] Stop gracefully on shutdown via context

**Done when:** the job runs in the minimal profile, and with Redis only one instance executes it.

### T4.5 Auth code cleanup · M
**Depends on:** Phase 3
- [ ] Remove duplicated `extractBearerToken` / validation between middleware and `TokenService`
- [ ] Extract the shared "menus + permissions" builder used by Login and Refresh
- [ ] Remove dead code: commented-out entities, `IsAdmin()`, `UserToken.IsRevoked`,
      `Repository.DeleteUserSessions`, `TokenService.ValidateAndGetClaims`,
      `TokenService.ValidateRefreshTokenAndGetUser`, `TokenStorage.MarkTokenAsUsedAsync`,
      unused expiry constants in `domain/auth/error.go`
- [ ] All logging through the structured logger

**Done when:** `golangci-lint` is clean and no unused functions remain.

### T4.6 HTTP hardening: client IP and security headers · S
- [ ] Config `server.trusted_proxies` (CIDR list) and `server.proxy_header`; enable Fiber
      `EnableTrustedProxyCheck` so forwarded headers are honored only from trusted proxies
- [ ] `DeviceService` and rate limiters use the resolved `c.IP()`; stop parsing
      `X-Forwarded-For` / `X-Real-IP` manually
- [ ] Security headers via Fiber `helmet` (`X-Content-Type-Options`, `X-Frame-Options`,
      `Referrer-Policy`; `Strict-Transport-Security` when `server.hsts` is on)
- [ ] Partner rate-limit key uses a hash of the API key, not the raw key
- [ ] Per-IP auth limiter applies only to unauthenticated endpoints (login, register, forgot
      password, OTP); `/auth/refresh` and `/auth/logout` are limited per session instead

**Done when:** a spoofed `X-Forwarded-For` from an untrusted source does not change the client IP,
and security headers are present on every response.

### T4.7 Security event logging · S
**Depends on:** T1.7
- [ ] Structured `security` log events with `user_id`, `session_id`, IP, and user agent (never secrets):
  - login success / failure
  - account locked
  - refresh token reuse detected
  - logout-all and session revoked
  - password changed / reset requested / reset completed
  - email changed
- [ ] Distinct log label so events can be routed to a SIEM or alerting

**Done when:** each event above emits exactly one structured log entry (covered by tests).

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
