---
description: Pattern guide for creating new tables, GORM models, domain entities, and migration files
---

When the user asks to create a new table, follow these conventions derived from the existing schema in `internal/migrations/`.

**IMPORTANT:** Every new table MUST have its migration written to `internal/migrations/<timestamp>_create_<table>_table.up.sql` (and a corresponding `.down.sql`). Do this as the first step before anything else. Use `make migrate-create name=<name>` to generate the files, then fill in the SQL.

## SQL Migration Pattern

```sql
-- Migration: create_<table>_table

CREATE TABLE <table_name> (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- domain columns here
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMP    NULL     DEFAULT NULL,
    deleted_by VARCHAR(255)          DEFAULT NULL
);

COMMENT ON TABLE <table_name> IS '...';

CREATE INDEX idx_<table>_<col> ON <table_name>(<col>);
CREATE INDEX idx_<table>_deleted_at ON <table_name>(deleted_at);
```

### Rules
- Primary key: `UUID`, default `gen_random_uuid()`
- All tables include the 6 audit columns (`created_*`, `updated_*`, `deleted_*`) — soft delete via `deleted_at`
- FK constraint name: `fk_<table>_<column>`
- Index name: `idx_<table>_<column>` — add an index for every FK column and `deleted_at`
- Use `VARCHAR(255)` for most strings, `TEXT` for long/unbounded strings
- Use `UUID` for FK references (matching UUID PKs)
- Use `NUMERIC`/`DECIMAL` for financial values — never `FLOAT` or `DOUBLE PRECISION`
- Down migration must DROP TABLE

## GORM Model Pattern

```go
package model

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type <Entity> struct {
    ID        string     `gorm:"primaryKey;type:char(36);default:uuid()"`
    // domain fields with gorm tags
    CreatedAt time.Time  `gorm:"not null"`
    CreatedBy string     `gorm:"type:varchar(255);not null"`
    UpdatedAt time.Time  `gorm:"not null"`
    UpdatedBy string     `gorm:"type:varchar(255);not null"`
    DeletedAt *time.Time `gorm:"index"`
    DeletedBy *string    `gorm:"type:varchar(255)"`
}

func (<Entity>) TableName() string {
    return "<table_name>"
}

func (e *<Entity>) BeforeCreate(tx *gorm.DB) error {
    if e.ID == "" {
        e.ID = uuid.NewString()
    }
    return nil
}
```

- Place in `internal/infrastructure/model/<entity>.go`
- Nullable DB columns → pointer types (`*string`, `*time.Time`)
- Use `gorm` struct tags only — no `json` tags on models

## Domain Entity Pattern

```go
package <domain>

type <Entity> struct {
    ID     string
    // domain fields — no GORM tags, no json tags
}
```

- Place in `internal/domain/<name>/entity.go`
- Domain entities have **no GORM, no json tags** — pure Go structs
- Use `github.com/shopspring/decimal` for any financial field (never `float64`)

## Checklist when adding a new table

- [ ] Run `make migrate-create name=create_<table>_table` to generate migration files
- [ ] Write SQL in `.up.sql`; write `DROP TABLE` in `.down.sql`
- [ ] GORM model in `internal/infrastructure/model/<entity>.go`
- [ ] Domain entity in `internal/domain/<name>/entity.go`
- [ ] Repository interface in `internal/domain/<name>/repository.go`
- [ ] Repository implementation in `internal/infrastructure/repository/<name>.go`
- [ ] `deleted_at IS NULL` filter on all SELECT queries (soft delete)
- [ ] FK columns have DB indexes
