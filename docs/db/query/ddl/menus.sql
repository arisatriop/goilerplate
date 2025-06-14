CREATE TABLE menus (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    path VARCHAR(100) NOT NULL,
    permission VARCHAR(100) NOT NULL,
    parent_id UUID NULL,
    icon VARCHAR(100) NULL,
    "order" INT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE;
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP NULL,
    updated_by VARCHAR(100) NULL,
    deleted_at TIMESTAMP NULL,
    deleted_by VARCHAR(100) NULL,

    -- Foreign key constraints
    CONSTRAINT fk_menus_parent_id FOREIGN KEY (parent_id) REFERENCES menus(id)
);

CREATE INDEX idx_menus_permission ON menus(permission);
CREATE INDEX idx_menus_parent_id ON menus(parent_id);