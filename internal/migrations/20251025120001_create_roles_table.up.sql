-- Migration: create_roles_table
-- Created at: 2025-10-25T12:00:01+07:00

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by VARCHAR(255)
);

-- Indexes for roles
CREATE INDEX idx_roles_slug ON roles(slug);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);

-- Comments for roles
COMMENT ON TABLE roles IS 'Roles table for RBAC system';
COMMENT ON COLUMN roles.id IS 'Unique identifier for the role';
COMMENT ON COLUMN roles.name IS 'Display name of the role';
COMMENT ON COLUMN roles.slug IS 'Unique slug identifier for the role';
COMMENT ON COLUMN roles.description IS 'Description of the role';
