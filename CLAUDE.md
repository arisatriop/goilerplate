# Goilerplate — Claude Instructions

## Project Overview
Go backend boilerplate using Clean Architecture. Provides a ready-to-use foundation for REST APIs with auth, RBAC, file uploads, and multi-database support.

## Tech Stack
- **Language**: Go 1.24
- **Router**: GoFiber v2
- **gRPC**: google.golang.org/grpc, proto contract at [goilerplate-proto](https://github.com/arisatriop/goilerplate-proto)
- **Database**: PostgreSQL via GORM + pgx, MySQL via GORM
- **Cache**: Redis (go-redis/v9)
- **Config**: Viper (YAML — `config/config.yaml`) + `.env` for secrets
- **Auth**: JWT (golang-jwt/jwt v5), access + refresh tokens
- **Migration**: golang-migrate (SQL files in `internal/migrations/`)
- **Decimal**: shopspring/decimal (use for all financial calculations — never float64)
- **Validation**: go-playground/validator v10
- **Storage**: AWS S3 (aws-sdk-go-v2)
- **DI**: Manual wire (`internal/wire/`)

## Project Structure
```
cmd/            Entry points (server, migrate, seed)
config/         YAML config + .env secrets
internal/
  application/  Use-case implementations (app services)
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
- `domain/` — interfaces (Usecase, Repository) + entities + domain errors
- `application/` — use-case implementations (depend only on domain interfaces)
- `infrastructure/repository/` — GORM repository implementations
- `delivery/http/handler/` — Fiber handlers (depend on domain Usecase interface)
- `delivery/grpc/handler/` — gRPC handlers (depend on same domain Usecase interface)
- `wire/` — wires everything together

## gRPC
- Proto contract lives in a separate repo: [github.com/arisatriop/goilerplate-proto](https://github.com/arisatriop/goilerplate-proto)
- Server reflection is **disabled** — clients must import the proto module
- gRPC port: `50051` (configured in `config/config.yaml` under `grpc.port`)
- When adding a new gRPC service: add proto to goilerplate-proto → tag new version → `go get github.com/arisatriop/goilerplate-proto@<version>` → write handler → register → wire
- See [docs/guides/grpc.md](docs/guides/grpc.md) for full guide

## Development
```bash
make run              # run application via air (hot reload)
make test             # go test -v ./...
make lint             # golangci-lint run
make migrate-up       # run pending migrations
make migrate-down     # rollback last migration
make migrate-create name=<name>  # create new migration files
```

Config file: `config/config.yaml` (copy from `config/config.example.yaml`)
Secrets: `config/.env` (copy from `config/.env.example`)

## Branching & Commit Convention
- Branches: `feat/<topic>`, `fix/<topic>`, `chore/<topic>`
- Commit format: `<type>(<scope>): <description>` (conventional commits)
  - e.g. `feat(auth): add refresh token rotation`
  - e.g. `fix(bar): correct pagination offset calculation`
- Types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`

## Important Rules
Detailed coding rules live in `.claude/rules/` (`code-style.md`, `api-conventions.md`, `testing.md`) and are always in effect — financial values, error handling, Clean Architecture boundaries, naming, and API conventions are defined there. Project-specific notes not covered by those rules:

- Config secrets come from `config/.env`; never hardcode credentials
- When adding a new domain: create `domain/<name>/`, `application/<name>/`, `infrastructure/repository/<name>.go`, `delivery/http/handler/<name>.go`, then wire it up in `internal/wire/`
- Migration files live in `internal/migrations/` — use `make migrate-create` to generate them

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
