package auth

import (
	"time"

	"goilerplate/pkg/utils"
)

// One-time token types, matching the CHECK constraint on one_time_tokens.token_type.
// #nosec G101 -- token *type* names stored in a column, not credentials.
const (
	OneTimeTokenEmailVerification = "email_verification"
	OneTimeTokenPasswordReset     = "password_reset"
	OneTimeTokenEmailChange       = "email_change"
)

// OneTimeToken is a single-use token for email verification, password reset, or email change.
// Session tokens are not stored here: they live in user_sessions.
type OneTimeToken struct {
	ID        string
	UserID    string
	TokenType string
	TokenHash string // HMAC-SHA256 of a short OTP, or SHA-256 of a high-entropy token
	Attempts  int
	ExpiresAt time.Time
	UsedAt    *time.Time
	IPAddress string
	UserAgent string
	CreatedAt time.Time
}

// IsExpired reports whether the token can no longer be used.
func (t *OneTimeToken) IsExpired() bool {
	return !t.ExpiresAt.After(utils.Now())
}

// IsUsed reports whether the token was already consumed.
func (t *OneTimeToken) IsUsed() bool {
	return t.UsedAt != nil
}
