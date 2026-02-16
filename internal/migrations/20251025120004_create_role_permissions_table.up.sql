-- Migration: create_role_permissions_table
-- Created at: 2025-10-25T12:00:04+07:00

CREATE TABLE role_permissions (
    role_id CHAR(36) NOT NULL COMMENT 'Reference to role',
    permission_id CHAR(36) NOT NULL COMMENT 'Reference to permission',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when relationship was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this relationship',
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Many-to-many relationship between roles and permissions';

-- Indexes for role_permissions
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
