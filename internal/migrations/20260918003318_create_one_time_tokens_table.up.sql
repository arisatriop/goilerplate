-- Migration: create_one_time_tokens_table
-- Created at: 2026-09-18T00:33:18Z

-- Single-use tokens for email verification, password reset, and email change.
-- Session tokens live in user_sessions; this table never holds access or refresh tokens.
CREATE TABLE one_time_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    token_type VARCHAR(50) NOT NULL,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    attempts INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL DEFAULT NULL,
    ip_address INET NULL DEFAULT NULL,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT one_time_tokens_token_type_check
        CHECK (token_type IN ('email_verification', 'password_reset', 'email_change')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Comments
COMMENT ON TABLE one_time_tokens IS 'Single-use tokens for email verification, password reset, and email change';
COMMENT ON COLUMN one_time_tokens.id IS 'Unique identifier for the token';
COMMENT ON COLUMN one_time_tokens.user_id IS 'User the token belongs to';
COMMENT ON COLUMN one_time_tokens.token_type IS 'email_verification, password_reset, or email_change';
COMMENT ON COLUMN one_time_tokens.token_hash IS 'HMAC-SHA256 of a short OTP, or SHA-256 of a high-entropy token';
COMMENT ON COLUMN one_time_tokens.attempts IS 'Failed verification attempts; the token is invalidated past the configured limit';
COMMENT ON COLUMN one_time_tokens.expires_at IS 'When this token expires';
COMMENT ON COLUMN one_time_tokens.used_at IS 'When this token was consumed (NULL while unused)';
COMMENT ON COLUMN one_time_tokens.ip_address IS 'IP address that requested the token';
COMMENT ON COLUMN one_time_tokens.user_agent IS 'User agent that requested the token';
COMMENT ON COLUMN one_time_tokens.created_at IS 'When the token was issued';

-- Indexes: the UNIQUE constraint on token_hash already provides the lookup index
CREATE INDEX idx_one_time_tokens_user_type ON one_time_tokens(user_id, token_type);
CREATE INDEX idx_one_time_tokens_expires_at ON one_time_tokens(expires_at);
