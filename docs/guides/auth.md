# Authentication Guide

How tokens are issued, rotated and revoked, what each configuration mode costs you, and what a
client has to implement.

← [Back to Documentation](../README.md)

---

## 🧭 The model in one paragraph

The access token is **stateless**: it is verified by signature alone, so no request has to ask the
database whether it is still good. That is what makes it cheap, and it is also why it cannot be
withdrawn — a signed token stays valid until it expires. So every token also carries a
**session ID**, and the session row in `user_sessions` is the one place a login can be switched
off. Logging out revokes the session, not the token; the token simply stops passing the session
check. Everything below follows from that split.

---

## 🔑 Token flow

### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{"email": "first@example.test", "password": "a-documented-password", "remember_me": false}
```

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": { "id": "01a0b541-...", "name": "Docs User", "email": "...", "isActive": true, "emailVerified": false, "lastLoginAt": null },
    "menus": [],
    "permissions": [],
    "tokens": {
      "accessToken": "eyJhbGciOiJIUzI1NiIsImtp...",
      "accessTokenType": "Bearer",
      "accessTokenExpiresIn": 900,
      "accessTokenExpiresAt": "2026-09-18T16:17:13.979061Z",
      "refreshToken": "eyJhbGciOiJIUzI1NiIsImtp...",
      "refreshTokenType": "Bearer",
      "refreshTokenExpiresIn": 604800,
      "refreshTokenExpiresAt": "2026-09-25T16:02:13.979061Z"
    },
    "session": {
      "id": "01a0b541-0b5b-70ea-8f71-c55b5a345f6c",
      "deviceId": "fp_113e2a2393e30", "deviceType": "web", "deviceName": "Web Browser",
      "ipAddress": "127.0.0.1", "userAgent": "curl/8.7.1",
      "isActive": true, "expiresAt": "2026-09-25T16:02:13.979139Z", "lastUsedAt": "2026-09-18T16:02:13.979139Z"
    }
  }
}
```

`permissions` is the list the server will actually enforce on later requests — it is resolved by
the same service the per-request check uses, so the menu a client renders and the access it gets
cannot drift apart.

### Authenticated request

```http
Authorization: Bearer <accessToken>
```

### Refresh

```http
POST /api/v1/auth/refresh
Authorization: Bearer <refreshToken>
```

The **refresh** token goes in the header, not the access token. The response is the same shape as
login, with a new token pair and the **same session ID** — refreshing does not start a new login.

### Refresh token in a cookie (browsers)

With `auth.refresh_transport: cookie` the refresh token never appears in a response body. Login and
refresh set it as a cookie instead, and refresh reads it from there:

```http
Set-Cookie: __Secure-refresh_token=<token>; Path=/api/v1/auth; HttpOnly; Secure; SameSite=Strict; Expires=...

POST /api/v1/auth/refresh          # no Authorization header; the browser sends the cookie
```

| Property | Why |
|---|---|
| `HttpOnly` | Page JavaScript cannot read it, so an XSS anywhere on the site does not yield a long-lived credential |
| `Path=/api/v1/auth` | Only refresh and logout receive it; no other request carries it |
| `Secure` + `__Secure-` prefix | Sent only over HTTPS, and a network attacker cannot plant one over HTTP. Off only for `app.env` local/dev/development/test, so the flow works on `http://localhost` |
| `SameSite=Strict` (default) | The browser does not attach it to a request another site starts |
| `Origin` check | A request carrying `Origin` must come from the API's own origin or `server.cors.allow_origin`, else `403 origin_not_allowed`. A request without `Origin` is not a browser page (curl, a server) and cannot be CSRF, so it passes |

- The access token still comes in the body; keep it in memory, not `localStorage`
- Cookie mode reads **only** the cookie — a refresh token sent as a Bearer header gets 401 — so a
  deployment has one answer to where the refresh token lives. Mobile and server clients use
  `body`, the default
