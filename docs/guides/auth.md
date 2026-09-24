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
| wrong password | `{"success": false, "message": "Invalid credential"}` |
| unknown email | `{"success": false, "message": "Invalid credential"}` — **identical**, so login cannot be used to discover which addresses are registered |
| correct password, account locked | `{"success": false, "message": "Too many failed attempts, please try again later"}` |
| wrong password, account locked | same "Too many failed attempts" message |

The locked response does say the account is locked rather than imitating a wrong password. That is
deliberate: hiding it would leave a locked-out user with a correct password and no explanation, and
the only person it tells anything new is someone who just spent the attempts to cause the lock.

⚠️ **Registration is different, and it is a known gap.** `POST /api/v1/auth/register` answers
`400 {"message": "email is already registered"}` for a taken address and `201` for a free one, so
it *can* be used to discover which addresses are registered — undoing on one endpoint what login
is careful about on another. Closing it is part of
[T5.1](../roadmap/improvement-tasks.md), which replaces the immediate answer with a verification
email either way.

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
  reuse, lockout, anti-enumeration), run against a real PostgreSQL under **every** cache mode,
  because `none` cannot detect a broken cache eviction.
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
