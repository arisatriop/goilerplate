-- Migration: create_examples_table
-- Created at: 2025-10-11T15:30:00Z

CREATE TABLE examples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(255) NOT NULL UNIQUE,
    example TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    deleted_by VARCHAR(255) DEFAULT NULL
);

-- Comments
COMMENT ON TABLE examples IS 'Examples table for storing example records';
COMMENT ON COLUMN examples.code IS 'Unique code identifier';
COMMENT ON COLUMN examples.example IS 'Example text content';
COMMENT ON COLUMN examples.is_active IS 'Whether the record is active';
COMMENT ON COLUMN examples.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN examples.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN examples.deleted_at IS 'Timestamp when record was soft deleted';
COMMENT ON COLUMN examples.created_by IS 'User who created this record';
COMMENT ON COLUMN examples.updated_by IS 'User who last updated this record';
COMMENT ON COLUMN examples.deleted_by IS 'User who deleted this record';

-- Create indexes for better performance
CREATE INDEX idx_examples_code ON examples(code);
CREATE INDEX idx_examples_is_active ON examples(is_active);
CREATE INDEX idx_examples_deleted_at ON examples(deleted_at);

-- Composite index for active records
CREATE INDEX idx_examples_active ON examples(code, is_active, deleted_at);