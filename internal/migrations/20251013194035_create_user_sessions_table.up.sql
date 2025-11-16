-- Migration: create_user_sessions_table
-- Created at: 2025-10-13T19:40:35+07:00

-- Create user_sessions table for managing refresh tokens and multiple device logins
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Unique identifier for the session
    user_id UUID NOT NULL, -- Reference to the user who owns this session
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE, -- Hashed refresh token for security
    device_name VARCHAR(255) DEFAULT NULL, -- Human-readable device name (e.g., "John iPhone")
    device_type VARCHAR(50) DEFAULT NULL, -- Type of device: mobile, desktop, tablet, web
    device_id VARCHAR(255) DEFAULT NULL, -- Unique identifier for the device
    ip_address VARCHAR(45) DEFAULT NULL, -- IP address of the session
    user_agent TEXT, -- Browser/app user agent string
    location VARCHAR(255) DEFAULT NULL, -- Approximate location based on IP
    is_active BOOLEAN NOT NULL DEFAULT TRUE, -- Whether this session is currently active
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL, -- When this refresh token expires
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- When this session was last used to refresh tokens
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Add table and column comments
COMMENT ON TABLE user_sessions IS 'Table for managing user refresh tokens and device sessions';
COMMENT ON COLUMN user_sessions.id IS 'Unique identifier for the session';
COMMENT ON COLUMN user_sessions.user_id IS 'Reference to the user who owns this session';
COMMENT ON COLUMN user_sessions.refresh_token_hash IS 'Hashed refresh token for security';
COMMENT ON COLUMN user_sessions.device_name IS 'Human-readable device name (e.g., "John iPhone")';
COMMENT ON COLUMN user_sessions.device_type IS 'Type of device: mobile, desktop, tablet, web';
COMMENT ON COLUMN user_sessions.device_id IS 'Unique identifier for the device';
COMMENT ON COLUMN user_sessions.ip_address IS 'IP address of the session';
COMMENT ON COLUMN user_sessions.user_agent IS 'Browser/app user agent string';
COMMENT ON COLUMN user_sessions.location IS 'Approximate location based on IP';
COMMENT ON COLUMN user_sessions.is_active IS 'Whether this session is currently active';
COMMENT ON COLUMN user_sessions.expires_at IS 'When this refresh token expires';
COMMENT ON COLUMN user_sessions.last_used_at IS 'When this session was last used to refresh tokens';

-- Create indexes for performance
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_refresh_token ON user_sessions(refresh_token_hash);
CREATE INDEX idx_user_sessions_device_id ON user_sessions(device_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_is_active ON user_sessions(is_active);
CREATE INDEX idx_user_sessions_last_used ON user_sessions(last_used_at);
CREATE INDEX idx_user_sessions_ip_address ON user_sessions(ip_address);

-- Composite indexes for common queries
CREATE INDEX idx_user_sessions_user_active ON user_sessions(user_id, is_active);
CREATE INDEX idx_user_sessions_user_device ON user_sessions(user_id, device_id);
