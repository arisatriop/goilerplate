-- Migration: create_permissions_table
-- Created at: 2025-10-25T12:00:02+07:00

CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(150) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by VARCHAR(255)
);

-- Indexes for permissions
CREATE INDEX idx_permissions_slug ON permissions(slug);
CREATE INDEX idx_permissions_deleted_at ON permissions(deleted_at);

-- Comments for permissions
COMMENT ON TABLE permissions IS 'Permissions table for RBAC system';
COMMENT ON COLUMN permissions.id IS 'Unique identifier for the permission';
COMMENT ON COLUMN permissions.name IS 'Display name of the permission';
COMMENT ON COLUMN permissions.slug IS 'Unique slug identifier for the permission (e.g., create:example, update:example)';
COMMENT ON COLUMN permissions.description IS 'Description of what this permission grants';
