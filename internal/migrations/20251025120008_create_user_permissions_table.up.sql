-- Migration: create_user_permissions_table
-- Created at: 2025-10-25T12:00:08+07:00

CREATE TABLE user_permissions (
    user_id CHAR(36) NOT NULL COMMENT 'Reference to user',
    permission_id CHAR(36) NOT NULL COMMENT 'Reference to permission',
    is_granted TINYINT(1) DEFAULT 1 COMMENT 'TRUE = explicitly granted, FALSE = explicitly revoked',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when relationship was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this relationship',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when relationship was last updated',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this relationship',
    PRIMARY KEY (user_id, permission_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User-specific permission overrides (grants or revokes)';

-- Indexes for user_permissions
CREATE INDEX idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX idx_user_permissions_permission_id ON user_permissions(permission_id);
CREATE INDEX idx_user_permissions_is_granted ON user_permissions(is_granted);
