-- Migration: create_examples_table
-- Created at: 2025-10-11T15:30:00Z

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE examples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(255) NOT NULL UNIQUE,
    example TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    deleted_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better performance
CREATE INDEX idx_examples_code ON examples(code);
CREATE INDEX idx_examples_is_active ON examples(is_active);
CREATE INDEX idx_examples_created_by ON examples(created_by);
CREATE INDEX idx_examples_deleted_at ON examples(deleted_at);

-- Create partial index for active records only
CREATE INDEX idx_examples_active ON examples(code, is_active) WHERE deleted_at IS NULL;