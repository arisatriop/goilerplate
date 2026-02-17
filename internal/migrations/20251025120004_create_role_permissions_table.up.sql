-- Migration: create_role_permissions_table
-- Created at: 2025-10-25T12:00:04+07:00

CREATE TABLE role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- Comments
COMMENT ON COLUMN role_permissions.role_id IS 'Reference to role';
COMMENT ON COLUMN role_permissions.permission_id IS 'Reference to permission';
COMMENT ON COLUMN role_permissions.created_at IS 'Timestamp when relationship was created';
COMMENT ON COLUMN role_permissions.created_by IS 'User who created this relationship';
COMMENT ON TABLE role_permissions IS 'Many-to-many relationship between roles and permissions';

-- Indexes for role_permissions
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
