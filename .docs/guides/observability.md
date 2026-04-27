# Observability Guide

Goilerplate includes built-in OpenTelemetry (OTel) distributed tracing across HTTP, gRPC, and database layers. Traces are exported via OTLP/gRPC to any compatible backend.

---

## What is instrumented

| Layer | Package |
|---|---|
| HTTP (Fiber) | `otelfiber` — every request gets a trace span |
| gRPC | `otelgrpc` stats handler — every RPC gets a trace span |
| Database | `gorm/plugin/opentelemetry` — every SQL query gets a span |

All instrumentation shares a single global `TracerProvider` initialized at startup (`internal/bootstrap/otel.go`).

---

## Configuration

```yaml
# config/config.yaml
otel:
  enabled: true            # false = no-op provider, zero overhead
  endpoint: localhost:4317 # OTLP gRPC endpoint of your backend
  insecure: true           # set false in production (requires TLS)
```

When `enabled: false` (default), a no-op provider is used — no performance impact.

---

## Reading traces

Each trace represents the full lifecycle of a request as a waterfall of spans:

```
POST /api/v1/bars          45ms   ← HTTP span
  └─ INSERT INTO bars...   38ms   ← DB span
```

Each span includes relevant attributes:
- HTTP: `http.method`, `http.status_code`, `http.route`
- DB: `db.statement` — the SQL query that ran
- Both: `error` with details if the span failed

### Identifying bottlenecks

- DB span nearly as long as the HTTP span → database is the bottleneck
- Multiple DB spans per request → possible N+1 query problem
- Long HTTP span with no child spans → logic/compute bottleneck

---

## Compatible backends

Any OTLP-compatible backend works — only `otel.endpoint` needs to change:

| Backend | Type |
|---|---|
| Jaeger | Self-hosted |
| Grafana Tempo | Self-hosted |
| Grafana Cloud | Managed |
| Datadog | Managed |
| New Relic | Managed |
| Elastic APM | Self-hosted / Managed |

No code changes are needed when switching backends.

---

## Production notes

- Set `otel.insecure: false` in production and configure TLS on your backend
- Use `AlwaysSample` (current default) for dev; consider a ratio sampler for high-traffic production

---

## Related

- [Configuration Guide](../deployment/configuration.md)
- [Development Guide](../getting-started/development.md)
