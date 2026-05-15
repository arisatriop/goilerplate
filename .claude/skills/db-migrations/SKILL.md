---
name: db-migrations
description: Conventions for writing SQL database migrations in this project. Use when creating or editing files in internal/migrations/, running make migrate-* commands, or designing a new table schema.
---

# Database Migrations

Migrations use **golang-migrate** with plain SQL files in `internal/migrations/`.
Each migration is a pair: `<version>_<name>.up.sql` and `<version>_<name>.down.sql`.

## Creating a migration

Always generate the file pair with the Makefile target — never create the files
by hand (it produces the correct timestamp version prefix):

```bash
make migrate-create name=create_products_table
```

Version prefix is a `YYYYMMDDHHMMSS` timestamp. Name uses `snake_case`, e.g.
`create_<plural>_table`, `add_<field>_to_<table>`, `create_<a>_<b>_table` (join table).

Run / roll back / inspect:
```bash
make migrate-up       # apply all pending migrations
make migrate-down     # roll back the last migration
make migrate-status   # show applied / pending
```
Never run migrations against the database automatically — leave that to the user.

## up.sql conventions

Header comment, then `CREATE TABLE`. A standard CRUD table:

```sql
-- Migration: create_products_table
-- Created at: 2026-05-15T10:00:00+07:00

CREATE TABLE products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(255) NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by  VARCHAR(255) NOT NULL,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by  VARCHAR(255) NOT NULL,
    deleted_at  TIMESTAMP NULL DEFAULT NULL,
    deleted_by  VARCHAR(255) DEFAULT NULL
);

-- Comments
COMMENT ON TABLE products IS 'Products table';
COMMENT ON COLUMN products.code IS 'Unique code identifier';
-- ... one COMMENT ON COLUMN per column

-- Indexes
CREATE INDEX idx_products_code ON products(code);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_deleted_at ON products(deleted_at);
```

Rules:
- Primary key is always `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`.
- Every table has the full audit column set: `created_at` / `created_by` /
  `updated_at` / `updated_by` (all `NOT NULL`) and `deleted_at` / `deleted_by`
  (both nullable — soft delete).
- `is_active BOOLEAN NOT NULL DEFAULT true` when the entity has an active flag.
- Add `COMMENT ON TABLE` and a `COMMENT ON COLUMN` for every column.
- Index the business key, `deleted_at`, and `is_active` (if present). Add a
  composite index when queries filter on several columns together.
- Match column names/types to the GORM model in `internal/infrastructure/model/`.

## down.sql conventions

Header comment, then the reverse operation. For a `CREATE TABLE`:

```sql
-- Rollback: create_products_table
-- Created at: 2026-05-15T10:00:00+07:00

DROP TABLE IF EXISTS products;
```

The `down` file must fully reverse the `up` file — every migration is reversible.
For `ALTER TABLE ... ADD COLUMN`, the down is `ALTER TABLE ... DROP COLUMN`.
