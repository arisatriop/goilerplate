# Configuration Guide

Three profiles, from "PostgreSQL and nothing else" to everything switched on, and how secrets are
supplied in each.

← [Back to Documentation](../README.md)

---

## ⚡ Minimal profile — from clone to first login

PostgreSQL only. No Redis, no gRPC, no OTel, no S3.

```bash
cp config/config.example.yaml config/config.yaml   # edit the `db` block
make migrate-up
make run
```

Then create the first account:

```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"First User","email":"first@example.test","password":"a-strong-password"}'

curl -X POST http://localhost:3000/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"first@example.test","password":"a-strong-password"}'
```

That is the whole Minimal profile. `config.example.yaml` already has development JWT secrets in it,
so **no `.env` file is needed** for local work. The `owner` role that registration assigns is
created by a migration, so `make migrate-up` is enough — there is no seeding step.

`migrate` validates the **whole** config, the same as the server, but connects only to PostgreSQL.
A setting it never uses (CORS, say) can therefore stop it. That is deliberate: migrations run
before the new server starts, so a config the server would refuse stops the deploy before the
schema changes, instead of leaving a migrated database behind a server that cannot boot. Redis
being down, on the other hand, never blocks a migration.

What you get: HTTP API, JWT auth with server-side sessions, RBAC, local file storage, per-instance
rate limiting and idempotency. `auth.session_cache` resolves to `none`, so every check reads the
database.

---

## 🏗️ Standard profile — more than one instance

Add Redis. It is the only change that matters once you run more than one replica, because three
things are per-process without it:

| Without Redis | Consequence with several instances |
|---|---|
| Rate limiting | each instance counts separately, so the effective limit is `max × instances` |
| Idempotency | a duplicate that lands on another instance is processed again |
| Session / permission cache | `memory` mode lets one instance serve a session another revoked, for up to `session_cache_ttl` |

```yaml
redis:
  enabled: true
  host: redis:6379
  password: ${REDIS_PASSWORD}
  db: 0

auth:
  session_cache: auto   # resolves to redis once redis.enabled is true
  revocation: strict
```

Also set, for anything behind a load balancer or ingress:

```yaml
server:
  trusted_proxies: ["10.0.0.0/8"]   # empty means trust nobody
  proxy_header: X-Forwarded-For
```

Leaving `trusted_proxies` empty is the safe default — the client IP is then the peer that actually
connected. Setting it to a range you do not control lets a caller spoof their own IP and defeat
per-IP rate limiting.

---

## 🧰 Full profile — every optional component

[`config/config.full.example.yaml`](../../config/config.full.example.yaml) documents **every**
available key, with a comment on each. It is kept in sync by a test
(`config/example_test.go`), which fails if an option exists in code but is not documented there.

Top-level sections: `app`, `server`, `db`, `redis`, `jwt`, `auth`, `log`, `grpc`, `otel`,
`rate_limit`, `filesystem`, `crypto`, `api_key`, `internal_auth`, `jobs`, `service`.

Copy the sections you need; anything you leave out falls back to its default.

---

## ✅ Startup validation

The config is validated before anything starts (`config/validate.go`). On a problem the app prints
every one of them and exits:

```json
{"level":"ERROR","msg":"invalid configuration","errors":[
  "jwt.access_secret looks like an example value (contains \"changeme\"); use a randomly generated secret in production"
]}
```

It catches missing settings for an enabled feature, `<...>` placeholders left in, and secrets
shorter than 32 bytes. With `app.env: production` it additionally rejects the example and
low-entropy secrets that ship in `config.example.yaml` — so the development config **cannot** be
promoted to production by accident.

Non-fatal advisories are logged as warnings rather than blocking startup, for example
`internal_auth.mode=none` in production.

---

## 🔐 Secrets

### The naming rule

Any config key can be overridden by an environment variable: **take the path and replace dots with
underscores.**

| Config key | Environment variable |
|---|---|
| `db.password` | `DB_PASSWORD` |
| `jwt.access_secret` | `JWT_ACCESS_SECRET` |
| `jwt.refresh_secret` | `JWT_REFRESH_SECRET` |
| `redis.password` | `REDIS_PASSWORD` |
| `internal_auth.secret` | `INTERNAL_AUTH_SECRET` |
| `grpc.auth.secret` | `GRPC_AUTH_SECRET` |
| `crypto.encryption_key` | `CRYPTO_ENCRYPTION_KEY` |

