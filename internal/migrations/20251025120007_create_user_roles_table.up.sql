-- Migration: create_user_roles_table
-- Created at: 2025-10-25T12:00:07+07:00

CREATE TABLE user_roles (
    user_id CHAR(36) NOT NULL COMMENT 'Reference to user',
    role_id CHAR(36) NOT NULL COMMENT 'Reference to role',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when relationship was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this relationship',
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Many-to-many relationship between users and roles';

-- Indexes for user_roles
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
