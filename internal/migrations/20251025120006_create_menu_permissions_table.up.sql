-- Migration: create_menu_permissions_table
-- Created at: 2025-10-25T12:00:06+07:00

CREATE TABLE menu_permissions (
    menu_id CHAR(36) NOT NULL COMMENT 'Reference to menu',
    permission_id CHAR(36) NOT NULL COMMENT 'Reference to permission',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when relationship was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this relationship',
    PRIMARY KEY (menu_id, permission_id),
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Many-to-many relationship between menus and permissions';

-- Indexes for menu_permissions
CREATE INDEX idx_menu_permissions_menu_id ON menu_permissions(menu_id);
CREATE INDEX idx_menu_permissions_permission_id ON menu_permissions(permission_id);