- `logout` clears the cookie; `logout-all` clears it unless `keep_current=true`
- **Frontend on another site** (not just another subdomain): `same_site: none`, which config
  validation accepts only with a Secure cookie and `server.enable_cors` + `allow_credentials` +
  explicit origins. The browser must send `credentials: "include"`

### Logout

```http
POST /api/v1/auth/logout                     # this device
POST /api/v1/auth/logout-all                 # every device
POST /api/v1/auth/logout-all?keep_current=true   # every device except this one
Authorization: Bearer <accessToken>
```

### Managing devices

```http
GET    /api/v1/users/me/sessions        # devices this user is signed in on
DELETE /api/v1/users/me/sessions/{id}   # sign one of them out
Authorization: Bearer <accessToken>
```

The list holds active, unexpired sessions, most recently used first. Each entry carries the device
name and type, IP, `createdAt`, `lastUsedAt`, `expiresAt`, and `current: true` on the session making
the request. The refresh `jti` and the device fingerprint are never included.

`DELETE` answers **404** for a session that does not exist, is already revoked, or belongs to
someone else — the three are indistinguishable on purpose, so the endpoint cannot be used to
confirm another user's session IDs. Ownership is enforced in the same conditional `UPDATE` that
revokes the session, so there is no check-then-act window. A malformed ID is a 400 and never
reaches the database.

### Forgot and reset password

Exists only with `auth.email.enabled: true`; otherwise the routes are not registered at all.
Both are unauthenticated and rate limited per IP, like login.

```http
POST /api/v1/auth/forgot-password
{"email": "ana@example.org"}

POST /api/v1/auth/reset-password
{"token": "<from the emailed link>", "newPassword": "..."}
```

1. `forgot-password` **always answers 200** with the same message — for a registered address, an
   unknown one, a disabled account, a repeat inside the cooldown, and even when the email could not
   be queued. The response cannot be used to discover which addresses are registered; the
   difference shows only in the security log (`password_reset_requested`, with `reason`).
2. For a registered, active account it issues a 256-bit random token, stores only its SHA-256, and
   emails a link to `frontend.base_url` + `frontend.reset_password_path` with `?token=…`. The
   frontend page reads the token and posts it to `reset-password`; the API serves no pages itself.
3. Issuing a link **expires every earlier one** for that account, in the same transaction. A
   second request inside `auth.password_reset.resend_cooldown` (default `1m`) sends nothing, so the
   endpoint cannot be used to flood someone's inbox.
4. `reset-password` checks the new password against the policy **first**, so a refused password
   does not use up the link. It then consumes the token, replaces the password and revokes **every**
   session — the device asking included — in one transaction, and evicts the user's cached
   sessions. Consumption is one conditional `UPDATE`, so two requests racing with the same link
   give one success. A successful reset also clears a lockout.
5. An unknown, used, superseded or expired token is one answer:
   `400 {"code": "invalid_reset_token"}`. An account disabled after the link was sent stays
   disabled: its token is refused the same way.

Email is sent from an in-memory background queue (`pkg/email.Queue`), so the response time does
not depend on the provider either — a synchronous send would make a registered address measurably
slower than an unknown one. The queue is drained on shutdown; a message still queued when the
process is killed is lost, and the user asks again.

### Email verification

Also only with `auth.email.enabled: true`, and also unauthenticated: with
`auth.require_email_verification` on, an unverified user cannot sign in to get a token first.

```http
POST /api/v1/auth/send-verification-email
{"email": "ana@example.org"}

POST /api/v1/auth/verify-email
{"email": "ana@example.org", "code": "042917"}
```

1. A code is sent **automatically after registration**, once the account has committed. A
   failure to send it does not fail the registration; the user asks again.
2. `send-verification-email` answers 200 with one message for every case — unknown, disabled,
   already verified, inside `auth.otp.resend_cooldown`. A new code expires the previous one.
3. Codes are 6 random digits stored as **HMAC-SHA256** keyed by `auth.otp.secret` and the token's
   own ID. An unkeyed digest of six digits is reversed by hashing all million candidates; the
   key makes a leaked table useless on its own.
