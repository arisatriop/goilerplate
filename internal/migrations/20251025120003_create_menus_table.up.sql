-- Migration: create_menus_table
-- Created at: 2025-10-25T12:00:03+07:00

CREATE TABLE menus (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Unique identifier for the menu
    parent_id UUID DEFAULT NULL, -- Reference to parent menu for hierarchical structure
    name VARCHAR(100) NOT NULL, -- Display name of the menu item
    slug VARCHAR(100) UNIQUE NOT NULL, -- Unique slug identifier for the menu
    icon VARCHAR(100) DEFAULT NULL, -- Icon identifier for the menu item
    route VARCHAR(255) DEFAULT NULL, -- Route/path for the menu item
    display_order DECIMAL(10,2) DEFAULT 0, -- Order in which menu items should be displayed (supports decimals like 1.5 for easy insertion between items)
    is_active BOOLEAN DEFAULT TRUE, -- Whether the menu item is active/visible
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp when menu was created
    created_by VARCHAR(255) NOT NULL, -- User who created this menu
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp when menu was last updated
    updated_by VARCHAR(255) NOT NULL, -- User who last updated this menu
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL, -- Timestamp when menu was soft deleted
    deleted_by VARCHAR(255) DEFAULT NULL, -- User who deleted this menu
    FOREIGN KEY (parent_id) REFERENCES menus(id) ON DELETE SET NULL
);

-- Add table and column comments
COMMENT ON TABLE menus IS 'Menus table for navigation structure';
COMMENT ON COLUMN menus.id IS 'Unique identifier for the menu';
COMMENT ON COLUMN menus.parent_id IS 'Reference to parent menu for hierarchical structure';
COMMENT ON COLUMN menus.name IS 'Display name of the menu item';
COMMENT ON COLUMN menus.slug IS 'Unique slug identifier for the menu';
COMMENT ON COLUMN menus.icon IS 'Icon identifier for the menu item';
COMMENT ON COLUMN menus.route IS 'Route/path for the menu item';
COMMENT ON COLUMN menus.display_order IS 'Order in which menu items should be displayed (supports decimals like 1.5 for easy insertion between items)';
COMMENT ON COLUMN menus.is_active IS 'Whether the menu item is active/visible';
COMMENT ON COLUMN menus.created_at IS 'Timestamp when menu was created';
COMMENT ON COLUMN menus.created_by IS 'User who created this menu';
COMMENT ON COLUMN menus.updated_at IS 'Timestamp when menu was last updated';
COMMENT ON COLUMN menus.updated_by IS 'User who last updated this menu';
COMMENT ON COLUMN menus.deleted_at IS 'Timestamp when menu was soft deleted';
COMMENT ON COLUMN menus.deleted_by IS 'User who deleted this menu';

-- Indexes for menus
CREATE INDEX idx_menus_parent_id ON menus(parent_id);
CREATE INDEX idx_menus_slug ON menus(slug);
CREATE INDEX idx_menus_is_active ON menus(is_active);
CREATE INDEX idx_menus_display_order ON menus(display_order);
CREATE INDEX idx_menus_deleted_at ON menus(deleted_at);
