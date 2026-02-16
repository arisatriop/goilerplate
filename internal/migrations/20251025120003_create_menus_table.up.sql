-- Migration: create_menus_table
-- Created at: 2025-10-25T12:00:03+07:00

CREATE TABLE menus (
    id CHAR(36) PRIMARY KEY DEFAULT (gen_random_uuid()) COMMENT 'Unique identifier for the menu',
    parent_id CHAR(36) DEFAULT NULL COMMENT 'Reference to parent menu for hierarchical structure',
    name VARCHAR(100) NOT NULL COMMENT 'Display name of the menu item',
    slug VARCHAR(100) UNIQUE NOT NULL COMMENT 'Unique slug identifier for the menu',
    icon VARCHAR(100) DEFAULT NULL COMMENT 'Icon identifier for the menu item',
    route VARCHAR(255) DEFAULT NULL COMMENT 'Route/path for the menu item',
    display_order DECIMAL(10,2) DEFAULT 0 COMMENT 'Order in which menu items should be displayed (supports decimals like 1.5 for easy insertion between items)',
    is_active TINYINT(1) DEFAULT 1 COMMENT 'Whether the menu item is active/visible',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when menu was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User who created this menu',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when menu was last updated',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User who last updated this menu',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when menu was soft deleted',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User who deleted this menu',
    FOREIGN KEY (parent_id) REFERENCES menus(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Menus table for navigation structure';

-- Indexes for menus
CREATE INDEX idx_menus_parent_id ON menus(parent_id);
CREATE INDEX idx_menus_slug ON menus(slug);
CREATE INDEX idx_menus_is_active ON menus(is_active);
CREATE INDEX idx_menus_display_order ON menus(display_order);
CREATE INDEX idx_menus_deleted_at ON menus(deleted_at);
