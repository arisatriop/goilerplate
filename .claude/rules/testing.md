# Testing Rules

## Running
- `make test` — `go test -race -shuffle=on ./...`. No database or Redis needed
- `make test-integration` — the same run against a real PostgreSQL and Redis
- `make cover` — coverage with the integration suites included

The repository, cache and `internal/integration` suites **skip** without `POSTGRES_TEST_DSN` and
`REDIS_TEST_ADDR`. A green `make test` says nothing about them. Run `make test-integration`
before trusting a change that touches SQL, caching, or auth. CI sets both.

## What goes where

| Code under test | How | Why |
|---|---|---|
| Domain use cases and services | Unit tests with hand-written fakes of the repository interface | Fast and deterministic, and able to reach every branch |
| Repositories | Real PostgreSQL | Constraints, locking, `NULL`s and time zones are what is being tested. A mocked driver would only prove the mock |
| HTTP handlers | `app.Test()` with a stub use case. Assert the status **and** the body shape | The handler's job is the HTTP contract |
| Behaviour that emerges only when layers agree (revocation, auth lifecycle, caching) | `internal/integration/`, over HTTP, with the real middleware, use case, repository and Redis | A stub would only prove its own semantics |
| External services (S3, SMTP, payment gateways) | A fake behind the interface the code already depends on | Never call a real third party from a test |

- **Fakes:** embed the interface and implement only the methods the test reaches (see
  `internal/domain/auth/usecase_refresh_test.go`). An unexpected call then panics loudly. Prefer
  this over generated mocks. Add a mock generator only if hand-written fakes become repetitive
- **Never mock SQL.** No sqlmock and no GORM mock. Use a real database for anything below the
  repository interface

## Writing tests
- **Table-driven** with `t.Run` subtests when cases share a shape. Name each case for the
  behaviour ("expired session is 401"), not the input
- **Names:** `TestType_Method`, or `TestFunction`, plus `_Scenario` when one test pins one behaviour
- **Structure:** Arrange/Act/Assert comments are optional. Use them when the phases are not obvious
- `t.Helper()` in helpers. `t.Cleanup` rather than `defer` for anything a helper creates.
  `t.Context()` for contexts that should end with the test
- `t.Parallel()` for pure unit tests. Leave it off tests that swap globals (`slog.SetDefault`) or
  share database state
- **No `time.Sleep` to wait for something.** Synchronise on a channel, or poll with a deadline
  (`assert.Eventually`). Sleeps make tests slow and flaky, and H1 caught one passing for the
  wrong reason
- Inject time (`utils.Now`) and randomness rather than asserting around them
- Assert observable behaviour: the status and body, the rows written, the events emitted. Assert
  that a method was called only when the call **is** the behaviour (for example, a cache eviction)
- `require` for preconditions whose failure makes the rest meaningless. `assert` for the checks
- Black-box tests (`package x_test`) by default. Use same-package tests when the logic under
  test is unexported
- Every test must pass under `-race` and `-shuffle=on`. Tests must not depend on execution order or
  on data another test left behind. Create what you need, clean it up with `t.Cleanup`, and
  **close every pool or connection you open**. The integration harness once exhausted
  PostgreSQL's connection limit this way

## Coverage
- New logic ships with its happy path and every error path it introduces
- Security-relevant code (auth, permissions, ownership, input validation) also needs the hostile
  case: someone else's resource, an expired or revoked token, malformed input
- A bug fix ships with a test that fails without the fix
- Coverage is a signal, not a target. Don't write tests for wiring that only a running process
  exercises, or for the `foo` template

## Financial calculations
Dedicated unit tests, always covering:
- zero, negative, and very large magnitudes
- values at the maximum scale, and rounding exactly at the boundary (`x.xx5`)
- precision cases that break floats (`0.1 + 0.2`)
- division by zero, and inputs whose scale differs from the currency's
