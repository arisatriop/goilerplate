# Observability Guide

Goilerplate includes built-in OpenTelemetry (OTel) tracing across HTTP, gRPC, and database layers, with support for log aggregation via Loki and visualization via Grafana.

---

## Overview

```
App (HTTP/gRPC/DB spans)
  └─► OTel SDK ──► OTLP/gRPC ──► Tempo ──► Grafana (traces)

App (stdout logs)
  └─► tee ──► storage/logs/app.log ──► Promtail ──► Loki ──► Grafana (logs)
```

### What is instrumented automatically

| Layer | Instrumentation |
|---|---|
| HTTP (Fiber) | `otelfiber` — every request gets a trace span |
| gRPC | `otelgrpc` stats handler — every RPC gets a trace span |
| Database | `gorm/plugin/opentelemetry` — every SQL query gets a span |

---

## Local Setup

### Prerequisites

The observability stack lives in a separate `docker-image` repo alongside this project:

```
others/
├── goilerplate/         # this repo
└── docker-image/
    └── observability/   # Tempo + Loki + Promtail + Grafana
```

### 1. Start the observability stack

```bash
cd ../docker-image/observability
docker compose up -d
```

Services started:

| Service | Port | Purpose |
|---|---|---|
| Tempo | `4317` | OTLP gRPC receiver (traces) |
| Tempo | `3200` | HTTP API (queried by Grafana) |
| Loki | `3100` | Log storage (queried by Grafana) |
| Promtail | — | Scrapes `storage/logs/app.log` → Loki |
| Grafana | `3001` | UI for traces and logs |

### 2. Enable OTel in config

```yaml
# config/config.yaml
otel:
  enabled: true
  endpoint: localhost:4317
  insecure: true   # disable TLS for local dev
```

### 3. Run the app with log redirect

Logs must be written to a file so Promtail can scrape them.

**Using air (default):**
```bash
make run
# .air.toml is already configured to tee logs to storage/logs/app.log
```

**Using nodemon:**
```bash
mkdir -p ./storage/logs && nodemon -e go --signal SIGINT \
  --exec "sh -c 'go run cmd/server/main.go 2>&1 | tee ./storage/logs/app.log'"
```

### 4. View in Grafana

Open `http://localhost:3001` → **Explore**:

- **Traces**: datasource = **Tempo** → Search → Run query
- **Logs**: datasource = **Loki** → query `{job="goilerplate"}` → Run query

---

## Reading Traces

Each trace shows the full lifecycle of a request as a waterfall:

```
POST /api/v1/bars          45ms   ← HTTP span (otelfiber)
  └─ INSERT INTO bars...   38ms   ← DB span (otelgorm)
```

Click a span to see:
- `http.method`, `http.status_code`, `http.route`
- `db.statement` — the SQL query that ran
- `error` — with stack trace if the span failed

### Identifying bottlenecks

- If the DB span is nearly as long as the HTTP span → database is the bottleneck
- Multiple DB spans for one request → possible N+1 query problem
- Long HTTP span with no DB spans → logic/compute bottleneck

---

## Reading Logs in Loki

Logs are structured JSON. Use LogQL to filter:

```logql
# All logs
{job="goilerplate"}

# Only errors
{job="goilerplate"} | json | level="ERROR"

# Specific request
{job="goilerplate"} | json | request_id="abc-123"

# Incoming requests only
{job="goilerplate"} | json | label="incoming-request-log"
```

---

## Configuration Reference

```yaml
otel:
  enabled: true           # false = no-op provider, zero overhead
  endpoint: localhost:4317 # OTLP gRPC endpoint
  insecure: true          # set false in production (requires TLS)
```

In production, point `endpoint` to your observability backend (Grafana Cloud, Datadog, etc.) — no code changes required.

---

## Production Notes

- The observability stack (`docker-image/observability`) is for **local development only**
- In production, use a managed service (Grafana Cloud, Datadog, New Relic) or a dedicated infra repo
- Only the `otel.endpoint` in `config.yaml` needs to change per environment
- Set `otel.insecure: false` and configure TLS in production

---

## Related

- [Development Guide](../getting-started/development.md)
- [Configuration Guide](../deployment/configuration.md)
- [gRPC Guide](./grpc.md)
