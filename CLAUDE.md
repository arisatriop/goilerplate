# Goilerplate — Claude Instructions

## Project Overview
Go backend boilerplate using Clean Architecture. Provides a ready-to-use foundation for REST APIs with auth, RBAC, file uploads, and PostgreSQL.

## Tech Stack
- **Language**: Go 1.26 (`go.mod` pins the exact toolchain; the Dockerfile builder matches it)
- **Router**: GoFiber v2
- **gRPC**: google.golang.org/grpc, proto contract at [goilerplate-proto](https://github.com/arisatriop/goilerplate-proto)
- **Database**: PostgreSQL only, via GORM + pgx
- **Cache**: Redis (go-redis/v9)
- **Config**: Viper (YAML — `config/config.yaml`), overridden by environment variables (dots become underscores: `jwt.access_secret` → `JWT_ACCESS_SECRET`)
- **Auth**: JWT (golang-jwt/jwt v5), access + refresh tokens
- **Migration**: in-house (`pkg/migration`), SQL files in `internal/migrations/`. Not
  golang-migrate — it has never been in `go.mod`. The migrator takes a PostgreSQL advisory lock,
  applies each file in one transaction, and hands the whole file to the server rather than
  splitting it, so dollar-quoted function bodies and comments work. A failed migration rolls back
  both the schema change and its bookkeeping row, so there is no dirty state to force
- **Decimal**: shopspring/decimal (use for all financial calculations — never float64)
- **Validation**: go-playground/validator v10
- **Storage**: AWS S3 (aws-sdk-go-v2)
- **DI**: Manual wire (`internal/wire/`)

## Project Structure
```
cmd/            Entry points (server, migrate, seed)
config/         YAML config (.env here is for the MCP servers, not the app)
internal/
  application/  Cross-domain orchestration services (only where a flow spans domains)
  bootstrap/    App initialization (Fiber, DB, Redis, gRPC, Viper)
  delivery/
    http/       HTTP handlers, middleware, router, DTOs
    grpc/       gRPC handlers, middleware, service registry
  domain/       Core domain: entities, interfaces, errors
  infrastructure/ GORM models, repository implementations, transactions
  migrations/   SQL migration files
  wire/         Dependency injection wiring
pkg/            Shared utilities (errors, response helpers, grpcclient, etc.)
storage/        Uploaded file storage
```

## Architecture Layers
- `domain/` — entities, domain errors, and the `Usecase` / `Repository` interfaces together with the use-case implementation for that domain
- `application/` — orchestration for flows that span **several** domains, owning the transaction boundary and the audit identity. Most domains never need one; see [docs/guides/architecture.md](docs/guides/architecture.md#-application-layer)
- `infrastructure/repository/` — GORM repository implementations
- `delivery/http/handler/` — Fiber handlers (depend on domain Usecase interface)
- `delivery/grpc/handler/` — gRPC handlers (depend on same domain Usecase interface)
- `wire/` — wires everything together

## gRPC
- Proto contract lives in a separate repo: [github.com/arisatriop/goilerplate-proto](https://github.com/arisatriop/goilerplate-proto)
- Server reflection is **off by default** (`grpc.reflection`) — clients import the proto module
- Calls are authenticated by `grpc.auth.mode` (default `token`, sharing HTTP's validator and session store)
- gRPC port: `50051` (configured in `config/config.yaml` under `grpc.port`)
- When adding a new gRPC service: add proto to goilerplate-proto → tag new version → `go get github.com/arisatriop/goilerplate-proto@<version>` → write handler → register → wire
- See [docs/guides/grpc.md](docs/guides/grpc.md) for full guide

## Development
```bash
make run              # run application via air (hot reload)
make test             # go test -race -shuffle=on ./... (no database needed)
make test-integration # the same, against a real PostgreSQL + Redis
make lint             # golangci-lint run
make migrate-up       # run pending migrations
make migrate-down     # rollback last migration
make migrate-create name=<name>  # create new migration files
```

Config file: `config/config.yaml` (copy from `config/config.example.yaml`)
Secrets: environment variables. **`config/.env` is not read by the application** — no dotenv
dependency exists; the file is sourced by the MCP servers in `.mcp.json`. To make it reach the
app: `set -a && . config/.env && set +a`. See [docs/deployment/configuration.md](docs/deployment/configuration.md).

The repository, cache and `internal/integration` suites need `POSTGRES_TEST_DSN` and
`REDIS_TEST_ADDR`. Without them they **skip**, so plain `make test` can be green while they
never ran — use `make test-integration` before trusting a result. CI sets both.

## Branching & Commit Convention
- Branches: `feat/<topic>`, `fix/<topic>`, `chore/<topic>`
- Commit format: `<type>(<scope>): <description>` (conventional commits)
  - e.g. `feat(auth): add refresh token rotation`
  - e.g. `fix(bar): correct pagination offset calculation`
- Types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`

## Important Rules
Detailed coding rules live in `.claude/rules/` (`code-style.md`, `api-conventions.md`, `testing.md`) and are always in effect — financial values, error handling, Clean Architecture boundaries, naming, and API conventions are defined there. They follow general Go and backend practice; where a rule and common practice disagree, practice wins and the rule is corrected in the same change. Where the code still falls short of a rule, the gap is tracked in phase R of `docs/roadmap/improvement-tasks.md` — follow the rule in new code, and don't copy the gap. Project-specific notes not covered by those rules:

- Secrets come from environment variables (`JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `DB_PASSWORD`, ...); never hardcode credentials and never put them in `config.yaml`. There is no `JWT_SECRET_KEY` — access and refresh tokens are signed with separate secrets
- Auth behaviour (rotation, revocation modes, cache-mode trade-offs, client contract) is documented in [docs/guides/auth.md](docs/guides/auth.md) — read it before changing anything under `domain/auth` or the auth middleware
- Baseline rows that the application cannot start without (e.g. the `owner` role registration assigns) belong in a migration, not a seeder
- When adding a new domain: create `domain/<name>/`, `infrastructure/repository/<name>.go`, `delivery/http/handler/<name>.go`, then wire it up in `internal/wire/`. Add `application/<name>/` **only** when the flow crosses domains — a single-domain flow goes from the handler straight to that domain's `Usecase`
- Inside `txManager.Do`, call `WithTx` repositories, never a `Usecase`: `WithTx` is defined on `Repository` only, so a use case called there would write outside the transaction and survive a rollback
- Migration files live in `internal/migrations/` — use `make migrate-create` to generate them.
  One file is one transaction, so a migration either lands whole or not at all; there is no
  partially-applied state to repair by hand

## Claude Commands
Project-level slash commands available:

| Command | Description |
|---|---|
| `/ship [TICKET_ID]` | **Full workflow**: fetch ticket → start comment → implement → commit & push → PR → done comment → transition |
| `/work-on [TICKET_ID]` | Fetch a Jira ticket's description and implement the work it describes |
| `/commit` | Create a conventional commit for current changes |
| `/commit-body` | Create a commit with subject + detailed body |
| `/push` | Push the current branch to GitHub (refuses if on `main`) |
| `/pr` | Create a GitHub pull request for the current branch |
| `/code-review` | Review current git diff for correctness, security, and conventions |
| `/pr-review` | Review an open GitHub PR using the `gh` CLI |
| `/add-domain <name>` | Scaffold a new Clean Architecture domain following the `bar` pattern |
| `/fix-issue <topic>` | Guided workflow to investigate and fix an issue |
| `/mark-as-start [TICKET_ID]` | Post a "starting work" comment to a Jira ticket |
| `/mark-as-done [TICKET_ID]` | Post a work-summary comment to a Jira ticket |
| `/next-transition [TICKET_ID]` | Move a Jira ticket to the next status via transition |

The Jira-aware commands share helper scripts in `.claude/scripts/` (`jira-lib.sh`,
`jira-tickets.sh`, `jira-ticket.sh`, `jira-comment.sh`, `jira-transition.sh`) so
credential handling and the Jira REST calls live in one place.

## MCP Setup
See [docs/reference/mcp-setup.md](docs/reference/mcp-setup.md) for MCP server configuration guide.