4. Each code tolerates `auth.otp.max_attempts` guesses (default 5). The attempt is counted in
   one atomic `UPDATE` **before** the code is compared, so concurrent guesses share the ceiling
   instead of all being compared at once. The last wrong guess expires the code.
5. Every failure is `400 {"code": "invalid_verification_code"}`.
6. A successful password reset also marks the address verified: following the emailed link
   already proves the user reads that inbox.

With `auth.require_email_verification: true` (default `false`), login with the right password
and an unverified address answers `403 {"code": "email_not_verified"}`. It is checked after the
password, like a disabled account, so it tells nothing to someone who does not know it.

### Changing the email address

Authenticated, and only with `auth.email.enabled: true`:

```http
POST /api/v1/users/me/email-change
{"newEmail": "ana.new@example.org", "currentPassword": "..."}

POST /api/v1/users/me/email-change/confirm
{"code": "042917"}
```

1. The current password is required (`401 invalid_credentials` without it): a signed-in session
   is enough to read the account, not to hand it to another inbox.
2. A code is sent to the **new** address; nothing changes until it is confirmed. The pending
   address is stored on the token (`one_time_tokens.new_email`).
3. An address another account holds is answered with the same 200 and **no** code, so a
   signed-in user cannot probe which addresses are registered. The same address as now is
   `400 email_unchanged`.
4. Confirmation uses the same code rules as verification: keyed hash, attempt ceiling counted
   before the comparison, `400 invalid_verification_code` for every failure. If the address was
   registered by someone else in the meantime, the unique constraint refuses the move:
   `409 email_already_registered`.
5. On success the account moves, the new address counts as verified, and the **old** address is
   told — the only warning a user gets if someone else did it. Sessions are kept; the warning
   points the user at a password reset, which signs out every device. Tokens issued from the next
   refresh on carry the new address.

The request body carries the password, so `/api/v1/users/me/email-change` is always left out of
body logging, like `/api/v1/auth` and `/api/v1/users/me/password`.

The mail provider is `email.driver`: `log` (prints the message — reset link included — to the
log; refused in production), `smtp`, `ses` or `resend`. See
[`config/config.full.example.yaml`](../../config/config.full.example.yaml).

---

## 🔄 Rotation, reuse and the grace window

Every refresh issues a new refresh token and invalidates the one presented. The session stores the
current `refresh_jti`, and the rotation is a conditional update: the token you hand in must be the
one the session currently expects.

A refresh token that the session does not accept, and cannot explain, means **someone is replaying
a token that was already spent** — the legitimate client and a thief cannot both hold the current
one. The session is revoked and a `refresh_token_reuse_detected` security event is written.

Revocation is deliberately narrow: **only that session**. The user's other devices keep working. A
wider rule would hand anyone who captured a single token the power to sign the victim out
everywhere.

### Why there is a grace window

Rotation has an unavoidable race. A client refreshes, the server rotates and replies, and the reply
is lost — or a second browser tab refreshes a moment after the first. The client then retries with
a token that is now one rotation old. That is not an attack, and treating it as one would log
people out for having flaky wifi.

So `auth.refresh_reuse_grace` (default `10s`) covers exactly one case: the token the **last**
rotation replaced, replayed **recently**. Within it, the server returns the session's current
token rather than rotating again, so both callers converge on the same token:

```
refresh with token A  → 200, issues B
refresh with token A  → 200, returns B again   (within the grace: a retry)
refresh with token A  → 401, session revoked   (two rotations behind: reuse)
```

Everything else — a token two rotations old, a token from a different session — is reuse.

---

## 🚪 Revocation modes (`auth.revocation`)

| Mode | The session is checked | Logout takes effect | Cost |
|---|---|---|---|
| `strict` (default) | on every authenticated request | immediately | one session lookup per request (cached) |
| `refresh_only` | only when refreshing | when the access token expires | no lookup on ordinary requests |

`refresh_only` bounds the exposure by `jwt.access_token_expiry` (default `15m`): after a logout,
the already-issued access token keeps working for up to that long. Shorten the access token expiry
if you choose this mode — that number *is* your revocation delay.

