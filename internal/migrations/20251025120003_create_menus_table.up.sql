-- Migration: create_menus_table
-- Created at: 2025-10-25T12:00:03+07:00

CREATE TABLE menus (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES menus(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    icon VARCHAR(100),
    route VARCHAR(255),
    display_order DECIMAL(10,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by VARCHAR(255)
);

-- Indexes for menus
CREATE INDEX idx_menus_parent_id ON menus(parent_id);
CREATE INDEX idx_menus_slug ON menus(slug);
CREATE INDEX idx_menus_is_active ON menus(is_active);
CREATE INDEX idx_menus_display_order ON menus(display_order);
CREATE INDEX idx_menus_deleted_at ON menus(deleted_at);

-- Comments for menus
COMMENT ON TABLE menus IS 'Menus table for navigation structure';
COMMENT ON COLUMN menus.id IS 'Unique identifier for the menu';
COMMENT ON COLUMN menus.parent_id IS 'Reference to parent menu for hierarchical structure';
COMMENT ON COLUMN menus.name IS 'Display name of the menu item';
COMMENT ON COLUMN menus.slug IS 'Unique slug identifier for the menu';
COMMENT ON COLUMN menus.icon IS 'Icon identifier for the menu item';
COMMENT ON COLUMN menus.route IS 'Route/path for the menu item';
COMMENT ON COLUMN menus.display_order IS 'Order in which menu items should be displayed (supports decimals like 1.5 for easy insertion between items)';
COMMENT ON COLUMN menus.is_active IS 'Whether the menu item is active/visible';
