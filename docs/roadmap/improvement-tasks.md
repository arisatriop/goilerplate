# Goilerplate Improvement Tasks

A repo-wide engineering plan derived from a full scan of the codebase on 2026-09-22
(194 Go files, ~19k LOC, 44 test files, commit `726eb76`).

This document replaces the earlier auth-only roadmap. That effort finished — its 27 tasks are
summarised in [Appendix A](#appendix-a--work-already-delivered) — and the remaining items from
it were folded into the phases below, re-prioritised against everything else the scan found.

## How this plan was written

Findings are measured against **general Go and backend practice**, not against this repository's
own `.claude/rules/*` or `CLAUDE.md`. Where the two disagree, this plan follows common practice
and the rule file is treated as the thing to correct — several tasks below do exactly that.

Every finding cites `file:line` so it can be checked rather than believed. Nothing here is a
style preference dressed up as a defect: tasks are ordered by whether the software is currently
wrong (P0), structurally misleading (P1), unguarded (P2), incomplete (P3), or noisy (P4).

**Severity:** P0 defect · P1 architecture · P2 hygiene · P3 feature · P4 cleanup
**Sizing:** S = ≤ ½ day · M = 1–2 days · L = 3–5 days

## Summary

| Phase | Theme | Tasks | Estimate |
|---|---|---|---|
| [P0](#p0--defects) | Defects — the code does not do what it says | D1 – D7 | ✅ complete |
| [P1](#p1--architecture-and-contracts) | Architecture and contracts | A1 – A5 | ~4–6 days |
| [P2](#p2--engineering-hygiene) | Build, CI, supply chain, tests | H1 – H6 | ~5–7 days |
| [P3](#p3--feature-completion) | Feature completion | F1 – F6 | ~10–13 days |
| [P4](#p4--cleanup) | Dead code and drift | C1 – C4 | ~1 day |

Recommended order: **D1 → D2 → H1 → D3–D7 → A1 → A2 → C1–C4 → H2–H6 → A3–A5 → P3**.
D1 and D2 are live defects; H1 (`-race`, stronger linters) is placed early because it changes
what every later task is checked against.

---

## P0 — Defects

Things that are measurably wrong today. Each one is a small change; none needs a design decision.

### D1 Apply the configured server timeouts · S — ✅ done
**Evidence:** `internal/bootstrap/fiber.go:15-19`, `config/config.go:186-192`,
`config/config.example.yaml:16-18`

`server.read_timeout`, `server.write_timeout` and `server.idle_timeout` appear in all three
example configs and in the `Server` struct, but `fiber.New` is only given `AppName`,
`ErrorHandler`, `Prefork`, `BodyLimit` and the trusted-proxy settings. The timeouts are parsed
and then dropped.

The consequence is not cosmetic: a Fiber server with no read timeout holds a connection open for
as long as a client keeps dribbling bytes, which is the whole of a slowloris attack. Operators
setting `read_timeout: 5s` believe they are protected and are not.

- [x] Pass `ReadTimeout`, `WriteTimeout` and `IdleTimeout` to `fiber.New`
- [x] Unset falls back to 15s/15s/60s via `OrDefault` helpers rather than to zero, because zero
      is "no limit" and is what a minimal config produces. Documented in both example configs
- [x] Tests asserting the values reach the Fiber config, and a real-socket test proving a
      stalled request is hung up on

**Done when:** a request that stalls mid-body is cut off at `server.read_timeout`.

### D2 Fix shutdown ordering and bound the gRPC drain · S — ✅ done
**Evidence:** `cmd/server/main.go:108-155`

`gracefulShutdown` runs in this order: drain Fiber → close GORM → close the pgx pool → close
Redis → `GrpcServer.GracefulStop()`. gRPC is therefore still accepting and serving calls after
the database and cache it depends on have been closed. Any in-flight gRPC request during a
rolling deploy fails on a closed pool rather than completing.

`GracefulStop()` also has no deadline. It blocks until every active RPC finishes, so a single
long-lived stream keeps the process alive past the pod's `terminationGracePeriodSeconds` and the
container is SIGKILLed instead of exiting cleanly.

- [x] `shutdown` now drains HTTP and gRPC concurrently, flushes telemetry, and closes DB/Redis
      last. Telemetry moved after the drain so spans from in-flight requests are still exported
- [x] Bounded gRPC drain. Note for anyone revisiting it: `Stop()` does **not** release an
      in-progress `GracefulStop()`, because `GracefulStop` also waits on `handlersWG` and a
      handler that never returns is the case this branch exists for. The timeout path calls
      `Stop()` and returns without waiting, rather than reintroducing the unbounded wait
- [x] Every `fmt.Printf` on the lifecycle path replaced with `app.Log`

**Done when:** SIGTERM drains both servers before any connection pool closes, and the process
always exits within the shutdown timeout.

### D3 Stop discarding the bcrypt error when hashing a password · S — ✅ done
**Evidence:** `internal/domain/user/entity.go:19-21`, `internal/application/register/service.go:49`

```go
func (u *User) HashPassword() {
	hasPassword, _ := utils.HashPassword(u.PasswordHash)
	u.PasswordHash = hasPassword
}
```

If bcrypt fails, the error is dropped and `PasswordHash` is assigned the empty string. The
account is then persisted with an empty hash. Nothing downstream can tell that apart from a
successful hash.

The signature is the root cause: a method that cannot fail forces the caller to ignore the
failure. The field name is a second problem — `PasswordHash` holds the *plaintext* until this
method runs, which `register/service.go:49` has to explain in a comment.

- [x] `SetPassword(plaintext string) error`. It also leaves `PasswordHash` untouched on
      failure rather than blanking it
- [x] The plaintext now travels in its own field — `Register.Password` for the registration
      flow, an explicit argument on `auth.Usecase.Register`. No field named for a hash holds an
      unhashed value
- [x] Hashing moved after the duplicate-email check: bcrypt at cost 12 costs ~250ms and there
      is no reason to spend it on a request already refused
- [x] Tests: `SetPassword` reports failure and keeps the old value; registration persists a
      hash and never the plaintext; a refused password reaches no repository and opens no
      transaction

**Done when:** no code path can persist a user whose password hash was never computed.

### D4 Stop returning internal error text from the health endpoint · S — ✅ done
**Evidence:** `internal/delivery/http/router/router.go:40-110`

`/healthcheck` is unauthenticated and embeds `err.Error()` from PostgreSQL, GORM and Redis
directly in its JSON. Driver errors routinely carry host names, ports, database names, user
names and network topology. That is free reconnaissance for anyone who can reach the port.

The two endpoints are also mis-named for what they do: `/health` (`router.go:33`) checks nothing
and is the liveness probe; `/healthcheck` checks dependencies and is the readiness probe.

- [x] Each dependency reports `healthy` or `unhealthy` and nothing else; the driver error is
      logged server-side
- [x] 503 when any check fails
- [x] `/livez` and `/readyz`, with `/health` and `/healthcheck` kept as aliases
- [x] **All three manifests pointed `readinessProbe` at `/health`** — the endpoint that checks
      nothing — so a pod with a dead database reported Ready and kept taking traffic. Startup
      and liveness now use `/livez`, readiness `/readyz`
- [x] Tests: 503 with a dependency down, 200 with none configured, liveness unaffected by the
      database, and an assertion that the body names no host, port, user or database

**Done when:** an unauthenticated caller can learn that the service is unhealthy and nothing
about why.

### D5 Remove the unused second connection pool · S — ✅ done
**Evidence:** `internal/bootstrap/database/database.go:14-18`, `internal/bootstrap/app.go:105`

The app opens a `pgxpool.Pool` *and* a GORM handle against the same database. Searching the whole
tree, `PgxDB` is referenced in exactly three places: it is created, pinged by the health check,
and closed at shutdown. **Not one query runs through it.**

It still opens `db.min_open_connections` connections at startup and holds up to
`db.max_open_connections`, so a deployment sized for 20 connections silently consumes 40. On
Cloud SQL, where connections are a hard quota, that is half the budget spent on nothing.

- [x] `PgxDB` and `NewPostgres` deleted. GORM already reaches PostgreSQL through the pgx
      stdlib driver, so nothing was lost
- [x] The health check now has one database entry rather than a `postgresql` and a `gorm` one
- [x] `db.min_open_connections` now configures the pool that exists, via `SetMaxIdleConns` —
      database/sql's spelling of the same warm floor. It was previously read only by the
      deleted pool
- [x] `db.health_check_period` dropped: it configured the pgx pool only, database/sql has no
      equivalent, and a config key nothing reads is the next defect waiting to happen
- [x] Decision recorded in `docs/guides/architecture.md`, including where to reach for raw pgx
      if a future flow needs it (`gdb.DB()`, not a second pool)

**Done when:** the process opens one pool per database.

### D6 Remove `ParseDecimal`, which silently zeroes bad input · S — ✅ done
**Evidence:** `pkg/utils/parse.go:5-8`

```go
func ParseDecimal(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}
```

Unparseable input becomes `0` with no signal. On a money value that is the worst possible
failure mode: an invoice line silently becomes free. The function currently has no callers, so
this is cheap to fix now and expensive to fix after someone reaches for it.

- [x] Function and file deleted
- [x] If a helper is wanted later it must return `(decimal.Decimal, error)`; callers can use
      `decimal.NewFromString` directly in the meantime

**Done when:** no helper in the tree converts an invalid numeric string into a valid zero.

### D7 Make the JSON casing of the response envelope consistent · S — ✅ done
**Evidence:** `pkg/response/format.go:16-20` vs `internal/delivery/http/dto/response/auth.go:26-34`

`Meta` serialises `request_id` in snake_case while every DTO in the API uses camelCase
(`accessToken`, `refreshTokenExpiresAt`, `lastLoginAt`) and so does `pkg/pagination`
(`totalPages`, `hasNext`). One envelope emits both conventions in a single response body.

- [x] camelCase: `request_id` → `requestId`
- [x] `TestEnvelope_EveryKeyIsCamelCase` decodes a full envelope and walks every object key at
      any depth, so the next field added cannot reintroduce the split. The rest of the new
      suite pins each helper's status code, success flag and default message — `pkg/response`
      was at 0%, which is how this drifted unnoticed
- [x] Convention stated in `docs/api/router.md`, with JWT claims and log attributes recorded as
      deliberate exceptions

**Done when:** every key the API emits uses one casing.

---

## P1 — Architecture and contracts

Nothing here is broken at runtime. Each item is a place where the code teaches a reader something
untrue — which, in a boilerplate whose product *is* the example, is the expensive kind of wrong.

### A1 Make the application layer's purpose unambiguous · S — ✅ done
**Evidence:** `docs/guides/architecture.md` (Application Layer), `CLAUDE.md:25,39-40,87-88`,
`internal/application/bar/service.go`, `internal/application/register/service.go`

`internal/application/` exists for **cross-domain orchestration**: flows that span several
domains and usually one transaction. `docs/guides/architecture.md:174` says exactly that, and
`application/register` is a correct instance of it — it coordinates `user`, `role` and `userrole`
inside a single `txManager.Do`, and owns the audit context for the whole unit of work.

Single-domain use cases living in their own `domain/<name>/usecase.go` is therefore intentional,
not misplaced. The layer is small because most flows do not need it. **Nothing needs to move.**

What is wrong is everything written *about* the layer:

- `CLAUDE.md:25` calls `application/` "Use-case implementations (app services)" and
  `CLAUDE.md:40` repeats "use-case implementations". Read alone — and it is what a new
  contributor reads first — that says every use case belongs here, which contradicts both the
  architecture guide and the code.
- `CLAUDE.md:87` instructs that adding a domain means creating `domain/<name>/`,
  `application/<name>/`, a repository and a handler. Under the real design, most new domains
  should have **no** `application/` package at all.

`application/bar` demonstrated the pattern badly. Its own comment said "handles multi-domain
orchestration" while it touched exactly one domain, so it showed the layer being used where it is
not needed. It also injected `barUC bar.Usecase`, never read that field, and called
`barRepo.CreateBar` directly. It was wired into the container and consumed by no handler.

It is deleted rather than repaired: `register` is a real, correct, cross-domain example, and a
fabricated second one that contradicts the rule is worse than none.

That last point exposed a real design gap rather than a typo. The guide's original example
orchestrated `Usecase` values, but both real services reach for repositories inside the
transaction, because `WithTx` is defined on `Repository` and not on `Usecase`. A use case called
inside `txManager.Do` would write on a different connection, outside the transaction, and survive
a rollback — so the documented shape was not achievable.

**Decision:** keep `WithTx` on `Repository` only. Adding transaction propagation to every
`Usecase` would push a persistence detail across every use-case interface in the project to serve
a handful of flows. Orchestration composes repositories inside the transaction; rules that must
be both shared and transactional move into a domain service.

- [x] Architecture guide rewritten: the fictional `CreateOrderService` is replaced by the real
      `register` service, with the transaction rule and its reasoning stated
- [x] `CLAUDE.md` corrected — the layer is described as cross-domain orchestration,
      `application/<name>/` is now conditional, and the `WithTx` rule is recorded
- [x] `internal/application/bar/` deleted, along with `BarSvc` and the two parameters
      `WireApplicationServices` no longer needed
- [x] The `*config.Config` dependency is gone from `register`. It turned out to be stored and
      never read, so no options struct was needed — the field and its constructor argument were
      simply removed

**Done when:** one description of the layer exists, its example follows it, and the example is
achievable with the interfaces the project actually has.

### A2 One pagination contract · M — ✅ done
**Evidence:** `pkg/response/format.go:31-38,183`, `pkg/pagination/paginator.go:27-50`,
`internal/delivery/http/handler/bar.go:134,153`, `.claude/rules/api-conventions.md`

There are **three** descriptions of a paginated response and no two agree:

| Source | Shape | Status |
|---|---|---|
| `.claude/rules/api-conventions.md` | `data: [...]` + `meta: {page, limit, total}` | not implemented anywhere |
| `response.Paginated` (`format.go:183`) | `data` + `pagination` + `meta` | no callers |
| `pagination.NewPaginatedResponse` | `data: {items, total, page, limit, totalPages, hasNext, hasPrev}` | what `bar` actually returns |

The reference handler makes it worse: `bar.go:134` documents the endpoint in Swagger as
`response.PaginatedResponse{data=[]dtoresponse.BarResponse}` while `bar.go:153` returns the
nested `pagination.PaginatedResponse` inside `response.Success`. **The published API documentation
does not describe the response the endpoint sends.**

Recommended shape, as the common REST convention and the one that keeps `data` a plain array:

```json
{ "success": true, "message": "...", "data": [ ... ],
  "meta": { "page": 1, "limit": 10, "total": 42, "totalPages": 5, "hasNext": true, "hasPrev": false } }
```

- [x] Implemented once in `pkg/response`; both losers deleted. `Pagination` is embedded in
      `Meta` **by pointer**, so its fields flatten into `meta` on a list response and are absent
      everywhere else. Embedding rather than tagging each field `omitempty` is what keeps `0`
      and `false` meaningful — `"total": 0` on an empty page has to survive
- [x] `data` is always an array: a nil slice is normalised to `[]`, so a client can iterate it
      without a nil check to tell "no results" from "no field"
- [x] `pagination.ParsePagination` and `PaginationRequest` kept; the response type removed
- [x] `bar` handler, its Swagger annotation, and `.swagger/` regenerated — the spec now
      describes what the endpoint actually sends
- [x] Pagination section of `.claude/rules/api-conventions.md` rewritten against the code
- [x] `TestPaginated_ExactShape` and `TestBarList_ReturnsTheDocumentedShape` assert the body
      byte for byte, plus counter tables and handler tests for paging pass-through and errors
- [x] **Found while rewriting the parser:** `limit` had no maximum, so `?limit=1000000` was a
      request for a million rows and the cost of refusing it fell on the database. Now capped
      at 100. Out-of-range values are clamped rather than rejected — a page past the end is a
      reasonable thing to ask for, and an empty page answers it better than a 400

**Done when:** one shape exists in code, docs and Swagger, and a test fails if they diverge.

### A3 Split the 588-line auth use case · M
**Evidence:** `internal/domain/auth/usecase.go`

The largest file in the repo carries login, refresh, logout, logout-all and password change in
one type. The package has already grown sensible collaborators around it — `session_service.go`,
`permission_service.go`, `device_service.go`, `user_validator.go`, `one_time_token.go` — so the
seams exist; the orchestrator simply never shrank.

At 588 LOC it is the file most likely to be copied as the model for a new domain, which
propagates the shape.

- [ ] Split by flow: sign-in, token lifecycle, credential management
- [ ] Keep the public `Usecase` interface unchanged so no caller moves
- [ ] Coverage for `internal/domain/auth` is 48.4%; the split should not lower it

**Done when:** no file in `domain/auth` exceeds ~250 LOC and the interface is untouched.

### A4 Decide what to do about the hand-rolled migrator · M
**Evidence:** `pkg/migration/migrator.go` (537 LOC, 0% covered), `CLAUDE.md`, `go.mod`

`CLAUDE.md` states the project uses golang-migrate. It does not — **golang-migrate is not in
`go.mod`**. Migrations run through 537 lines of in-house tooling in `pkg/migration/` that parses
files, maintains its own `migrations` table, takes a PostgreSQL advisory lock, and applies
statements in a transaction. It is thoughtfully written and completely untested (0.0% coverage),
and it uses stdlib `log` while the rest of the process uses `slog`.

Hand-rolled schema migration is the highest-blast-radius code in any backend: its failure mode is
a half-applied production schema.

- [ ] Either adopt golang-migrate (the file naming is already compatible) and delete the package,
      or keep it and give it real tests — dirty state, partial failure, concurrent runners, down
      migrations
- [ ] Either way, correct `CLAUDE.md`
- [ ] Switch it to the structured logger

**Done when:** the migration path is either a maintained dependency or covered by tests, and the
documentation names the one actually in use.

### A5 Give the two example domains one purpose each · S
**Evidence:** `internal/domain/foo/*` (23 `panic("Implement me")` across 5 files),
`internal/delivery/http/router/public.go:47`, `.claude/skills/crud-operations/SKILL.md:24`

`foo` is the blank template and `bar` is the worked example — a good pairing, and
`crud-operations/SKILL.md` names `ToFooFilter`, `ToFooResponse` and `ToFooListResponse` as the
symbols to copy. **This scaffolding should stay.**

The problem is only that `foo` is wired into live routes and into the gRPC registry
(`internal/wire/handler_grpc.go:14`). On a freshly cloned project, `POST /api/v1/foos` panics,
gets caught by the recover middleware, and returns 500 — a template shipped as a working
endpoint.

- [ ] Keep every `foo` file; stop registering its routes by default
- [ ] Either guard the registration behind a flag or comment the single `r.foo(v1)` line with a
      pointer to the skill
- [ ] Say plainly at the top of each `foo` file that it is a template

**Done when:** a fresh clone exposes no endpoint that panics, and the template is still there to
copy.

---

## P2 — Engineering hygiene

### H1 Make the test and lint gate meaningful · M — ✅ done
**Evidence:** `.github/workflows/ci-cd.yaml:64`, `Makefile:16-27`, `.golangci.yml`

- **No `-race`.** CI runs `go test -v ./...`. This codebase runs background jobs, an in-memory
  session store with a sweeper, an in-memory cache, and per-session rate limiting — all shared
  mutable state across goroutines. The race detector is the single highest-value flag available
  and it is not on.
- **Linter set is thin.** `.golangci.yml` enables errcheck, govet, staticcheck, ineffassign,
  misspell and unused. Missing, in rough order of value here: `gosec`, `errorlint`, `bodyclose`,
  `rowserrcheck`, `sqlclosecheck`, `noctx`, `contextcheck`, `copyloopvar`, `nilerr`.
- **No coverage signal.** Numbers are not tracked, so drift is invisible.

- [x] `go test -race -shuffle=on ./...` in CI and in `make test`
- [x] Linters added. They surfaced 28 findings once the default `max-same-issues: 3` cap was
      removed — that cap was hiding a third of them. All resolved: 11 `errorlint`, 14 `gosec`,
      2 `contextcheck`, 1 `noctx`. Each suppression carries a `#nosec` with its reasoning
- [x] `-coverprofile` in CI, summary written to the job summary
- [x] `make test` (race), `make test-quick` (no race), `make cover`, `make vuln`; `.PHONY`
      completed

Two things `-race` found that no other step would have:

- **A latent bug in the D2 fix.** `drainGRPC` called `GrpcServer.Stop()` in its timeout branch.
  grpc-go holds `s.mu` via a deferred unlock across `handlersWG.Wait()`
  (`server.go:1963-1986`, v1.80.0), so a concurrent `Stop()` blocks on that mutex for exactly as
  long as the stuck handler it was meant to rescue. The drain hung for the full budget. It now
  fires `Stop()` in a goroutine and returns.
- **A test passing for the wrong reason.** `TestDrainGRPC_BoundedByTheContext` waited 200ms for
  its RPC to land. Often it had not, so `GracefulStop` finished on its own and the test passed
  without exercising the timeout path at all. It now blocks on a signal from the handler.

- [ ] Local `golangci-lint` is 2.11.4 while CI pins v2.12.0 and `make lint-install` installs
      v2.12.0. Harmless today, but it is how "clean locally, red in CI" starts

**Done when:** a data race or an unchecked SQL error fails the build.

### H2 Close the supply-chain gaps in CI · M
**Evidence:** `.github/workflows/ci-cd.yaml`

- [x] `govulncheck` runs on every PR — added with H1, but **`continue-on-error: true`**
- [ ] **Upgrade, then make it blocking.** The first scan found **27 reachable vulnerabilities**
      (not merely present — reachable from this module's call graph). All have fixes:
  - 13 in the Go standard library: `go 1.26.2` → `1.26.6` covers `crypto/tls`, `crypto/x509`,
    `encoding/asn1`, `encoding/xml`, `html/template`, `net`, `net/http`, `net/mail`,
    `net/textproto`, `net/url`
  - `google.golang.org/grpc` v1.80.0 → v1.83.1 (3)
  - `github.com/gofiber/fiber/v2` v2.52.8 → v2.52.12 (2)
  - `github.com/go-viper/mapstructure/v2` v2.2.1 → v2.4.0 (2)
  - `golang.org/x/net` v0.52.0 → v0.55.0 (2)
  - `github.com/jackc/pgx/v5` v5.7.5 → v5.9.2, `golang.org/x/text` v0.35.0 → v0.39.0,
    `go.opentelemetry.io/otel` v1.43.0 → v1.44.0,
    `github.com/aws/aws-sdk-go-v2/service/s3` v1.89.2 → v1.97.3,
    `.../aws/protocol/eventstream` v1.7.3 → v1.7.8, `github.com/ClickHouse/ch-go` v0.61.5 → v0.65.0
  - Then drop `continue-on-error` from the workflow step
- [ ] Pin actions to commit SHAs, not floating major tags
- [x] Workflow-level least-privilege `permissions:` block (done with H1, same file)
- [x] `concurrency:` group so superseded pushes cancel, `main` exempt (done with H1, same file)
- [ ] Publish an SBOM and build provenance from `docker/build-push-action`
- [ ] Delete the ~150 lines of commented-out deploy jobs; recover them from git history when the
      GKE deploy is switched on

**Done when:** a known-vulnerable dependency fails CI, and the workflow contains no dead YAML.

### H3 Harden the production image · S
**Evidence:** `Dockerfile`

- [ ] Pin the runtime base by digest — `FROM alpine:latest` makes builds unreproducible; prefer
      `gcr.io/distroless/static` for a CGO-free binary
- [ ] Build with `-trimpath -ldflags="-s -w"` and stamp version/commit into the binary
- [ ] Align the builder with `go.mod`: the Dockerfile pins `golang:1.26-alpine`, `go.mod` says
      `go 1.26.2`, and `CLAUDE.md` still claims Go 1.24
- [ ] Translate the Indonesian comments — the rest of the codebase is in English
- [ ] Add a trailing newline to the file

**Done when:** two builds of one commit produce the same image, on a pinned base.

### H4 Test the HTTP delivery layer · L
**Evidence:** coverage run of 2026-09-22

| Package | Coverage |
|---|---|
| `internal/delivery/http/handler` | **0.6%** |
| `internal/delivery/http/router` | 0.0% |
| `internal/delivery/http/presenter` | 0.0% |
| `pkg/response` | 0.0% |
| `pkg/pagination` | 0.0% |
| `pkg/filesystem` | 0.0% |
| `internal/application/*` | 0.0% |

Middleware (63.9%), `domain/auth` (48.4%), `domain/job` (91.7%) and most of `pkg/` are covered
properly. The gap is concentrated in the layer that defines the public contract — which is
exactly where A2's three-way disagreement was able to develop unnoticed.

- [ ] Handler tests over `app.Test()`: status code and marshalled body for success, validation
      failure, not-found and server error on at least one full CRUD surface
- [ ] Table-driven tests for `pkg/response` covering every helper's envelope
- [ ] Tests for `pkg/pagination` boundaries: page 0, limit 0, limit above maximum, total 0
- [ ] `pkg/filesystem` against the local driver, with S3 behind its interface

**Done when:** the response contract is pinned by tests rather than by inspection.

### H5 One logger, everywhere · S
**Evidence:** `cmd/server/main.go`, `cmd/migrate/main.go`, `pkg/migration/migrator.go`

Three files bypass the structured logger: `main.go` writes startup and shutdown messages with
`fmt.Printf`, and the migrator uses stdlib `log`. In a container these lines land in the same
stream as the JSON logs and break any structured log pipeline parsing it.

`pkg/logger/slog.go` also imports `goilerplate/config`, so a general-purpose package depends on
this application's configuration struct. It should take a small options struct instead.

- [ ] Route every operational message through `slog`
- [ ] Invert the `pkg/logger` → `config` dependency

**Done when:** the process emits one log format on one stream.

### H6 Modernise for the Go version in use · S
**Evidence:** `go.mod:3` (`go 1.26.2`), 58 occurrences of `interface{}`

The module targets Go 1.26 while 58 sites still use `interface{}` against 74 using `any`. Mixed
within the same files, this reads as churn rather than intent.

- [ ] `gofmt -r 'interface{} -> any' -w` across the tree
- [ ] Reconcile the Go version stated in `CLAUDE.md` (1.24), `go.mod` (1.26.2) and the Dockerfile
- [ ] Review error handling for `errors.Is`/`errors.As` once `errorlint` is enabled (H1)

**Done when:** one spelling of the empty interface, one stated Go version.

---

## P3 — Feature completion

Carried over from the previous roadmap. Unchanged in substance; F4 now also covers the
hardcoded body limit found during this scan.

### F1 Email module: verification, forgot/reset password, email change · L
**Depends on:** the `one_time_tokens` table and session revocation (both delivered)

- [ ] `EmailSender` interface with `log` (default) and `smtp` drivers
- [ ] Config `auth.email.enabled`, `email.driver`, `email.smtp.*`, `frontend.base_url`
- [ ] Flows: email verification by 6-digit OTP; forgot password → emailed reset link → reset,
      revoking all sessions; email change re-authenticated with the current password, OTP sent to
      the **new** address, address swapped only after confirmation
- [ ] Invalidate previous tokens of the same type before issuing a new one
- [ ] Store OTP hashes as HMAC-SHA256 with a server secret — a plain SHA-256 of a 6-digit code is
      trivially brute-forced if the database leaks. High-entropy reset tokens may use SHA-256
- [ ] Verification looks up the latest active token by `(user_id, token_type)`, compares in
      constant time, and increments `attempts` on mismatch; invalidate after
      `auth.otp.max_attempts` (default 5). Looking up by hash alone cannot count wrong guesses
- [ ] Registration without enumeration: when the address is already registered, return the same
      success response and send a "you already have an account" email
- [ ] All token consumption through the atomic `ConsumeToken`
- [ ] `ForgotPassword` always returns 200
- [ ] `auth.require_email_verification` (default `false`), checked after the password
- [ ] Routes registered only when `auth.email.enabled`
- [ ] Completes the one deferred test from the previous roadmap: OTP invalidated after
      `max_attempts` — the repository already counts attempts, nothing yet enforces the ceiling

**Done when:** every flow works end to end with the `log` driver and no SMTP server.

### F2 Session management endpoints · M

- [ ] `GET /users/me/sessions` — device name and type, IP, `last_used_at`, current-session flag
- [ ] `DELETE /users/me/sessions/:id` — revoke one of the caller's own sessions, ownership enforced
- [ ] `POST /auth/logout-all` keeps the option to exclude the current session

**Done when:** a user can list their devices and sign out one specific other device.

### F3 Refresh token via httpOnly cookie · M
**Depends on:** F4 (CORS with credentials)

The refresh token is currently returned in the login JSON and presented on
`POST /api/v1/auth/refresh` as `Authorization: Bearer <refreshToken>`
(`internal/delivery/http/middleware/auth.go:78-112`). For browser clients that means JavaScript
must be able to read it, so any XSS anywhere in the page yields long-lived account access.

- [ ] Config `auth.refresh_transport: body | cookie` (default `body`, which suits mobile and
      service-to-service clients)
- [ ] In `cookie` mode: `HttpOnly`, `SameSite=Strict`, `Path=/api/v1/auth`, refresh token omitted
      from the JSON body
- [ ] `Secure` follows `app.env` rather than a separate switch — always on in production, off in
      development so the flow is testable over `http://localhost`. There is precedent for this at
      `internal/delivery/http/router/router.go:133`
- [ ] CSRF defence: check `Origin` against an allowlist **only when the header is present**.
      Browsers always send `Origin` on cross-site POSTs, so absence means a non-browser client —
      which is not a CSRF vector and must keep working for curl, Postman and server-to-server
- [ ] Document that `SameSite=Strict` requires the frontend and API to be same-site; a cross-site
      frontend needs `SameSite=None` plus a CSRF token

**Done when:** in cookie mode, browser JavaScript cannot read the refresh token, refresh still
works, and a non-browser client is unaffected.

### F4 CORS and request size limits · S

- [ ] Enable the CORS middleware — the config struct exists (`config/config.go:214-216`) and the
      middleware sits commented out at `internal/bootstrap/fiber.go:42-46`
- [ ] Config validation rejects `allow_origin: *` together with `allow_credentials: true`
- [ ] Replace the hardcoded `BodyLimit: 100 * 1024 * 1024` (`fiber.go:19`) with
      `server.body_limit`, default 10MB. 100MB is an unbounded-memory invitation on a service
      whose own `filesystem.max_file_size` is 200KB

**Done when:** CORS headers appear only when enabled, and an oversized request gets 413.

### F5 S3 storage improvements · M

- [ ] `use_path_style` for MinIO and Cloudflare R2
- [ ] Correct bucket parameter per provider (bucket-in-domain vs path style)
- [ ] `cdn_base_url`: serve through a CDN, keep the bucket private, set
      `Cache-Control: public, max-age=31536000, immutable`
- [ ] Presigned GET URLs (`presigned_url_expiry`, default 1h) when no CDN is configured
- [ ] Pass the multipart file size through to avoid the `HeadObject` per upload at
      `pkg/filesystem/s3.go:99` — which also currently discards its error
- [ ] Allow `.webp`; English validation messages with KB/MB formatting
- [ ] Every new field zero-value safe

**Done when:** upload and URL generation work for AWS S3, for MinIO/R2 path style, and behind a CDN.

### F6 Google login · M

- [ ] Config `auth.google.enabled`, `google.client_id`
- [ ] Verify Google ID tokens; login creates or links the user
- [ ] Link and unlink endpoints for authenticated users; linking requires the Google address to
      match the account address
- [ ] `google_id` column on `users`, nullable and unique

**Done when:** the Google routes do not exist when the feature is disabled.

---

## P4 — Cleanup

One commit's worth of deletions. Listed individually so the diff can be reviewed rather than
trusted. `golangci-lint` reports 0 issues on all of this: its `unused` linter only sees
unexported symbols, as `.golangci.yml` itself notes.

### C1 Delete the dead token generators · S
**Evidence:** `pkg/utils/generator.go`

Left behind when access and refresh tokens became JWTs. These names now describe something
`pkg/jwt` does, so they actively mislead.

- [ ] `GenerateRefreshToken`, `GenerateVerificationToken`, `TokenLength` — no callers
- [ ] `GenerateRandomString`, `GenerateRandomNumberString` — never called, and both `panic(err)`
      under a comment reading *"In real applications, handle the error appropriately"*. A
      boilerplate should not ship an example that documents itself as not production-ready
- [ ] **Keep `GenerateSecureToken`** — a correct `crypto/rand` primitive that F1 needs for
      high-entropy reset tokens

### C2 Delete the remaining unreferenced exports · S

| Symbol | Location | Why |
|---|---|---|
| `SuccessWithMeta`, `ErrorWithDetails` | `pkg/response/format.go:231,236` | duplicate `Success(…, WithMeta(…))` and `CustomError`; two ways to do one thing |
| `GetAuditInfo` | `pkg/auditctx/auditctx.go:44` | thin wrapper over `GetUserID` + `GetUserName`, both already used |
| `NewSlogWriterWithLevel` | `internal/bootstrap/database/database.go:81` | unused variant of `NewSlogWriter`; also returns an unexported type from an exported function |

**Keep** `filesystem.NewManager` (`pkg/filesystem/manager.go:16`): unused today, but it is the
injection seam for substituting `Storage` in tests, and H4 will use it.

### C3 Remove leftovers · S

- [ ] The commented-out CORS block at `internal/bootstrap/fiber.go:42-46` — superseded by F4
- [ ] `internal/domain/baz/` and `internal/application/product/` — empty directories, untracked by
      git, present only on disk
- [ ] `postgres: goilerplate.session.sql` in the repository root — an editor scratch file;
      untracked, so add the pattern to `.gitignore`
- [ ] Translate the Indonesian comment in `deploy/k8s/deployment.prod.yaml:73`

### C4 Resolve the config keys nothing reads · S

Three keys are documented but never consulted. Each needs a decision, not a blanket deletion:

- [ ] **`crypto.encryption_key`** — `pkg/crypto` (239 LOC, 84.4% covered, AES-256-GCM) has zero
      callers and the config key has zero readers. It is a working primitive; the dangling key is
      the problem. Either wire the key through or drop it from `config.yaml` and
      `docs/deployment/configuration.md:128`. A configured encryption key that encrypts nothing
      is a false assurance
- [ ] **`service.*` / `config.Services`** — read only in `internal/bootstrap/viper_test.go:28`,
      but documented in `config.full.example.yaml:218` as a deliberate extension point. **Keep**,
      and say so in the struct comment
- [ ] **`Service.Apikey`** (`config/config.go:410`) — part of that same extension point; keep it
      with `Services`

---

## Appendix A — Work already delivered

The previous roadmap covered the authentication flow and optional-component architecture. All 27
of its tasks landed between PRs #56 and #70; the detail lives in those PRs and in
[docs/guides/auth.md](../guides/auth.md).

| Area | Delivered |
|---|---|
| Optional components | Startup config validation, cache abstraction with null-object fallbacks, conditional wiring, idempotency with a memory fallback, framework-free domain layer, secret redaction in logs |
| Schema | PostgreSQL only (MySQL removed), UTC migration baseline, `user_sessions` schema, `one_time_tokens` table |
| Token security | JWT hardening with key rotation, stateless access tokens with a session check, refresh token rotation with reuse detection, transactional login, idempotent logout and logout-all, session revocation on password change and deactivation, atomic lockout with anti-enumeration |
| Other surfaces | Gateway-controlled `/internal`, hardened partner API keys, gRPC auth interceptor, retention-bounded cleanup job, permission-resolution de-duplication, trusted-proxy client IP and security headers, structured security audit events |
| Tests and docs | Auth use-case unit tests, HTTP-driven integration tests across all three cache modes, the auth guide, corrected configuration guide |

Decisions from that effort that still bind this plan:

| # | Decision |
|---|---|
| D1 | Server-side sessions always exist; how often they are checked is configurable |
| D2 | `auth.revocation` defaults to `strict`; simple projects may opt into `refresh_only` |
| D3 | PostgreSQL only |
| D4 | Fresh baseline — no backward compatibility with existing data or deployments. Migrations, config keys and API contracts may be rewritten freely |
| D5 | `/internal` shares the public port; exposure is controlled at the gateway |

D4 is what makes A2 (the pagination contract) and D7 (JSON casing) cheap: breaking the response
shape costs nothing today and a great deal once projects depend on it.

## Appendix B — How the scan was run

Reproduce any finding:

```bash
golangci-lint run                                   # 0 issues — the P4 findings are all exported
go test -cover ./...                                # coverage table in H4
go build ./... && go vet ./...                      # clean
grep -rn "ReadTimeout" --include="*.go" internal    # D1: no hit in bootstrap/fiber.go
grep -rn "PgxDB" --include="*.go" internal cmd      # D5: created, pinged, closed — never queried
```

Dead exported symbols were found by listing every `^func [A-Z]` and `^type [A-Z]` declaration and
counting references with comment lines and the declaration itself excluded. Counting them without
excluding doc comments under-reports, because a symbol's own doc comment references it by name.

**Coverage caveat:** `internal/infrastructure/repository`, `internal/infrastructure/cache` and
`internal/integration` skip without `POSTGRES_TEST_DSN` and `REDIS_TEST_ADDR`, so `make test`
alone can be green while they never ran. The H4 table was produced without them; use
`make test-integration` for real numbers. CI sets both.
