---
name: db-migrations
description: Conventions for writing SQL database migrations in this project. Use when creating or editing files in internal/migrations/, running make migrate-* commands, or designing a new table schema.
---

# Database Migrations

Migrations run through the **in-house migrator** in `pkg/migration` — not golang-migrate. It takes
a PostgreSQL advisory lock and applies each file in **one transaction**, sending the whole file to
the server (so dollar-quoted function bodies and comments work). A failed file rolls back both the
schema change and its bookkeeping row: there is no dirty state to force.

Each migration is a pair in `internal/migrations/`: `<version>_<name>.up.sql` and
`<version>_<name>.down.sql`.

## Creating and running

```bash
make migrate-create name=create_products_table   # generates the pair with a timestamp version
make migrate-up        # apply pending
make migrate-down      # roll back the last one
make migrate-status    # applied / pending
```

- Always generate the pair with `make migrate-create`. The version prefix is `YYYYMMDDHHMMSS`.
  Names are snake_case: `create_<plural>_table`, `add_<column>_to_<table>`,
  `create_<a>_<b>_table` for a join table
- Don't run migrations against a shared database on the user's behalf
- **Never edit a migration that has been applied anywhere else.** Add a new one. The migrator
  records applied versions, not checksums, so an edited file silently diverges between
  environments

## Table conventions

```sql
-- Migration: create_products_table

CREATE TABLE products (
    id          UUID PRIMARY KEY,
    code        TEXT NOT NULL,
    name        TEXT NOT NULL,
    price       NUMERIC(19, 4) NOT NULL CHECK (price >= 0),
    category_id UUID NOT NULL REFERENCES categories(id),
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by  TEXT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by  TEXT NOT NULL,
    deleted_at  TIMESTAMPTZ NULL,
    deleted_by  TEXT NULL,

    -- code is the business key: stable across environments, where the UUID is not. Unique
    -- across every row, deleted ones included, so a code never comes to mean another product.
    CONSTRAINT uq_products_code UNIQUE (code)
);

COMMENT ON TABLE products IS 'Products offered for sale';
COMMENT ON COLUMN products.code IS 'Business key; unique across all rows including soft-deleted ones, and immutable';

-- PostgreSQL does not index foreign keys automatically; without this, joins and
-- ON DELETE checks on categories scan products.
CREATE INDEX idx_products_category_id ON products (category_id);
```

- **Primary key** `id UUID PRIMARY KEY` with **no database default**. The repository assigns it with
  `utils.GenerateUUID()` (UUIDv7, time-ordered for index locality). An insert that forgets the ID
  then fails loudly instead of quietly getting a random one
- **Types**: `TEXT` for strings; add a `CHECK (char_length(x) <= n)` when a limit is a real rule.
  `VARCHAR(255)` is a MySQL habit that buys nothing in PostgreSQL. `TIMESTAMPTZ` for every time
  column, never `TIMESTAMP`. `NUMERIC(p, s)` for money, never `REAL`/`DOUBLE PRECISION`
- **Constraints belong in the database**: `NOT NULL` by default, `CHECK` for invariants,
  `REFERENCES` for relations. They are the last line of defence when application code is wrong
- **Audit columns**: `created_at/by`, `updated_at/by` (`NOT NULL`) and `deleted_at/by` (nullable)
  on tables that model business entities. Log-like and join tables need not carry them
- **Soft delete and uniqueness**: decide what the key means before choosing the constraint.
  - A **business key** (`code` on master data) identifies one record in every environment and in
    history. Make it a plain `UNIQUE` across every row, deleted ones included, and never let an
    update change it. A deleted row keeps its code; bring it back by restoring the row, not by
    creating another with the same code
  - A key that may legitimately **pass to a new record** once the old one is gone (a username or
    email freed by account deletion) gets a partial unique index `WHERE deleted_at IS NULL`.
    Upserts against it must repeat the predicate: `ON CONFLICT (email) WHERE deleted_at IS NULL`
- **Comments**: `COMMENT ON TABLE`, plus `COMMENT ON COLUMN` wherever the name doesn't say
  everything (units, allowed values, why it is nullable)

## Indexes — derive them from queries
- Index every foreign key column
- Index what the queries filter and sort on, as composites in the order of the predicates, rather
  than one index per column
- Don't index a boolean or other low-cardinality column on its own; it is rarely used. If most
  queries read live rows, make the index partial (`WHERE deleted_at IS NULL`)
- Don't duplicate an index a constraint already creates: `UNIQUE` and `PRIMARY KEY` are indexes
- Name them `idx_<table>_<columns>` and `uq_<table>_<columns>`

## Changing a live table
Every file runs inside one transaction, which shapes what is safe:

- `CREATE INDEX CONCURRENTLY` and `ALTER TYPE ... ADD VALUE` (on older PostgreSQL) **cannot run in a
  transaction**, so they cannot go through this migrator. On a large table, build the index out of
  band and add a migration that only records it (`CREATE INDEX IF NOT EXISTS`)
- A plain `CREATE INDEX` or a table rewrite holds a lock that blocks writes for its whole duration.
  That is fine on a small table and an outage on a large one. Set `SET LOCAL lock_timeout = '5s'`
  at the top of the file so a blocked migration fails fast instead of queueing every query behind it
- Use **expand → migrate → contract** for anything the running version still reads. For example,
  to rename a column: add the new column, deploy code that writes both, backfill, switch reads,
  then drop the old column in a later release. A one-step rename breaks the pods still running
  the previous version during a rolling deploy
- Adding a column with a constant `DEFAULT` is cheap (no rewrite). Adding `NOT NULL` to an existing
  column scans the table; add a `CHECK (...) NOT VALID`, then `VALIDATE CONSTRAINT` separately

## down.sql
- Reverses the `up` schema change exactly: `CREATE TABLE` → `DROP TABLE IF EXISTS`,
  `ADD COLUMN` → `DROP COLUMN`
- A `down` that drops data cannot restore it. In production, prefer a new forward migration over
  running `down`. `down` is for development and for rolling back a release that has not written
  data yet
- Baseline rows the application cannot start without (for example the `owner` role) belong in a
  migration, not a seeder
