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
- Never use `float64` for financial values — always use `shopspring/decimal`
- Config secrets come from `config/.env`; never hardcode credentials
- All DB access goes through the repository interface — no GORM calls in handlers or use cases
- Follow Clean Architecture boundaries: domain ← application ← delivery; never import inward layers outward
- When adding a new domain: create `domain/<name>/`, `application/<name>/`, `infrastructure/repository/<name>.go`, `delivery/http/handler/<name>.go`, then wire it up in `internal/wire/`
- Migration files live in `internal/migrations/` — use `make migrate-create` to generate them

## Claude Commands
Project-level slash commands available:

| Command | Description |
|---|---|
| `/commit` | Create a conventional commit for current changes |
| `/commit-body` | Create a commit with subject + detailed body |
| `/code-review` | Review current git diff for correctness, security, and conventions |
| `/pr-review` | Review an open GitHub PR using the GitHub MCP server |
| `/fix-issue <topic>` | Guided workflow to investigate and fix an issue |

## MCP Setup
See [docs/reference/mcp-setup.md](docs/reference/mcp-setup.md) for MCP server configuration guide.
