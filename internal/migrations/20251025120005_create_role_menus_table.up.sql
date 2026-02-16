-- Migration: create_role_menus_table
-- Created at: 2025-10-25T12:00:05+07:00

CREATE TABLE role_menus (
    role_id CHAR(36) NOT NULL COMMENT 'Reference to role',
    menu_id CHAR(36) NOT NULL COMMENT 'Reference to menu',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when relationship was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this relationship',
    PRIMARY KEY (role_id, menu_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Many-to-many relationship between roles and menus';

-- Indexes for role_menus
CREATE INDEX idx_role_menus_role_id ON role_menus(role_id);
CREATE INDEX idx_role_menus_menu_id ON role_menus(menu_id);
