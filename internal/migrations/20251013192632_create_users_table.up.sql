-- Migration: create_users_table
-- Created at: 2025-10-13T19:26:32+07:00

-- Enable uuid-ossp extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Unique identifier for the user
    name VARCHAR(255) NOT NULL, -- Full name of the user
    phone VARCHAR(20), -- Phone number of the user
    email VARCHAR(255) NOT NULL UNIQUE, -- Email address of the user (unique)
    password_hash VARCHAR(255) NOT NULL, -- Bcrypt hashed password
    email_verified BOOLEAN NOT NULL DEFAULT FALSE, -- Whether the user email has been verified
    email_verified_at TIMESTAMP WITH TIME ZONE DEFAULT NULL, -- When the email was verified
    password_changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- When the password was last changed
    last_login_at TIMESTAMP WITH TIME ZONE DEFAULT NULL, -- Last successful login timestamp
    failed_login_attempts INTEGER NOT NULL DEFAULT 0, -- Number of consecutive failed login attempts
    locked_until TIMESTAMP WITH TIME ZONE DEFAULT NULL, -- Account locked until this timestamp (NULL if not locked)
    avatar TEXT, -- URL or path to user avatar image
    is_active BOOLEAN NOT NULL DEFAULT TRUE, -- Whether the user account is active
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp when the user was created
    created_by VARCHAR(255) NOT NULL, -- User ID who created this record
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp when the user was last updated
    updated_by VARCHAR(255) NOT NULL, -- User ID who last updated this record
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL, -- Timestamp when the user was soft deleted (NULL if not deleted)
    deleted_by VARCHAR(255) DEFAULT NULL -- User ID who deleted this record
);

-- Add table comment
COMMENT ON TABLE users IS 'Users table for storing user account information';

-- Add column comments
COMMENT ON COLUMN users.id IS 'Unique identifier for the user';
COMMENT ON COLUMN users.name IS 'Full name of the user';
COMMENT ON COLUMN users.phone IS 'Phone number of the user';
COMMENT ON COLUMN users.email IS 'Email address of the user (unique)';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password';
COMMENT ON COLUMN users.email_verified IS 'Whether the user email has been verified';
COMMENT ON COLUMN users.email_verified_at IS 'When the email was verified';
COMMENT ON COLUMN users.password_changed_at IS 'When the password was last changed';
COMMENT ON COLUMN users.last_login_at IS 'Last successful login timestamp';
COMMENT ON COLUMN users.failed_login_attempts IS 'Number of consecutive failed login attempts';
COMMENT ON COLUMN users.locked_until IS 'Account locked until this timestamp (NULL if not locked)';
COMMENT ON COLUMN users.avatar IS 'URL or path to user avatar image';
COMMENT ON COLUMN users.is_active IS 'Whether the user account is active';
COMMENT ON COLUMN users.created_at IS 'Timestamp when the user was created';
COMMENT ON COLUMN users.created_by IS 'User ID who created this record';
COMMENT ON COLUMN users.updated_at IS 'Timestamp when the user was last updated';
COMMENT ON COLUMN users.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN users.deleted_at IS 'Timestamp when the user was soft deleted (NULL if not deleted)';
COMMENT ON COLUMN users.deleted_by IS 'User ID who deleted this record';

-- Create indexes for better query performance
CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
-- Indexes for authentication queries
CREATE INDEX idx_users_email_verified ON users(email_verified);
CREATE INDEX idx_users_password_changed_at ON users(password_changed_at);
CREATE INDEX idx_users_last_login_at ON users(last_login_at);
CREATE INDEX idx_users_failed_attempts ON users(failed_login_attempts);
CREATE INDEX idx_users_locked_until ON users(locked_until);