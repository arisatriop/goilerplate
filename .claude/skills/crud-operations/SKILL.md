---
name: crud-operations
description: How to build or modify a CRUD feature across the Clean Architecture layers in this project. Use when manually editing domain / infrastructure / delivery files for an entity. For scaffolding a whole new domain from scratch, prefer the /add-domain command instead.
---

# CRUD Operations

How a CRUD feature is wired across the layers. Full step-by-step reference with
signatures: `docs/guides/crud-operations.md`. For a brand-new domain, the
`/add-domain` command scaffolds all of this at once — this skill is for **manual**
edits to an existing or partially-built domain.

Replace `Foo` with the entity name (PascalCase types, lowercase files/folders).

## The layered file set

A complete CRUD entity touches these files:

| Layer | Files |
|---|---|
| Domain (`internal/domain/foo/`) | `entity.go`, `repository.go`, `usecase.go`, optionally `error.go`, `filter.go`, `message.go` |
| Infrastructure | `infrastructure/model/foo.go` (GORM model), `infrastructure/repository/foo.go` (repo impl) |
| Delivery — DTOs | `delivery/http/dto/request/foo.go`, `delivery/http/dto/response/foo.go` |
| Delivery — converters | `delivery/http/request/foo.go` (`ToFooFilter`), `delivery/http/presenter/foo.go` (`ToFooResponse`, `ToFooListResponse`) |
| Delivery — handler | `delivery/http/handler/foo.go` |
| Wire | `wire/repository.go`, `wire/usecase.go`, `wire/handler.go` |
| Permissions | `pkg/constants/permission.go` |
| Routes | `router/public.go` \| `partner.go` \| `internal.go` |

## Key patterns

- **Repository interface** exposes `WithTx(ctx) Repository` plus CRUD: `CreateFoo`,
  `UpdateFoo`, `DeleteFoo`, `BulkCreate`, `CountFoo`, `GetFooList`, `GetFooByID`.
- **Usecase interface** exposes `Create`, `Update`, `Delete`, `BulkCreate`,
  `Count`, `GetList`, `GetByID`. The usecase depends only on the domain
  `Repository` interface.
- **GORM model** (`infrastructure/model/foo.go`) is separate from the domain
  entity — the model carries GORM tags, audit columns (`CreatedBy`/`UpdatedBy`/
  `DeletedBy`/`CreatedAt`/`UpdatedAt`/`DeletedAt`/`IsActive`); `id` has no DB
  default — the repository sets it with `utils.GenerateUUID()` (UUIDv7) on create and
  bulk create. Domain entities have no GORM tags.
- **DTOs**: never expose the GORM model directly — map model → domain entity →
  response DTO via the presenter. Request DTOs use `validate` struct tags.
- **Handler** struct is built with `NewFoo(validator, usecase)`; methods are
  `Create`, `Update`, `Delete`, `List`, `Get`.
- **Permissions**: add `PermissionFooCreate/Read/Update/Delete` constants in
  `pkg/constants/permission.go`.
- **Routes**: register in the router file matching the scope (public / partner /
  internal). Apply idempotency middleware (`RequireIdempotencyKey()` +
  `Middleware.Idempotency`) to sensitive `POST` endpoints only — `PUT`/`DELETE`
  are already idempotent.
- **Uniqueness**: enforce it with a database constraint and map the violation to 409 — a
  check-then-insert in the use case alone is racy, and two concurrent requests can both pass it.
  A business key such as `code` is unique across every row, deleted ones included, and
  immutable: the update use case loads the record and refuses a different code, and the
  repository never writes the column on update (see `bar`). The `db-migrations` skill covers
  when a key may instead be unique among live rows only.
- **Multi-domain**: if an endpoint needs data from more than one domain, add an
  application-layer service in `internal/application/` to orchestrate them, and
  have the handler call that service instead of a single usecase.

Conventions for every layer are in `.claude/rules/`; where `bar` or `foo` disagree with them,
the rules win (known gaps: phase R of `docs/roadmap/improvement-tasks.md`).

After changes, run `go build ./...`, `make lint` and `make test`, and add or update tests for the
behaviour you changed.
