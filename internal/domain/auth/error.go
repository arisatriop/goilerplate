package auth

// Configuration constants
const (
	MaxFailedLoginAttempts  = 5
	AccountLockDuration     = 10  // minutes
	AccessTokenExpiry       = 30  // minutes
	RefreshTokenExpiry      = 168 // hours (7 days)
	VerificationTokenExpiry = 24  // hours
	ResetTokenExpiry        = 1   // hours
)
