-- Migration: create_user_sessions_table
-- Created at: 2025-10-13T19:40:35+07:00

-- One row per login (per device/browser). Every access and refresh token carries this row's
-- id, so setting is_active = false revokes the whole login. The active refresh token is
-- identified by its jti, not by a stored hash: refresh tokens are JWTs and are rotated on
-- every refresh.
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    refresh_jti VARCHAR(64) NOT NULL,
    previous_refresh_jti VARCHAR(64) NULL DEFAULT NULL,
    rotated_at TIMESTAMPTZ NULL DEFAULT NULL,
    device_name VARCHAR(255) NULL DEFAULT NULL,
    device_type VARCHAR(50) NULL DEFAULT NULL,
    device_id VARCHAR(255) NULL DEFAULT NULL,
    ip_address INET NULL DEFAULT NULL,
    user_agent TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMPTZ NULL DEFAULT NULL,
    revoked_reason VARCHAR(50) NULL DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Comments
COMMENT ON TABLE user_sessions IS 'One row per login; the source of truth for token revocation';
COMMENT ON COLUMN user_sessions.id IS 'Session identifier carried by every token issued for this login';
COMMENT ON COLUMN user_sessions.user_id IS 'User the session belongs to';
COMMENT ON COLUMN user_sessions.refresh_jti IS 'jti of the refresh token currently accepted for this session';
COMMENT ON COLUMN user_sessions.previous_refresh_jti IS 'jti replaced by the last rotation; accepted within auth.refresh_reuse_grace';
COMMENT ON COLUMN user_sessions.rotated_at IS 'When refresh_jti was last rotated (NULL until the first refresh)';
COMMENT ON COLUMN user_sessions.device_name IS 'Human-readable device name (e.g. "iPhone")';
COMMENT ON COLUMN user_sessions.device_type IS 'mobile, desktop, tablet, or web';
COMMENT ON COLUMN user_sessions.device_id IS 'Device fingerprint';
COMMENT ON COLUMN user_sessions.ip_address IS 'IP address the session was created from';
COMMENT ON COLUMN user_sessions.user_agent IS 'Browser or app user agent string';
COMMENT ON COLUMN user_sessions.is_active IS 'False once the session is revoked; checked on every authenticated request';
COMMENT ON COLUMN user_sessions.expires_at IS 'Absolute session lifetime; rotation never extends it';
COMMENT ON COLUMN user_sessions.last_used_at IS 'When this session last refreshed its tokens';
COMMENT ON COLUMN user_sessions.revoked_at IS 'When the session was revoked (NULL while active)';
COMMENT ON COLUMN user_sessions.revoked_reason IS 'logout, logout_all, password_change, reuse_detected, or admin';
COMMENT ON COLUMN user_sessions.created_at IS 'When the user logged in';

-- Indexes: listing a user's sessions and the cleanup job's expiry sweep are the only scans.
-- Lookups by id are served by the primary key.
CREATE INDEX idx_user_sessions_user_active ON user_sessions(user_id, is_active);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