Refresh and logout always check the session, whichever mode is set. That is the point at which a
revoked login must stop being able to mint new access tokens.

---

## 🗃️ Cache modes (`auth.session_cache`)

The database is always the source of truth. A cache only saves lookups, and the risk it introduces
is always the same one: **serving a session the database has already revoked.**

| Mode | Shared across instances | Revocation lag | Use for |
|---|---|---|---|
| `auto` (default) | — | — | resolves to `redis` when `redis.enabled`, else `none` |
| `none` | n/a | none | the Minimal profile; every check reads the database |
| `memory` | ❌ per-process | up to `session_cache_ttl` on *other* instances | a single instance without Redis |
| `redis` | ✅ | none | anything running more than one instance |

`memory` is safe on one instance: the eviction and the request hit the same process. With several
instances, instance B may keep serving a session that instance A revoked until the entry expires —
which is what `session_cache_ttl` (default `30s`) actually bounds. A startup warning is logged when
that combination is configured.

`permission_cache_ttl` (default `15m`) is a safety net, not the mechanism: permission changes
invalidate the cache explicitly.

---

## 🔒 Lockout and what the client is told

After `auth.lockout.max_attempts` consecutive failures (default `5`) the account stops accepting
**any** password for `auth.lockout.duration` (default `10m`). While locked the password is never
evaluated, so a correct one cannot be distinguished from a wrong one — which is the whole point:
otherwise an attacker could keep guessing through the lockout and read the answer off the response.

Observed responses, all `401`:

| Attempt | Response |
|---|---|
| wrong password | `{"success": false, "code": "invalid_credentials", "message": "Invalid credential"}` |
| unknown email | `{"success": false, "code": "invalid_credentials", "message": "Invalid credential"}` — **identical**, so login cannot be used to discover which addresses are registered |
| correct password, account locked | `{"success": false, "code": "account_locked", "message": "Too many failed attempts, please try again later"}` |
| wrong password, account locked | same "Too many failed attempts" message |

The locked response does say the account is locked rather than imitating a wrong password. That is
deliberate: hiding it would leave a locked-out user with a correct password and no explanation, and
the only person it tells anything new is someone who just spent the attempts to cause the lock.

**Registration follows the same rule when email is on.** With `auth.email.enabled`,
`POST /api/v1/auth/register` answers a taken address exactly like a new one —
`201 {"message": "Registration received. Check your email to continue"}` — and emails the owner
"you already have an account" instead. Both paths run bcrypt, so the timing matches too; a
concurrent registration that wins the insert is answered the same way.

⚠️ Without email there is nobody to tell but the caller, so registration still answers
`409 email_already_registered` for a taken address and *can* be used to discover which addresses
are registered. Turn email on to close it.

---

## 📜 The client contract

What a client has to implement, in full:

1. **Store both tokens.** The access token is sent on every request; the refresh token is sent
   only to `/api/v1/auth/refresh`.
2. **Refresh on 401, once.** Any `401` on an ordinary request means the access token expired or
   the session was revoked. Try refresh once. If refresh also returns `401`, the session is gone —
   send the user to the login screen. Do not loop.
3. **Replace both tokens after every refresh.** The old refresh token is spent. Keeping it and
   retrying with it later is what reuse detection is looking for.
4. **Serialise refreshes.** Two concurrent refreshes from the same client are the case the grace
   window exists to forgive, but relying on it is a race. Hold one refresh in flight and let other
   requests wait for it.
5. **Do not parse the access token to decide what to show.** Use `permissions` from the login or
   refresh response. The server enforces its own copy regardless.
6. **Treat `403` as final.** It means the token is valid and the permission is missing; refreshing
   will not change that.

Requests may also send `X-Request-Id`; it is echoed into every log line for that request, which is
what makes a report traceable. A missing one is generated.

---

## 🔁 Rotating JWT secrets

`jwt.key_id` is published as each token's `kid` header. To rotate without signing everyone out,
move the current key into `previous_keys` and set a new active one:

