-- Migration: create_users_table
-- Created at: 2025-10-13T19:26:32+07:00

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(255) NOT NULL UNIQUE,
    avatar TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    deleted_by VARCHAR(255) DEFAULT NULL
);

-- Comments
COMMENT ON COLUMN users.id IS 'Unique identifier for the user';
COMMENT ON COLUMN users.name IS 'Full name of the user';
COMMENT ON COLUMN users.phone IS 'Phone number of the user';
COMMENT ON COLUMN users.email IS 'Email address of the user (unique)';
COMMENT ON COLUMN users.avatar IS 'URL or path to user avatar image';
COMMENT ON COLUMN users.is_active IS 'Whether the user account is active';
COMMENT ON COLUMN users.created_at IS 'Timestamp when the user was created';
COMMENT ON COLUMN users.created_by IS 'User ID who created this record';
COMMENT ON COLUMN users.updated_at IS 'Timestamp when the user was last updated';
COMMENT ON COLUMN users.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN users.deleted_at IS 'Timestamp when the user was soft deleted (NULL if not deleted)';
COMMENT ON COLUMN users.deleted_by IS 'User ID who deleted this record';
COMMENT ON TABLE users IS 'Users table for storing user account information';

-- Create indexes for better query performance
CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);