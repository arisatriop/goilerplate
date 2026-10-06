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

// EmailVerificationNotice carries a verification code to the address being verified. Like a
// reset token, the plaintext code exists only here.
type EmailVerificationNotice struct {
	Email     string
	Name      string
	Code      string
	ExpiresAt time.Time
}

// AccountExistsNotice tells the owner of an address that someone tried to register it again. It
// is what registration sends instead of answering "already registered", which would tell anyone
// which addresses have accounts.
type AccountExistsNotice struct {
	Email string
	Name  string
}

// EmailChangeNotice carries the code that confirms a change of address. It goes to the new
// address, so the change proves the user reads it.
type EmailChangeNotice struct {
	NewEmail  string
	Name      string
	Code      string
	ExpiresAt time.Time
}

// EmailChangedNotice tells the old address that the account has moved. If the user did not do
// it, this is the only warning they get that someone else is in their account.
type EmailChangedNotice struct {
	OldEmail string
	NewEmail string
	Name     string
}

// Notifier delivers the messages the auth flows send to a user. The domain decides when a
// message is due and what it must carry; how it is worded, which link it contains and which
// provider sends it are the implementation's business.
//
// Implementations should return as soon as the message is accepted for delivery. A caller
// that waits for the provider would reveal, by its latency, whether a message was sent at all.
type Notifier interface {
	SendPasswordReset(ctx context.Context, notice PasswordResetNotice) error
	SendEmailVerification(ctx context.Context, notice EmailVerificationNotice) error
	SendAccountExists(ctx context.Context, notice AccountExistsNotice) error
	SendEmailChangeCode(ctx context.Context, notice EmailChangeNotice) error
	SendEmailChanged(ctx context.Context, notice EmailChangedNotice) error
}

// EmailFlows configures the auth flows that send email. With a nil Notifier they are all
// unavailable, which is how the use case is built when auth.email.enabled is false.
type EmailFlows struct {
	Notifier Notifier
	// ResetTTL is how long a reset token stays usable.
	ResetTTL time.Duration
	// ResetResendCooldown is the minimum gap between two reset emails to one account.
	ResetResendCooldown time.Duration
	OTP                 OTPPolicy
	// RequireVerifiedEmail refuses login to an account whose address is not verified.
	RequireVerifiedEmail bool
}

// OTPPolicy governs the 6-digit codes sent by email.
type OTPPolicy struct {
	// Secret keys the HMAC the codes are stored under.
	Secret []byte
	// TTL is how long a code stays usable.
	TTL time.Duration
	// MaxAttempts is how many guesses one code tolerates. Zero means one.
	MaxAttempts int
	// ResendCooldown is the minimum gap between two codes to one account.
	ResendCooldown time.Duration
}

// RequestOrigin is where a request came from, recorded with the token it issues so an
// incident can be traced back to the request that started it.
type RequestOrigin struct {
	IPAddress string
	UserAgent string
}
