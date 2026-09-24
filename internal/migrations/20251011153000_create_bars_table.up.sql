-- Migration: create_bars_table
-- Created at: 2025-10-11T15:30:00Z

-- code is the bar's business key: the identifier people and seed data use to find the same
-- record in every environment, where the UUID differs. It is therefore unique across every row,
-- soft-deleted ones included, and never reassigned — once EXP-1 has meant one bar, it cannot
-- start meaning another, or history and cross-environment references would silently change
-- meaning. The constraint's index also serves every lookup by code.
CREATE TABLE bars (
    id UUID PRIMARY KEY,
    code TEXT NOT NULL,
    bar TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted_by TEXT NULL,

    CONSTRAINT uq_bars_code UNIQUE (code)
);

COMMENT ON TABLE bars IS 'Bars: the worked example domain';
COMMENT ON COLUMN bars.code IS 'Business key: unique across all rows including soft-deleted ones, and immutable';
COMMENT ON COLUMN bars.bar IS 'Bar text content';
COMMENT ON COLUMN bars.is_active IS 'Whether the record is active';
COMMENT ON COLUMN bars.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN bars.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN bars.deleted_at IS 'Timestamp when record was soft deleted';
COMMENT ON COLUMN bars.created_by IS 'User who created this record';
COMMENT ON COLUMN bars.updated_by IS 'User who last updated this record';
COMMENT ON COLUMN bars.deleted_by IS 'User who deleted this record';

-- No other indexes. Lookups by code use the constraint's index, and lists are ordered by the
-- primary key. is_active and deleted_at alone have too few distinct values for the planner to
-- choose an index on them; add one when a real query shows it is needed.
