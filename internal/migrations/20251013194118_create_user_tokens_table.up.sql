-- Migration: create_user_tokens_table
-- Created at: 2025-10-13T19:41:18+07:00

-- Create user_tokens table for email verification and password reset tokens
CREATE TABLE user_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    token_type VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP NULL DEFAULT NULL,
    ip_address VARCHAR(45) DEFAULT NULL,
    user_agent TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Comments
COMMENT ON COLUMN user_tokens.id IS 'Unique identifier for the token';
COMMENT ON COLUMN user_tokens.user_id IS 'Reference to the user who owns this token';
COMMENT ON COLUMN user_tokens.token_hash IS 'Hashed token for security';
COMMENT ON COLUMN user_tokens.token_type IS 'Type of token: email_verification, password_reset, email_change';
COMMENT ON COLUMN user_tokens.expires_at IS 'When this token expires';
COMMENT ON COLUMN user_tokens.used_at IS 'When this token was used (NULL if not used yet)';
COMMENT ON COLUMN user_tokens.ip_address IS 'IP address where token was created';
COMMENT ON COLUMN user_tokens.user_agent IS 'User agent of the request that created the token';
COMMENT ON TABLE user_tokens IS 'Table for managing verification and reset tokens';

-- Create indexes for performance
CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX idx_user_tokens_token_hash ON user_tokens(token_hash);
CREATE INDEX idx_user_tokens_type ON user_tokens(token_type);
CREATE INDEX idx_user_tokens_expires_at ON user_tokens(expires_at);
CREATE INDEX idx_user_tokens_used_at ON user_tokens(used_at);

-- Composite indexes for common queries
CREATE INDEX idx_user_tokens_user_type ON user_tokens(user_id, token_type);
CREATE INDEX idx_user_tokens_type_expires ON user_tokens(token_type, expires_at);