`crypto.encryption_key` is the one entry in this table nothing reads yet. `pkg/crypto` is a
working AES-256-GCM primitive with no callers, and the key is the seam for when you encrypt a
column. Generate it rather than choosing it — the key is derived with a single unsalted
SHA-256, which does not stretch a memorable passphrase into anything an offline guess cannot
reach.

There is **no** `JWT_SECRET_KEY`: the access and refresh tokens are signed with separate secrets,
so that an attacker who obtains one cannot mint the other. There is no `DB_DRIVER` either —
PostgreSQL is the only supported database.

### Generating them

```bash
openssl rand -base64 48   # jwt.access_secret, jwt.refresh_secret, internal_auth.secret
```

Use a different value for each. Minimum 32 bytes, enforced at startup.

### ⚠️ `config/.env` is **not** read by the application

This surprises people, so it is worth being blunt: nothing in the server loads `config/.env`.
There is no dotenv dependency and no code that reads the file. Putting `DB_NAME` in it and
starting the app has no effect — the value in `config.yaml` is used instead, silently.

The file exists for the **MCP servers** in [`.mcp.json`](../../.mcp.json), which source it
themselves for the PostgreSQL, GitHub and Jira integrations.

If you want those values to reach the application, export them yourself:

```bash
set -a && . config/.env && set +a && make run
```

For the Minimal profile you need none of this: `config.yaml` carries everything.

### Production

Supply secrets as environment variables — never in `config.yaml`:

```bash
export APP_ENV=production
export DB_HOST=prod-db-host
export DB_PASSWORD=...
export JWT_ACCESS_SECRET=...
export JWT_REFRESH_SECRET=...
```

---

## 🔑 Rotating JWT secrets

Move the current key into `previous_keys` and set a new active one. Previous keys verify but never
sign, so nobody is signed out mid-rotation:

```yaml
jwt:
  key_id: v2
  access_secret: ${JWT_ACCESS_SECRET}
  refresh_secret: ${JWT_REFRESH_SECRET}
  previous_keys:
    - key_id: v1
      access_secret: ${JWT_PREVIOUS_ACCESS_SECRET}
      refresh_secret: ${JWT_PREVIOUS_REFRESH_SECRET}
```

Retire the old key once no token it signed can still be alive — `auth.session_expiry` after the
rotation, not `jwt.access_token_expiry`, because the refresh token lives as long as the session.

See the [Authentication Guide](../guides/auth.md#-rotating-jwt-secrets).

---

## ☸️ Kubernetes

ConfigMap for the non-sensitive config, Secret for the values above. See the
[Kubernetes Deployment Guide](./kubernetes.md) — including why `/internal` must never be exposed
through the ingress.

```bash
kubectl create configmap goilerplate-config -n <namespace> \
  --from-file=config.yaml=./config/config.yaml \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic goilerplate-secret -n <namespace> \
  --from-literal=DB_PASSWORD=... \
  --from-literal=JWT_ACCESS_SECRET=... \
  --from-literal=JWT_REFRESH_SECRET=... \
  --dry-run=client -o yaml | kubectl apply -f -
```

---

## 🧭 Choosing between the profiles

| | Minimal | Standard | Full |
|---|---|---|---|
| PostgreSQL | ✅ | ✅ | ✅ |
| Redis | ❌ | ✅ | ✅ |
| More than one instance | ❌ not safe | ✅ | ✅ |
| gRPC, OTel, S3 | ❌ | optional | ✅ |
| Session cache | `none` | `redis` | `redis` |

The honest summary: **Minimal is for one process.** Rate limiting, idempotency and cache
invalidation are all per-process without Redis, and each of them silently does the wrong thing
rather than failing loudly when you scale past one replica.

---

## 🔗 Related

- [Authentication Guide](../guides/auth.md) — revocation modes and cache-mode trade-offs in detail
- [Minimal config](../../config/config.example.yaml) · [Every option](../../config/config.full.example.yaml)
- [Kubernetes Guide](./kubernetes.md) · [Development Setup](../getting-started/development.md)
