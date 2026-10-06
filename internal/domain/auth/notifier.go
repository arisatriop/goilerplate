package auth

import (
	"context"
	"time"
)

// PasswordResetNotice is what a user needs to finish a password reset. The token is the
// plaintext; it exists nowhere else, because only its hash is stored.
type PasswordResetNotice struct {
	Email     string
	Name      string
	Token     string
	ExpiresAt time.Time
}

// Notifier delivers the messages the auth flows send to a user. The domain decides when a
// message is due and what it must carry; how it is worded, which link it contains and which
// provider sends it are the implementation's business.
//
// Implementations should return as soon as the message is accepted for delivery. A caller
// that waits for the provider would reveal, by its latency, whether a message was sent at all.
type Notifier interface {
	SendPasswordReset(ctx context.Context, notice PasswordResetNotice) error
}

// Recovery configures account recovery by email. With a nil Notifier the recovery flows are
// unavailable, which is how the use case is built when auth.email.enabled is false.
type Recovery struct {
	Notifier Notifier
	// ResetTTL is how long a reset token stays usable.
	ResetTTL time.Duration
	// ResendCooldown is the minimum gap between two reset emails to one account.
	ResendCooldown time.Duration
}

// RequestOrigin is where a request came from, recorded with the token it issues so an
// incident can be traced back to the request that started it.
type RequestOrigin struct {
	IPAddress string
	UserAgent string
}
