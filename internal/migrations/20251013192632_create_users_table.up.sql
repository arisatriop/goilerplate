-- Migration: create_users_table
-- Created at: 2025-10-13T19:26:32+07:00

CREATE TABLE users (
    id CHAR(36) PRIMARY KEY DEFAULT (gen_random_uuid()) COMMENT 'Unique identifier for the user',
    name VARCHAR(255) NOT NULL COMMENT 'Full name of the user',
    phone VARCHAR(20) COMMENT 'Phone number of the user',
    email VARCHAR(255) NOT NULL UNIQUE COMMENT 'Email address of the user (unique)',
    avatar TEXT COMMENT 'URL or path to user avatar image',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether the user account is active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when the user was created',
    created_by VARCHAR(255) NOT NULL COMMENT 'User ID who created this record',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Timestamp when the user was last updated',
    updated_by VARCHAR(255) NOT NULL COMMENT 'User ID who last updated this record',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Timestamp when the user was soft deleted (NULL if not deleted)',
    deleted_by VARCHAR(255) DEFAULT NULL COMMENT 'User ID who deleted this record'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Users table for storing user account information';

-- Create indexes for better query performance
CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);