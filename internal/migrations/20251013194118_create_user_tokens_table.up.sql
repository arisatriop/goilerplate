-- Migration: create_user_tokens_table
-- Created at: 2025-10-13T19:41:18+07:00

-- Create user_tokens table for email verification and password reset tokens
CREATE TABLE user_tokens (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()) COMMENT 'Unique identifier for the token',
    user_id CHAR(36) NOT NULL COMMENT 'Reference to the user who owns this token',
    token_hash VARCHAR(255) NOT NULL UNIQUE COMMENT 'Hashed token for security',
    token_type VARCHAR(50) NOT NULL COMMENT 'Type of token: email_verification, password_reset, email_change',
    expires_at TIMESTAMP NOT NULL COMMENT 'When this token expires',
    used_at TIMESTAMP NULL DEFAULT NULL COMMENT 'When this token was used (NULL if not used yet)',
    ip_address VARCHAR(45) DEFAULT NULL COMMENT 'IP address where token was created',
    user_agent TEXT COMMENT 'User agent of the request that created the token',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Table for managing verification and reset tokens';

-- Create indexes for performance
CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX idx_user_tokens_token_hash ON user_tokens(token_hash);
CREATE INDEX idx_user_tokens_type ON user_tokens(token_type);
CREATE INDEX idx_user_tokens_expires_at ON user_tokens(expires_at);
CREATE INDEX idx_user_tokens_used_at ON user_tokens(used_at);

-- Composite indexes for common queries
CREATE INDEX idx_user_tokens_user_type ON user_tokens(user_id, token_type);
CREATE INDEX idx_user_tokens_type_expires ON user_tokens(token_type, expires_at);
