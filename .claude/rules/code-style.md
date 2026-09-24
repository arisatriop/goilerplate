# Code Style Rules

The baseline is [Effective Go](https://go.dev/doc/effective_go),
[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) and the
[Google Go Style Guide](https://google.github.io/styleguide/go/). `gofmt`/`goimports` and
`golangci-lint` (`.golangci.yml`) enforce what can be checked mechanically; this file covers what
they cannot, and the choices specific to this project.

Where the code still falls short of a rule, the gap is listed in
`docs/roadmap/improvement-tasks.md` (phase R). Follow the rule for new code unless the rule says
otherwise.

## Naming
- MixedCaps everywhere, constants included: `DefaultServerReadTimeout`, never
  `DEFAULT_SERVER_READ_TIMEOUT`. No underscores in identifiers
- Name length follows scope. `i`, `r`, `ctx`, `err`, `tt` in a few-line scope are idiomatic; a
  package-level or long-lived name is descriptive. Receivers are one or two letters, the same on
  every method of a type
- Initialisms keep one case: `userID`, `apiKey`, `HTTPClient`, `parseURL` — never `userId`, `Url`
- Packages are short, lowercase, singular, and named for what they provide. Don't add new
  catch-all packages (`util`, `common`, `helpers`); `pkg/utils` exists, don't grow it
- No stutter: `bar.Usecase`, not `bar.BarUsecase`
- Files are lowercase with underscores between words (`usecase_signin.go`)

## Functions
- Keep the happy path unindented: handle errors and edge cases first and return early
- No hard line limit. When a function needs a comment per block to be followed, split it
- `context.Context` is the first parameter of anything that does I/O or may block, named `ctx`.
  **Never store a context in a struct** — it outlives the request it belonged to
- Accept interfaces, return concrete types. Interfaces belong to the package that consumes them
  and stay small — which is why `Repository` is declared in `domain/`, not `infrastructure/`
- Return errors; don't panic. `panic` is for programmer errors that cannot happen. Startup
  failures return an error to `main`, which logs it and exits non-zero
- Every goroutine has an owner that can stop it (context cancellation) and waits for it to exit
  (`sync.WaitGroup` or `errgroup`). No fire-and-forget goroutines. Never let a `*fiber.Ctx`
  escape into a goroutine: Fiber reuses it once the handler returns
- Make the zero value useful. Config fields that must not be zero get an `XxxOrDefault()` method
  rather than trusting every caller to remember a fallback

## Errors
- Handle each error once. Either handle it (degrade, retry, translate) or return it — never log
  **and** return. Logging happens at the boundary: `response.HandleError`, middleware, jobs, `main`
- Add context where it helps someone locate the failure: `fmt.Errorf("creating session: %w", err)`.
  Lowercase, no trailing punctuation, no `failed to` prefix (a chain would read
  `failed to x: failed to y: ...`). Return the error unchanged when the context adds nothing
- `%w` when callers may inspect the cause; `%v` when the cause is an implementation detail that
  must not become part of the API (for example a driver error crossing a layer)
- Compare errors with `errors.Is` / `errors.As` — never `==` on a possibly wrapped error, and never
  by message text
- Expected outcomes a client caused are declared once per domain in `domain/<name>/error.go`
  with `apperr.New(kind, code, message)`: the **kind** (`Invalid`, `Unauthenticated`, `Forbidden`,
  `NotFound`, `Conflict`) decides the transport status, the **code** is a stable snake_case
  identifier prefixed with the domain (`bar_not_found`), and the message is safe to show.
  Compare with `errors.Is` against the sentinel; attach a cause with `WithCause`
- The domain carries no transport knowledge: no `net/http`, no status codes. `pkg/response` and
  `pkg/grpcresponse` each map kinds in a single table
- Don't discard an error with `_` unless a comment says why it is safe
- Internal error text never reaches a client. `response.HandleError` logs it and sends a generic
  500

## Data types
- **Money and other exact quantities**: `github.com/shopspring/decimal`, never `float64`, on every
  step from request to database. Parse with `decimal.NewFromString` and handle its error. Store as
  `NUMERIC(p, s)`. On the wire, send it as a JSON string so JavaScript clients don't lose
  precision. Round explicitly, at a defined point, with a named rounding mode
- **Time**: `time.Time` in UTC, `TIMESTAMPTZ` in the database, RFC 3339 on the wire. Get "now" from
  `utils.Now` in code that tests need to control. Durations are `time.Duration`, including in
  new config keys (`15s`, not `15`)
- **IDs**: UUIDv7 from `utils.GenerateUUID()`, assigned by the application, never by a database
  default
- Domain entities (`domain/<name>/entity.go`) have no persistence or transport tags. GORM models
  live in `infrastructure/model/`. API shapes live in `delivery/http/dto/{request,response}/`.
  Never return a GORM model or a domain entity from a handler

## Architecture boundaries
Dependencies point inward.

| Package | May import | Must not import |
|---|---|---|
| `domain/` | stdlib, small pure helpers in `pkg/` | GORM, Fiber, gRPC, Redis clients, `infrastructure/`, `delivery/`, `config` |
| `application/` | `domain/` | `infrastructure/`, `delivery/` |
| `infrastructure/` | `domain/`, drivers | `delivery/`, `application/` |
| `delivery/` | `domain/` and `application/` interfaces, `pkg/` | `infrastructure/`, GORM |
| `pkg/` | stdlib, third-party | anything under `internal/`, `config` |
| `wire/` | everything | — it is the composition root |

- All database access goes through a repository interface
- Inside `txManager.Do`, call `WithTx` repositories, never a `Usecase` (see `CLAUDE.md`)

## Logging and security
- Log with `log/slog` through `pkg/logger`, as structured key/value attributes. Don't use `fmt.Print*`
  or the stdlib `log` package in application code. CLI output meant for a terminal is the
  exception
- Never log secrets, tokens, passwords, API keys, or full auth request bodies. `pkg/redact` and
  `log.omit_body_paths` exist for this. Security-relevant outcomes go through `logger.Security`
- SQL is always parameterised (`Where("code = ?", code)`). Never build SQL with `fmt.Sprintf`,
  and that includes `ORDER BY`: map a sort parameter to a column from a fixed allowlist
- Secrets come from environment variables, never from committed files
- Authorization is decided on the server from the authenticated subject. An ID in the request
  body or path never identifies the caller

## Comments
- Every exported identifier has a doc comment that begins with its name
- Comments explain why — the constraint, the trade-off, the incident. Say what the code does only
  when that isn't obvious from reading it
- Don't commit commented-out code. Delete it; git keeps it

## Dependencies
- Prefer the standard library and modules already in `go.mod`. A new module needs a stated reason
  in the PR: what it replaces, its maintenance and license, and its transitive weight (H2 found
  ClickHouse in the graph through an OTel plugin)
- `govulncheck` blocks CI. A dependency bump goes in its own commit so a regression can be
  bisected
