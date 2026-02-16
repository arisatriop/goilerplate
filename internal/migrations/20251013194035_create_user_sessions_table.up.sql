-- Migration: create_user_sessions_table
-- Created at: 2025-10-13T19:40:35+07:00

-- Create user_sessions table for managing refresh tokens and multiple device logins
CREATE TABLE user_sessions (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()) COMMENT 'Unique identifier for the session',
    user_id CHAR(36) NOT NULL COMMENT 'Reference to the user who owns this session',
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE COMMENT 'Hashed refresh token for security',
    device_name VARCHAR(255) DEFAULT NULL COMMENT 'Human-readable device name (e.g., "John iPhone")',
    device_type VARCHAR(50) DEFAULT NULL COMMENT 'Type of device: mobile, desktop, tablet, web',
    device_id VARCHAR(255) DEFAULT NULL COMMENT 'Unique identifier for the device',
    ip_address VARCHAR(45) DEFAULT NULL COMMENT 'IP address of the session',
    user_agent TEXT COMMENT 'Browser/app user agent string',
    location VARCHAR(255) DEFAULT NULL COMMENT 'Approximate location based on IP',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'Whether this session is currently active',
    expires_at TIMESTAMP NOT NULL COMMENT 'When this refresh token expires',
    last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'When this session was last used to refresh tokens',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Table for managing user refresh tokens and device sessions';

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