```yaml
jwt:
  key_id: v2
  access_secret: <JWT_ACCESS_SECRET>
  refresh_secret: <JWT_REFRESH_SECRET>
  previous_keys:
    - key_id: v1
      access_secret: <JWT_PREVIOUS_ACCESS_SECRET>
      refresh_secret: <JWT_PREVIOUS_REFRESH_SECRET>
```

Previous keys verify but never sign. Retire one once no token it signed can still be alive — that
is `auth.session_expiry` after the rotation, because the refresh token lives as long as the
session, not `jwt.access_token_expiry`.

---

## ⚙️ Configuration reference

| Key | Default | What it controls |
|---|---|---|
| `jwt.access_token_expiry` | `15m` | access token lifetime; in `refresh_only` this is also the revocation delay |
| `auth.session_expiry` | `168h` | absolute session lifetime and the refresh token's lifetime; rotation never extends it |
| `auth.remember_me_expiry` | `720h` | the same, when the client logs in with `remember_me` |
| `auth.revocation` | `strict` | `strict` \| `refresh_only` |
| `auth.session_cache` | `auto` | `auto` \| `none` \| `memory` \| `redis` |
| `auth.session_cache_ttl` | `30s` | with `memory`, the cross-instance revocation lag |
| `auth.permission_cache_ttl` | `15m` | safety net; changes invalidate explicitly |
| `auth.refresh_reuse_grace` | `10s` | how long the just-replaced refresh token stays usable |
| `auth.lockout.max_attempts` | `5` | consecutive failures before the account locks |
| `auth.lockout.duration` | `10m` | how long the lock lasts |
| `auth.email.enabled` | `false` | registers forgot/reset password; requires `email.*` and `frontend.base_url` |
| `auth.password_reset.ttl` | `30m` | how long a reset link stays usable |
| `auth.password_reset.resend_cooldown` | `1m` | minimum gap between two reset emails to one account |
| `auth.require_email_verification` | `false` | refuse login (403) until the address is verified; needs `auth.email.enabled` |
| `auth.otp.secret` | — | HMAC key for verification codes, at least 32 bytes (`AUTH_OTP_SECRET`) |
| `auth.otp.ttl` | `15m` | how long a code stays usable |
| `auth.otp.max_attempts` | `5` | wrong guesses one code tolerates |
| `auth.otp.resend_cooldown` | `1m` | minimum gap between two codes to one account |
| `email.driver` | `log` | `log` \| `smtp` \| `ses` \| `resend`; `log` is refused in production |
| `frontend.base_url` | — | the web app reset links open; `https` in production |
| `frontend.reset_password_path` | `/reset-password` | the page that reads `?token=` |

Every option is documented in [`config/config.full.example.yaml`](../../config/config.full.example.yaml).

---

## 🧹 Retention

Revoked and expired sessions are kept: a revoked session is the record of a logout, and an incident
is usually investigated long after it happened. The optional cleanup job
(`jobs.cleanup`, off by default) deletes them once they are older than `retention` — which
validation refuses to set shorter than `auth.session_expiry`.

---

## 🧪 Where this is tested

- `internal/integration/` — the lifecycle end to end (login → refresh → logout, multi-device,
  reuse, lockout, anti-enumeration, password reset, email verification, email change), run against a real PostgreSQL under
  **every** cache mode, because `none` cannot detect a broken cache eviction.
- `internal/domain/auth/` — the rotation, session, permission and validator logic in isolation.
- `pkg/jwt/` — signing, `kid` selection, and rejecting an access token where a refresh token
  belongs (and the reverse).

Run them with `make test-integration`. Without `POSTGRES_TEST_DSN` and `REDIS_TEST_ADDR` those
suites **skip** rather than fail, so plain `make test` can be green while none of them ran.

---

## 🔗 Related

- [Configuration Guide](../deployment/configuration.md) — Minimal / Standard / Full profiles
- [Router & API Routes](../api/router.md) — public, partner and internal route groups
- [gRPC Guide](./grpc.md) — the same tokens and session store over gRPC
