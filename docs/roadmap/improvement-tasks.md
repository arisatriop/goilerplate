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

| Phase | Theme | Tasks | Status |
|---|---|---|---|
| [P0](#p0--defects) | Defects — the code does not do what it says | D1 – D7 | ✅ complete |
| [P1](#p1--architecture-and-contracts) | Architecture and contracts | A1 – A5 | ✅ complete |
| [P2](#p2--engineering-hygiene) | Build, CI, supply chain, tests | H1 – H6 | ✅ complete |
| [P3](#p3--feature-completion) | Feature completion | F1 – F6 | F4 done |
| [P4](#p4--cleanup) | Dead code and drift | C1 – C4 | ✅ complete |
| [R](#r--alignment-with-the-rules) | Code that falls short of the rewritten `.claude/rules/*` | R1 – R10 | open · ~6–8 days |

**P0, P1, P2 and P4 are complete**, and F4 with them. What remains is the rest of P3 — feature
work rather than correction, and each item needs a scope decision before it starts.

F3 (refresh token via httpOnly cookie) lists F4 as a dependency, so it is now unblocked.

Original order: **D1 → D2 → H1 → D3–D7 → A1 → A2 → C1–C4 → H2–H6 → A3–A5 → P3**. H2 was
deferred and A3–A5 pulled forward; everything else was done in this order.

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

### A3 Split the 588-line auth use case · M — ✅ done
**Evidence:** `internal/domain/auth/usecase.go`

The largest file in the repo carries login, refresh, logout, logout-all and password change in
one type. The package has already grown sensible collaborators around it — `session_service.go`,
`permission_service.go`, `device_service.go`, `user_validator.go`, `one_time_token.go` — so the
seams exist; the orchestrator simply never shrank.

At 588 LOC it is the file most likely to be copied as the model for a new domain, which
propagates the shape.

- [x] Split by flow, all four files on the same `authUseCase` type:

      | File | LOC | Holds |
      |---|---|---|
      | `usecase.go` | 89 | the type, the `Usecase` interface, `NewUseCase`, `SessionExpiry` |
      | `usecase_signin.go` | 215 | `Login`, `Logout`, `LogoutAll`, menu/permission assembly |
      | `usecase_token.go` | 195 | `RefreshToken`, rotation, reuse detection, token minting |
      | `usecase_credentials.go` | 139 | `Register`, `ChangePassword`, `DeactivateUser` |

- [x] The `Usecase` interface is untouched, so no caller moved
- [x] **Every declaration moved byte for byte.** Verified mechanically: the file was parsed
      into its top-level declarations before the split and re-parsed from the four files after,
      then compared by name and by body — nothing missing, nothing added, no body changed
- [x] Coverage is still exactly 48.4%, which is the other half of the same proof
- [x] The package doc now describes the layout. `device_service.go` also carried one, so the
      two were being concatenated in `go doc`; the one-liner is gone

**Done when:** no file in `domain/auth` exceeds ~250 LOC and the interface is untouched.

### A4 Decide what to do about the hand-rolled migrator · M — ✅ done
**Evidence:** `pkg/migration/migrator.go`, `CLAUDE.md`, `go.mod`

**Decision: keep the in-house migrator, delete its SQL parser, and test it.**

The premise held up — `CLAUDE.md` claimed golang-migrate and golang-migrate is not in `go.mod`
— but two things in the original framing were wrong, and correcting them is what decided it.

**"0.0% covered" was a measurement artefact.** The package had three tests; they `t.Skip`
without `POSTGRES_TEST_DSN`, and the coverage run that produced the table did not set one. With
a database they ran and reported 53.8%. The migrator was under-tested, not untested.

**It had a real defect, and it was worse than "untested" suggested.** `splitSQLStatements` cut
each file on any line ending in a semicolon. Two consequences, both now pinned by tests:

- **A dollar-quoted body was torn apart** at the semicolons inside it. That is every PL/pgSQL
  function, including the `updated_at` trigger most schemas have. It fails loudly —
  `unterminated dollar-quoted string` — so the first person to write one would have hit it.
- **A line carrying an inline block comment was dropped whole.** `id int PRIMARY KEY, /* ... */`
  simply vanished, leaving *valid* SQL that creates the table without its primary key. Nothing
  errors, and the migration is recorded as applied. **This is the half-applied production schema
  that made A4 a P1 in the first place, and it was already in the code.** Confirmed against a
  real PostgreSQL: the table was created with `name` and without `id`.

**The fix deletes the parser rather than repairing it.** pgx forces the simple query protocol
whenever a query carries no arguments, and the simple protocol accepts several statements in one
message — so the whole file now goes to PostgreSQL in a single `Exec` and the server parses it.
Delegating to the only correct SQL parser available beats both a better splitter and a new
dependency. It needs the raw `*sql.DB` rather than `m.db.Transaction`, because GORM runs with
`PrepareStmt: true` and a prepared statement uses the extended protocol, which takes one
statement per message.

**Why not golang-migrate**, given the file naming is already compatible:

- The objection was never "hand-rolled", it was "unverified". Testing answers it at a fraction
  of the disruption, and the tests are worth having whichever tool runs underneath
- What remains after the parser is gone is small and does one thing: a PostgreSQL advisory lock,
  a `migrations` table, and one transaction per file
- **Transactional migrations mean there is no dirty state to force.** A failed run leaves neither
  the schema change nor a record of it, which is strictly better operationally than
  golang-migrate's `force` dance
- Adding a dependency to the highest-blast-radius path buys a different set of failure modes,
  not fewer. The door stays open: the file naming is still compatible

- [x] `splitSQLStatements` and `executeSQL` deleted (~65 lines); the file is handed to
      PostgreSQL whole
- [x] `context.Context` threaded through `Up` and `Down`. A migration that hangs no longer hangs
      the deploy forever: cancelling cancels the statement server-side, rolls the transaction
      back, and releases the advisory lock, so the next attempt is not locked out by the last
- [x] Tests for everything A4 asked for and the defects found on the way: dollar-quoted bodies,
      inline block comments, semicolons and comment markers inside string literals, partial
      failure rolling back both schema and bookkeeping, retry after a fix, down migrations,
      a failing down keeping its record, ordering, and stopping at the first failure
- [x] Coverage 53.8% → 67.2%, run against a real PostgreSQL
- [x] `CLAUDE.md` corrected — it names the in-house migrator and says why
- [x] Structured logger — done with H5

Two things found while writing the tests, both in the harness rather than the code, and both
worth recording because each produced a wrong answer before it was caught:

- The first version isolated tests with `SET search_path`. `SET` applies to one connection and
  GORM hands out a pool, so the migrator ran against `public` on whichever connection it drew
  and the harness reported a failure that was its own.
- The second version put `search_path` in the DSN by editing the string with `url.Parse`. That
  passed locally and **failed in CI**, because `POSTGRES_TEST_DSN` is a URL locally and libpq
  key/value in CI (`host=localhost port=5432 ...`). `url.Parse` does not fail on the key/value
  form — it quietly returns a path-shaped URL — so the mangled DSN surfaced only as a DNS lookup
  for a hostname containing the whole connection string.

Both are now avoided by building the connection through `pgx.ParseConfig`, which understands
both DSN shapes, and setting `search_path` as a runtime parameter. The helper also asserts
`current_schema()` before any test uses it, so a scoping failure can no longer masquerade as a
migrator failure.

**Done when:** the migration path is covered by tests and the documentation names the one
actually in use.

### A5 Give the two example domains one purpose each · S — ✅ done
**Evidence:** `internal/domain/foo/*` (23 `panic("Implement me")` across 5 files),
`internal/delivery/http/router/public.go:47`, `.claude/skills/crud-operations/SKILL.md:24`

`foo` is the blank template and `bar` is the worked example — a good pairing, and
`crud-operations/SKILL.md` names `ToFooFilter`, `ToFooResponse` and `ToFooListResponse` as the
symbols to copy. **This scaffolding should stay.**

The problem is only that `foo` is wired into live routes and into the gRPC registry
(`internal/wire/handler_grpc.go:14`). On a freshly cloned project, `POST /api/v1/foos` panics,
gets caught by the recover middleware, and returns 500 — a template shipped as a working
endpoint.

- [x] Every `foo` file kept; the single `r.foo(v1)` call is commented out with a pointer to
      `.claude/skills/crud-operations/SKILL.md`. A config flag was considered and rejected —
      adding a runtime switch for a template is more machinery than the problem justifies
- [x] The `foo` method itself is kept (that is the shape a new domain copies) with a `nolint`
      noting why it is uncalled
- [x] Each of the ten `foo` files opens with a banner saying it is a template, that `bar` is the
      worked example, and that the panics are deliberate
- [x] The gRPC side needed no change: `grpchandler.Foo` already returns `codes.Unimplemented`
      rather than panicking, which is the correct answer for an unimplemented method

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
  (`server.go:1966-1989`, v1.83.1), so a concurrent `Stop()` blocks on that mutex for exactly as
  long as the stuck handler it was meant to rescue. The drain hung for the full budget. It now
  fires `Stop()` in a goroutine and returns.
- **A test passing for the wrong reason.** `TestDrainGRPC_BoundedByTheContext` waited 200ms for
  its RPC to land. Often it had not, so `GracefulStop` finished on its own and the test passed
  without exercising the timeout path at all. It now blocks on a signal from the handler.

- [ ] Local `golangci-lint` is 2.11.4 while CI pins v2.12.0 and `make lint-install` installs
      v2.12.0. Harmless today, but it is how "clean locally, red in CI" starts

**Done when:** a data race or an unchecked SQL error fails the build.

### H2 Close the supply-chain gaps in CI · M — ✅ done
**Evidence:** `.github/workflows/ci-cd.yaml`

**27 reachable vulnerabilities → 0.** Reachable means reachable from this module's call graph,
not merely present in `go.sum`, so every one was worth acting on.

Done as one upgrade per commit, so a regression can be bisected to a single module:

| Step | Change | Closed |
|---|---|---|
| 1 | Go toolchain `1.26.2` → `1.26.8` | **13** |
| 2 | `google.golang.org/grpc` `v1.80.0` → `v1.83.1` | 3 |
| 3 | `fiber` `v2.52.12`, `x/text` `v0.39.0`, `x/net` `v0.56.0` | 5 |
| 4 | `jackc/pgx/v5` `v5.7.5` → `v5.9.2` | 1 |
| 5 | `mapstructure` `v2.4.0`, `grpc` `v1.83.2`, `clickhouse-go` `v2.48.0` | 4 |
| 6 | `aws-sdk-go-v2/service/s3` `v1.97.3`, `.../eventstream` `v1.7.8` | 2 |

- [x] **Half of them were the standard library.** One toolchain bump closed 13 without touching
      a dependency. `1.26.8` rather than the minimum `1.26.6`, so the whole patch series is
      taken once instead of four times
- [x] **The ClickHouse finding needed working out.** This project is PostgreSQL only, so why is
      `ch-go` in the graph at all? `internal/bootstrap/database` → `gorm.io/plugin/opentelemetry/tracing`
      → `gorm.io/driver/clickhouse` → `clickhouse-go/v2` → `ch-go`. The OTel GORM plugin depends
      on the ClickHouse driver, and it is reachable through package init. No connection is ever
      opened, but init code is code. Bumping `ch-go` alone does not compile — `v0.65.0` changes
      `compress.NewWriter`'s signature and `clickhouse-go v2.30.0` cannot build against it — so
      the consumer was upgraded instead, which resolves `ch-go` to `v0.74.0`, well past the fix
- [x] **The grpc upgrade was checked against D2, not assumed.** The bounded-drain fix reasons
      about grpc-go's `stop()` internals. Re-read in `v1.83.1`: structure unchanged, `s.mu.Lock()`
      with a deferred unlock still spans `handlersWG.Wait()`, so the fix still holds. Only the
      line numbers moved, and the citation in `cmd/server/main.go` was corrected
      (`server.go:1963-1986` → `1966-1989`). A comment pointing at the wrong lines is worse than
      no comment
- [x] **One upgrade surfaced a real security improvement.** `google.golang.org/api` v0.264.0
      deprecates `option.WithCredentialsFile` "because of a potential security risk" — it accepts
      any credential type without validating which. The Drive driver's call site is explicitly
      the service-account branch, so it now says so:
      `option.WithAuthCredentialsFile(option.ServiceAccount, path)`. The path comes from config,
      and an external-account configuration swapped in there can name an arbitrary URL as its
      token source
- [x] `continue-on-error` dropped — **govulncheck now blocks.** It ran advisory for exactly as
      long as it took to clear the backlog; a scanner allowed to stay red teaches everyone to
      ignore it. Pinned to `v1.8.0` rather than `@latest`, because a scanner that silently
      changes version makes "CI went red and nothing changed" impossible to reason about
- [x] Every action pinned to a commit SHA with the tag in a trailing comment. A tag is mutable:
      `actions/checkout@v4` is whatever that ref points at today, which is a supply-chain
      foothold in a workflow that holds `packages: write`
- [x] `sbom: true` and `provenance: mode=max` on the image build. The SBOM means the next
      advisory is answered by querying the registry rather than rebuilding and re-scanning;
      provenance signs which workflow and commit produced the image
- [x] The ~165 lines of commented-out GKE deploy jobs deleted, with a pointer to the `git log -S`
      that recovers them. Commented-out YAML cannot be linted or tested and drifts silently
      while looking like a plan
- [x] Workflow-level least-privilege `permissions:` and the `concurrency:` group (both with H1)

**Done when:** a known-vulnerable dependency fails CI, and the workflow contains no dead YAML.

### H3 Harden the production image · S — ✅ done
**Evidence:** `Dockerfile`

- [x] Runtime is `gcr.io/distroless/static-debian12:nonroot` **pinned by digest**. No shell, no
      package manager, no libc — nothing to patch and nothing for an attacker who reaches RCE to
      pivot with. `:nonroot` runs as uid 65532, so the `adduser` step is gone too
- [x] `-trimpath -ldflags="-s -w"`, plus `version`, `commit` and `buildDate` stamped in and
      logged as the process's first line. They default to `dev`/`unknown`, so a locally built
      binary is distinguishable from a released one rather than pretending to be `0.0.0`.
      `docker/build-push-action` passes them as build args
- [x] Builder pinned to `golang:1.26.2-alpine`, matching `go.mod` exactly. `CLAUDE.md`'s
      "Go 1.24" corrected
- [x] Indonesian comments translated; trailing newline added; `--mount=type=cache` for the
      module and build caches, which also keeps them out of the image layers
- [x] **Found while doing this:** the `build` job runs `if: github.event_name == 'push'`, so
      nothing ever exercised the Dockerfile on a pull request — a change that broke the image
      was found on `main`. A `docker-build` job now builds it without pushing on every PR

**Done when:** two builds of one commit produce the same image, on a pinned base.

### H4 Test the HTTP delivery layer · L — ✅ done
**Evidence:** coverage run of 2026-09-22

The gap was concentrated in the layer that defines the public contract — which is exactly where
A2's three-way disagreement was able to develop unnoticed.

| Package | Before | After |
|---|---|---|
| `internal/delivery/http/handler` | 0.6% | **29.0%** |
| `internal/delivery/http/router` | 0.0% | 23.8% |
| `internal/delivery/http/presenter` | 0.0% | 31.2% |
| `pkg/response` | 0.0% | 55.4% |
| `pkg/pagination` | 0.0% | 52.5% |
| `pkg/filesystem` | 0.0% | 28.8% |
| `internal/application/register` | 0.0% | 81.5% |

The handler and router numbers stay moderate on purpose: both packages are mostly `foo`
(a template, deliberately untested) and wiring that only a running process exercises.

- [x] Handler tests over `app.Test()` across a full CRUD surface — status code and marshalled
      body for success, malformed JSON, each validation rule, and every domain error. The
      status-code table is the valuable one: it proves a duplicate code is a 409 and a missing
      row a 404 rather than both flattening to 500
- [x] `pkg/response` — every helper's envelope, default vs. given message, the exact list shape,
      and `TestEnvelope_EveryKeyIsCamelCase` (delivered with D7 and A2)
- [x] `pkg/pagination` boundaries: page 0, negative page, limit 0, non-numeric, limit above the
      maximum, and `limit=0` into `NewPagination`, which would have divided by zero
- [x] `pkg/filesystem` — the local driver end to end on a temp directory, plus `NewManager` with
      a spy `Storage`, which is what that seam was kept for (C2)
- [x] Presenter mapping, including that an empty result is an empty slice and not nil
- [x] **Found while writing these:** `validateUpload` reported the size limit as
      `fmt.Sprintf("%.0fmb", convertToMB(max)-1)`. The `-1` subtracts a whole megabyte, so the
      default 200KB limit produced *"Ukuran file melebihi batas maksimum -1mb"* — wrong number,
      wrong unit, wrong language. Replaced with `humanSize`, which picks bytes/KB/MB

**Done when:** the response contract is pinned by tests rather than by inspection.

### H5 One logger, everywhere · S — ✅ done
**Evidence:** `cmd/server/main.go`, `cmd/migrate/main.go`, `pkg/migration/migrator.go`

Three files bypass the structured logger: `main.go` writes startup and shutdown messages with
`fmt.Printf`, and the migrator uses stdlib `log`. In a container these lines land in the same
stream as the JSON logs and break any structured log pipeline parsing it.

`pkg/logger/slog.go` also imports `goilerplate/config`, so a general-purpose package depends on
this application's configuration struct. It should take a small options struct instead.

- [x] `pkg/migration` and `cmd/migrate` moved off stdlib `log`. The migrator takes an
      `*slog.Logger` (nil falls back to `slog.Default()`, so a test can pass one argument), and
      `log.Fatal` is replaced by a structured error plus an explicit exit code
- [x] **Not everything moved, on purpose.** The `create` output and the status table are a CLI
      result a developer reads at a terminal and may redirect; those stay on stdout, and the
      usage text moved to stderr so `migrate -action=status > report.txt` captures the table
      rather than the help. Each carries a comment saying why
- [x] `pkg/logger` → `config` inverted: `logger.New(logger.Options{...})` takes a plain struct,
      and `internal/bootstrap` does the translation. **No package under `pkg/` imports `config`
      any more**
- [x] `cmd/server` was already clean — its `fmt.Printf` calls went with D2

**Done when:** the process emits one log format on one stream.

### H6 Modernise for the Go version in use · S — ✅ done
**Evidence:** `go.mod:3` (`go 1.26.2`), 58 occurrences of `interface{}`

The module targets Go 1.26 while 58 sites still use `interface{}` against 74 using `any`. Mixed
within the same files, this reads as churn rather than intent.

- [x] `gofmt -r 'interface{} -> any' -w` across the tree — 43 sites in 8 files, now 0
- [x] Go version reconciled: `go.mod` is the source of truth (`1.26.2`), the Dockerfile builder
      pins the same patch, CI already used `go-version-file: go.mod`, and `CLAUDE.md` no longer
      claims 1.24
- [x] `errorlint` review landed with H1 — eleven wrapping-unsafe comparisons fixed

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

### F4 CORS and request size limits · S — ✅ done

- [x] **CORS middleware enabled**, off by default. The config struct existed and nothing read
      it; the middleware sat commented out in `internal/bootstrap/fiber.go`. `CORS` also gained
      `allow_credentials`, `expose_headers` and `max_age` — without `expose_headers` a browser
      sees only the CORS-safelisted response headers whatever the server sent
- [x] `allow_methods` defaults to `GET,POST,PUT,PATCH,DELETE,OPTIONS` rather than Fiber's own
      default, **which omits PATCH**. This API has PATCH routes, and the failure would have been
      a preflight rejection with nothing in the config file to explain it
- [x] **Validation rejects `allow_origin: *` with `allow_credentials: true`.** This turned out
      to matter more than it reads: `cors.New` does not return an error on a bad configuration,
      it **panics** — so without the check the operator gets a stack trace out of a middleware
      constructor instead of a message naming the key. The same applies to an origin without a
      scheme (`app.example.com` is the natural thing to write, and it panics), so that is
      checked too, along with trailing slashes, empty list entries and a negative `max_age`
- [x] `allow_origin: *` is also refused in production outright, and `allow_origin` is required
      once `enable_cors` is on
- [x] **`BodyLimit` now comes from `server.body_limit`, default 10MB.** It was hardcoded at
      100MB on a service whose own `filesystem.max_file_size` defaults to 200KB — Fiber buffers
      the body before a handler runs, so upload validation never got the chance to refuse
      anything. A negative value is treated as unset rather than passed through, because Fiber
      reads a negative limit as *no* limit
- [x] Tests: CORS headers absent when disabled, a configured origin echoed, an unlisted origin
      refused, PATCH allowed at preflight; body limit from config, falling back on 0 and on
      negative, and **413 asserted over a real socket** — `app.Test` reports the refusal as a Go
      error, while a real client sees the status code it has to act on
- [x] Validation tests for every case above, plus six configurations that must keep working
      (several origins, wildcard subdomains, credentials with a named origin, `localhost` over
      http)
- [x] Both example configs document every new key; `TestConfigExample_FullDocumentsEveryOption`
      enforces that

Confirmed against a live server rather than only in tests:

```
preflight, allowed origin     204  ACAO="https://app.example.com" ACAC="true" ACMA="600"
preflight, unlisted origin    204  ACAO=""
POST within body limit        200  body="handler ran"
POST over body limit          413  handler never ran
```

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

### C1 Delete the dead token generators · S — ✅ done
**Evidence:** `pkg/utils/generator.go`

Left behind when access and refresh tokens became JWTs. These names now describe something
`pkg/jwt` does, so they actively mislead.

- [x] `GenerateRefreshToken`, `GenerateVerificationToken`, `TokenLength` deleted
- [x] `GenerateRandomString`, `GenerateRandomNumberString` deleted
- [x] `GenerateSecureToken` kept, and now has tests plus a doc comment saying what it is for
      (high-entropy reset and invite tokens) and what it is **not** for (OTPs, which need HMAC
      because a plain digest of a 6-digit code is brute-forced in seconds if the database leaks)

### C2 Delete the remaining unreferenced exports · S — ✅ done

| Symbol | Location | Why |
|---|---|---|
| `SuccessWithMeta`, `ErrorWithDetails` | `pkg/response/format.go:231,236` | duplicate `Success(…, WithMeta(…))` and `CustomError`; two ways to do one thing |
| `GetAuditInfo` | `pkg/auditctx/auditctx.go:44` | thin wrapper over `GetUserID` + `GetUserName`, both already used |
| `NewSlogWriterWithLevel` | `internal/bootstrap/database/database.go:81` | unused variant of `NewSlogWriter`; also returns an unexported type from an exported function |

All four deleted, including `AuditInfo`, which existed only as `GetAuditInfo`'s return type.
`NewSlogWriterWithLevel` went with the D5 commit, in a file that change was already editing.

**Kept** `filesystem.NewManager` (`pkg/filesystem/manager.go:16`): unused today, but it is the
injection seam for substituting `Storage` in tests, and H4 will use it.

### C3 Remove leftovers · S — ✅ done

- [x] The commented-out CORS block at `internal/bootstrap/fiber.go:42-46` replaced with a
      sentence naming F4 and the one rule a copied snippet gets wrong (`AllowOrigins: "*"` with
      `AllowCredentials: true`). Commented-out code that nobody can run is not guidance
- [x] `internal/domain/baz/` and `internal/application/product/` removed from disk
- [x] `postgres: goilerplate.session.sql` deleted. `.gitignore` already had `*.session.sql`,
      which is why it never showed up as untracked
- [x] Indonesian comments in all three `deploy/k8s/deployment.*.yaml` translated (done with D4,
      in files that change was already editing)

### C4 Resolve the config keys nothing reads · S — ✅ done

Three keys are documented but never consulted. Each got a decision rather than a blanket
deletion:

- [x] **`crypto.encryption_key`** — **kept, and the assurance made honest.** `pkg/crypto`'s API
      takes the key per call (`EncryptString(plaintext, key)`), so there is no wiring to do: the
      config key *is* the seam, and it works the moment something encrypts a column. Deleting it
      would only mean re-adding the plumbing later. The struct, `config.full.example.yaml` and
      `docs/deployment/configuration.md` now all say plainly that nothing reads it yet.
      Also documented while reading the code: the key is derived with **one unsalted SHA-256**,
      so it must be generated (`openssl rand -base64 32`) and not chosen — that derivation does
      not stretch a memorable passphrase
- [x] **`service.*` / `config.Services`** — kept, with the extension point explained in the
      struct comment: the map is keyed by a name of your choosing so a deployment adds an
      upstream in config rather than in a new struct, and an empty `service:` block is correct
- [x] **`Service.Apikey`** — kept, with a comment distinguishing it from the inbound `api_key`
      map: this one is *sent to* an upstream, that one is *accepted from* a partner

---

## R — Alignment with the rules

On 2026-09-24 `.claude/rules/*`, the `go-reviewer` agent, the `db-migrations` and `crud-operations`
skills and `/add-domain` were rewritten against general Go and backend practice, not against the
code as it stood. Each rule the code does not yet meet is listed here, so a rule never
silently describes behaviour that does not exist — the failure A2 was about. New code follows the
rule; these tasks bring the old code up to it.

**Order:** R2 → R5 → R6 first (defects, each S), then the rest by value.

### R2 Stop echoing submitted values in validation errors · S — **security**
**Evidence:** `pkg/response/errors.go:27-29`, `internal/delivery/http/middleware/request.go:24,160-170`,
`pkg/redact/redact.go:18-27`

`FormatValidationErrors` puts `fieldError.Value()` in every detail. Send
`PUT /api/v1/users/me/password` with a 7-character `newPassword` and the 400 **returns the
password in the body**. That path is not under the default `omit_body_paths` (`/api/v1/auth`),
the logger records response bodies, and `value` is not a redacted key — so the password is
also **written to the log in plaintext**. A short password that fails our policy is often one the
user really uses elsewhere.

- [ ] Drop `Value` from `ValidationErrorDetail`
- [ ] Report the **JSON** field name via `validator.RegisterTagNameFunc`; today it is
      `strings.ToLower(StructField)`, so clients see `newpassword`, matching neither the request
      key nor the camelCase contract
- [ ] Add `/api/v1/users/me/password` — or better, every route that accepts a credential — to the
      default omit list, and test that a failed password change logs no body
- [ ] Test: a validation 400 contains no submitted value, and field names equal the JSON keys

### R5 Finish A5: `foo` is still routed on `/internal` and `/partner` · S — defect
**Evidence:** `internal/delivery/http/router/internal.go:13`, `internal/delivery/http/router/partner.go:19`

A5 unrouted `foo` from `public.go` only. `POST /internal/foos` and `POST /partner/v1/foos` still
reach `panic("Implement me")` and answer 500, so A5's "done when" does not hold.

- [ ] Comment out `r.foo(...)` in both, with the same pointer to the template docs
- [ ] A router test that walks every registered route of a fresh app and asserts none reaches
      `foo`, so the next audience file cannot reintroduce it

### R6 Make `bar` uniqueness correct under soft delete and concurrency · S — defect
**Evidence:** `internal/migrations/*_create_bars_table.up.sql`, `internal/domain/bar/usecase.go:35-63`,
`internal/infrastructure/repository/bar.go:106`

- `code` is `UNIQUE` across **all** rows, but `ExistsByCode` only looks at live ones. Deleting a bar
  and creating another with the same code passes the check and then violates the constraint:
  **500 instead of 201**
- The check-then-insert is racy: two concurrent creates both pass `ExistsByCode`, and the loser
  gets a 500 instead of a 409
- `ExistsByCode` loads a list through `GetBarList` to answer a yes/no question
- `idx_bars_code` duplicates the index `UNIQUE` already creates; `idx_bars_is_active` and
  `idx_bars_deleted_at` index low-cardinality columns alone

- [ ] Replace `UNIQUE` with `uq_bars_code_live ... WHERE deleted_at IS NULL`; drop the redundant
      indexes (D4 allows rewriting the baseline)
- [ ] Map PostgreSQL `unique_violation` (SQLSTATE 23505) to the domain's conflict error in the
      repository, so the constraint — not the pre-check — is what guarantees the 409
- [ ] Repository tests: a deleted code can be reused; concurrent creates yield one 201 and one 409

### R7 Give the worked example the tests it tells others to write · M
**Evidence:** `internal/domain/bar/` and `internal/infrastructure/repository/` have no `bar` tests

`bar` is what `/add-domain` copies, and the testing rules rank use-case tests first — yet
`bar` has handler and presenter tests only. The scaffold now generates use-case and repository
tests; the example it copies from should have them too.

- [ ] `domain/bar/usecase_test.go` with a hand-written fake repository
- [ ] `infrastructure/repository/bar_test.go` against real PostgreSQL (lands with R6's cases)

### R1 Take HTTP out of the domain layer · M
**Evidence:** `net/http` imported by five files in `internal/domain/auth/`; `internal/domain/bar/error.go:7`
(`utils.ClientErr(409, ...)`)

Domain errors are built with HTTP status codes, so the domain knows its transport. The gRPC side
already has to translate them back (`pkg/grpcresponse/errors.go`). The rule says domain errors
carry meaning, delivery maps them.

- [ ] A small error kind in the domain (`NotFound`, `Conflict`, `Invalid`, `Unauthenticated`,
      `Forbidden`) carrying a safe message; no `net/http`
- [ ] `response.HandleError` and the gRPC mapper translate kinds to status codes in one table each
- [ ] Migrate `auth` and `bar`; delete `utils.ClientErr` once nothing uses it

### R3 Decide on a machine-readable error contract · M — decision first
Clients can branch only on the status code; `message` is prose and may change. Two options:

- a stable `code` field in the existing envelope (`"code": "session_not_found"`) — smallest change
- RFC 9457 Problem Details (`application/problem+json`) for errors — the standard, and what
  gateways and client libraries increasingly understand, but a second response shape

Either pairs naturally with R1's error kinds. Needs a decision before code.

### R4 `Location` on 201 and `Retry-After` on 429 · S
**Evidence:** `internal/delivery/http/handler/bar.go:59`, `internal/delivery/http/middleware/ratelimit.go:48,68`

- [ ] `response.Created` takes the new resource's path and sets `Location`; `bar` returns the
      created resource (or at least its ID) rather than `nil`
- [ ] The limiters set `Retry-After` from the window's remaining time, so clients back off
      instead of hammering

### R8 Stop storing contexts in structs · S — fold into F5
**Evidence:** `pkg/filesystem/s3.go:24`, `pkg/filesystem/drive.go:20`

Both drivers keep the context they were constructed with and use it for every call, so an upload
is neither cancelled when its request is nor traced under it.

- [ ] Thread `ctx` through the `Storage` interface methods

### R9 Return startup errors instead of panicking · S
**Evidence:** `internal/wire/infrastructure.go:43,50,92`, `internal/bootstrap/viper.go:32,43`

- [ ] Wiring returns `error` up to `main`, which logs it once with the structured logger and exits
      non-zero — instead of a panic trace on stderr that bypasses the log pipeline H5 unified

### R10 Validate the incoming `X-Request-ID` · S
**Evidence:** `internal/delivery/http/middleware/request.go:70`

The header is taken verbatim from the client, so any length and any content lands in every log
line of the request and is echoed back.

- [ ] Accept it only when it is a bounded, safe token (for example ≤ 128 chars of
      `[A-Za-z0-9._-]`); otherwise generate one

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
