-- Migration: create_bars_table
-- Created at: 2025-10-11T15:30:00Z

CREATE TABLE bars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(255) NOT NULL UNIQUE,
    bar TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    deleted_by VARCHAR(255) DEFAULT NULL
);

-- Comments
COMMENT ON TABLE bars IS 'Bars table for storing bar records';
COMMENT ON COLUMN bars.code IS 'Unique code identifier';
COMMENT ON COLUMN bars.bar IS 'Bar text content';
COMMENT ON COLUMN bars.is_active IS 'Whether the record is active';
COMMENT ON COLUMN bars.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN bars.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN bars.deleted_at IS 'Timestamp when record was soft deleted';
COMMENT ON COLUMN bars.created_by IS 'User who created this record';
COMMENT ON COLUMN bars.updated_by IS 'User who last updated this record';
COMMENT ON COLUMN bars.deleted_by IS 'User who deleted this record';

-- Create indexes for better performance
CREATE INDEX idx_bars_code ON bars(code);
CREATE INDEX idx_bars_is_active ON bars(is_active);
CREATE INDEX idx_bars_deleted_at ON bars(deleted_at);

-- Composite index for active records
CREATE INDEX idx_bars_active ON bars(code, is_active, deleted_at);