-- Migration: add_auth_fields_to_users
-- Created at: 2025-10-13T19:38:39+07:00

-- Add authentication fields to users table
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255) NOT NULL COMMENT 'Bcrypt hashed password';
ALTER TABLE users ADD COLUMN email_verified TINYINT(1) NOT NULL DEFAULT 0 COMMENT 'Whether the user email has been verified';
ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMP NULL DEFAULT NULL COMMENT 'When the email was verified';
ALTER TABLE users ADD COLUMN password_changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'When the password was last changed';
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP NULL DEFAULT NULL COMMENT 'Last successful login timestamp';
ALTER TABLE users ADD COLUMN failed_login_attempts INT NOT NULL DEFAULT 0 COMMENT 'Number of consecutive failed login attempts';
ALTER TABLE users ADD COLUMN locked_until TIMESTAMP NULL DEFAULT NULL COMMENT 'Account locked until this timestamp (NULL if not locked)';

-- Create indexes for authentication queries
CREATE INDEX idx_users_email_verified ON users(email_verified);
CREATE INDEX idx_users_password_changed_at ON users(password_changed_at);
CREATE INDEX idx_users_last_login_at ON users(last_login_at);
CREATE INDEX idx_users_failed_attempts ON users(failed_login_attempts);
CREATE INDEX idx_users_locked_until ON users(locked_until);
