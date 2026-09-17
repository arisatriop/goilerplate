# Configuration Guide

Guide for setting up environment variables and configuration for development and production.

---

## 📋 Local Development

### Quick Setup

Copy `config.example.yaml` to `config.yaml` and point `db` at your database:

```bash
cp config/config.example.yaml config/config.yaml
```

`config.example.yaml` is the **minimal profile**: PostgreSQL only, with Redis, gRPC, OTel off and
`local` storage. Every available option is documented in `config/config.full.example.yaml`; copy
the sections you need from there.

The config is validated at startup (`config/validate.go`). The app exits with a list of every
problem, such as missing settings for an enabled feature, `<...>` placeholders, or JWT secrets
shorter than 32 bytes. With `app.env: production`, example or low-entropy secrets (including the
development secrets in `config.example.yaml`) are rejected too.

### Auth cache modes

`auth.session_cache` selects how sessions and permissions are cached. The database is always the
source of truth; a cache only saves lookups.

| Mode | Behavior | Use for |
|---|---|---|
| `auto` (default) | `redis` when `redis.enabled`, otherwise `none` | Most setups |
| `none` | Every request reads the session and permissions from the database | Minimal profile |
| `memory` | In-process cache; with several instances, logout and permission changes can take up to `session_cache_ttl` / `permission_cache_ttl` to reach other instances (a startup warning is logged) | Single instance without Redis |
| `redis` | Shared by every instance; requires `redis.enabled` | Standard / Full profiles |

You don't need a `.env` file for local development, everything is in `config.yaml`.

### Configuration File: `config/config.yaml`

```yaml
app:
  env: local # Environment: local, development, production
  name: Goilerplate
  version: 1.0.0

server:
  host: localhost
  port: 3000
  prefork: false # Enable prefork for production
  read_timeout: 5s
  write_timeout: 5s
  idle_timeout: 120s
  enable_cors: true

db:
  driver: postgres # postgres or mysql
  host: localhost
  port: 5432
  name: postgres
  username: postgres
  password: postgres
  min_open_connections: 10
  max_open_connections: 100

redis:
  enabled: true # Set false to disable Redis (falls back to in-memory for rate limiting & idempotency)
  host: localhost:6379
  password: ""
  db: 0

jwt:
  access_token_expiry: 15m
  refresh_token_expiry: 168h # 7 days

log:
  level: debug # debug, info, warn, error
  source: false # Include source code location

otel:
  enabled: false          # Set true to enable distributed tracing
  endpoint: localhost:4317 # OTLP gRPC endpoint of your backend
  insecure: true          # Disable TLS — set false in production
```

---

## 🌍 Environment Variables

Override config values with environment variables:

```bash
export APP_ENV=production
export SERVER_PORT=8080
export DB_HOST=your-db-host
export DB_PORT=5432
export DB_NAME=your_database
export DB_USERNAME=your_username
export DB_PASSWORD=your_password
export REDIS_HOST=your-redis-host
export JWT_SECRET_KEY=your-super-secret-jwt-key
```

---

## 🚀 Production Configuration

### Setup in Production

1. **Copy and edit config**
   ```bash
   cp config/config.example.yaml config/config.yaml
   ```

2. **Set environment variables** (safer for secrets):
   ```bash
   export APP_ENV=production
   export SERVER_PORT=3000
   export DB_HOST=prod-db-host
   export REDIS_HOST=prod-redis-host
   export JWT_SECRET_KEY=your-secret-key
   ```

3. **Run server**
   ```bash
   go run cmd/server/main.go
   ```

---

## ☸️ Kubernetes Configuration

For Kubernetes deployment, use **ConfigMap** for non-sensitive config and **Secret** for sensitive values.

See: [Kubernetes Deployment Guide](./kubernetes.md)

### Quick Commands

**Create ConfigMap from file:**
```bash
kubectl create configmap goilerplate-config -n <namespace> \
  --from-file=config.yaml=./config/config.example.yaml \
  --dry-run=client -o yaml | kubectl apply -f -
```

**Create Secret from .env file:**
```bash
kubectl create secret generic goilerplate-secret -n <namespace> \
  --from-env-file=./config/.env \
  --dry-run=client -o yaml | kubectl apply -f -
```

**Create Secret from literal values:**
```bash
kubectl create secret generic goilerplate-secret -n <namespace> \
  --from-literal=DB_HOST=prod-db \
  --from-literal=DB_PASSWORD=secret123 \
  --dry-run=client -o yaml | kubectl apply -f -
```

---

## 🔐 Best Practices

✅ **DO:**
- Use `.env` file for local development (gitignored)
- Use environment variables for production
- Use ConfigMap for non-sensitive config in K8s
- Use Secret for sensitive values in K8s
- Rotate JWT secrets periodically
- Enable Redis in production — rate limiting and idempotency are in-memory only when Redis is disabled, which means limits are not shared across instances

❌ **DON'T:**
- Don't commit `.env` file to git
- Don't hardcode secrets in code
- Don't expose database credentials in logs
- Don't use default passwords in production
- Don't disable Redis in multi-instance deployments — idempotency and rate limiting will not work correctly

---

## 📝 Example `.env` File

Create `config/.env` for local development (gitignored):

```env
APP_ENV=local
SERVER_HOST=localhost
SERVER_PORT=3000

DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goilerplate
DB_USERNAME=postgres
DB_PASSWORD=postgres

REDIS_ENABLED=true
REDIS_HOST=localhost:6379
REDIS_PASSWORD=

JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
JWT_SECRET_KEY=local-secret-key-only-for-dev
```

---

## 🔗 Related

- [Main Configuration Example](../../config/config.example.yaml)
- [Kubernetes Guide](./kubernetes.md) - Deploy with ConfigMap & Secret
- [Development Setup](../getting-started/development.md) - Local development workflow
