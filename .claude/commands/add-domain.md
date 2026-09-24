---
description: Scaffold a new Clean Architecture domain following the existing bar pattern
argument-hint: <name>
---

Scaffold a new Clean Architecture domain following the existing `bar` pattern. The user will pass the domain name as an argument: `/add-domain <name>` (e.g. `/add-domain product`).

Use the lowercase singular form for the package name (e.g. `product`) and PascalCase for type names (e.g. `Product`). Plural form is used only for the route group and table name (`/products`, `products`).

Follow these steps in order:

## 1. Validate inputs
- The argument must be a single lowercase word, ASCII letters only, no underscores or hyphens.
- If no argument is given, abort and ask the user for the domain name.
- Check that none of the target files already exist. If any do, abort and report which ones — do not overwrite.

## 2. Reference the canonical pattern
Before writing any file, read these `bar` files in full and mirror their structure. Where `bar` departs from `.claude/rules/*` or `.claude/skills/db-migrations/SKILL.md`, follow the rules, not `bar` — the known gaps are listed in phase R of `docs/roadmap/improvement-tasks.md`:

- `internal/domain/bar/{entity,usecase,repository,error,filter,message}.go`
- `internal/infrastructure/repository/bar.go`
- `internal/infrastructure/model/bar.go`
- `internal/delivery/http/handler/bar.go`
- `internal/delivery/http/dto/request/bar.go`
- `internal/delivery/http/dto/response/bar.go`
- `internal/delivery/http/presenter/bar.go`
- `internal/delivery/http/request/bar.go`
- `internal/delivery/http/router/internal.go` (route registration block)
- `internal/wire/{repository,usecase,handler}.go`

## 3. Create the new files
Substitute `bar` → `<name>` and `Bar` → `<Name>` throughout. Use sensible default fields (`ID`, `Code`, `Name`) — the user can adjust them after. Generate:

**Domain layer** (`internal/domain/<name>/`)
- `entity.go` — entity struct + `validate()` + `Clone()`
- `usecase.go` — `Usecase` interface + `usecase` struct + `NewUseCase` constructor + Create/Update/Delete/GetByID/GetList/BulkCreate methods
- `repository.go` — `Repository` interface with `WithTx` + CRUD + Count/List/GetByID
- `error.go` — `Err<Name>NotFound`, `ErrCodeAlreadyExists`, etc., using `utils.ClientErr`
- `filter.go` — `Filter` struct with `Keyword`, `Code`, `Pagination`
- `message.go` — `Msg<Name>CreatedSuccessfully`, etc.

**Infrastructure layer**
- `internal/infrastructure/model/<name>.go` — GORM model with audit columns (`CreatedAt`, `UpdatedAt`, `DeletedAt`, `CreatedBy`, `UpdatedBy`, `DeletedBy`, `IsActive`) and `TableName()` returning the plural snake_case form
- `internal/infrastructure/repository/<name>.go` — `<name>Repo` struct, `New<Name>` constructor, `WithTx`, full CRUD, `applyFilters` helper, `modelToEntity` helper

**Delivery layer**
- `internal/delivery/http/handler/<name>.go` — Fiber handler with Create/Update/Delete/Get/List, including Swagger annotations
- `internal/delivery/http/dto/request/<name>.go` — `<Name>CreateRequest`, `<Name>UpdateRequest`, `<Name>ListRequest`
- `internal/delivery/http/dto/response/<name>.go` — `<Name>Response`
- `internal/delivery/http/presenter/<name>.go` — `To<Name>Response`, `To<Name>ListResponse`
- `internal/delivery/http/request/<name>.go` — `To<Name>Filter` mapper

## 4. Wire it up
Edit these existing files to register the new domain (insert alphabetically between existing entries):

- `internal/wire/repository.go` — add field to `Repositories` struct + entry in `WireRepositories` returning `repository.New<Name>(db)`
- `internal/wire/usecase.go` — add field to `UseCases` struct + entry in `WireUseCases` returning `<name>.NewUseCase(repos.<Name>Repo)`
- `internal/wire/handler.go` — add field to `Handlers` struct + entry in `WireHandlers` returning `handler.New<Name>(app.Validator, useCases.<Name>UC)`
- `internal/delivery/http/router/internal.go` — add `r.<name>(internal)` to the `register` method, then add a `<name>(internal fiber.Router)` method that registers `POST /`, `PUT /:id`, `DELETE /:id`, `GET /`, `GET /:id` on the `<names>` group

## 5. Create migration
Run `make migrate-create name=create_<names>_table` to generate up/down SQL files. Write the up migration by the `db-migrations` skill: `UUID PRIMARY KEY` without a DB default (the repository sets it with `utils.GenerateUUID()`), `TEXT` columns, `TIMESTAMPTZ` audit columns, soft delete via `deleted_at`, and `code` unique **among live rows only** (`CREATE UNIQUE INDEX uq_<names>_code_live ON <names> (code) WHERE deleted_at IS NULL`) — a plain `UNIQUE` would stop a deleted row's code from ever being reused. Populate the down migration with `DROP TABLE IF EXISTS`.

## 6. Create tests
Scaffolded code ships with tests, like any other code (`.claude/rules/testing.md`):
- `internal/domain/<name>/usecase_test.go` — a hand-written fake `Repository` (embed the interface, implement what is called); cover create, the duplicate-code 409, not-found 404 on get/update/delete, and a repository error surfacing as a non-client error
- `internal/delivery/http/handler/<name>_test.go` — mirror `bar_test.go`: status codes and body shape for success, malformed JSON, validation failure, and each domain error
- `internal/infrastructure/repository/<name>_test.go` — against real PostgreSQL (skips without `POSTGRES_TEST_DSN`): create/get, soft delete hides the row, and a deleted row's code can be reused

## 7. Verify
- Run `go build ./...`, `make lint`, and `make test` — all must pass. Run `make test-integration` if PostgreSQL is available.
- Report a checklist of every file created/modified and any compile errors. Do not run migrations against the database — leave that to the user.

## Notes
- Skip the `application/<name>/` layer unless the user asks for cross-domain orchestration. Most simple CRUD domains don't need it.
- Do not register the route on `public.go` or `partner.go` — internal is the default. The user can move it manually if needed.
- Do not commit. Leave the changes staged or unstaged for the user to review.
