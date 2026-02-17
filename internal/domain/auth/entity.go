package auth

import (
	"goilerplate/pkg/jwt"
	"goilerplate/pkg/utils"
	"time"
)

// LoginCredentials represents login input data
type LoginCredentials struct {
	Email      string
	Password   string
	RememberMe bool
}

// LoginResult represents the successful login result
type LoginResult struct {
	User       *User
	Menu       []Menu
	Permission []string
	Tokens     *jwt.TokenPair
	Session    *UserSession
}

// User represents the user entity for authentication
type User struct {
	ID                  string
	Name                string
	Email               string
	Avatar              string
	Password            string
	PasswordHash        string
	IsActive            bool
	EmailVerified       bool
	EmailVerifiedAt     *time.Time
	PasswordChangedAt   time.Time
	LastLoginAt         *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
	RememberMe          bool
}

// UserSession represents a user session/device
type UserSession struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	DeviceName       string
	DeviceType       string
	DeviceID         string
	IPAddress        string
	UserAgent        string
	Location         string
	IsActive         bool
	ExpiresAt        time.Time
	LastUsedAt       time.Time
}

// TokenPair represents token pair in domain layer
// type TokenPair struct {
// 	AccessToken           string
// 	AccessTokenType       string
// 	AccessTokenExpiresIn  int64
// 	AccessTokenExpiresAt  time.Time
// 	RefreshToken          string
// 	RefreshExpiresAt      time.Time
// 	RefreshTokenExpiresIn int64
// 	RefreshTokenExpiresAt time.Time
// }

// UserToken represents verification and reset tokens
type UserToken struct {
	ID        string
	UserID    string
	TokenHash string
	TokenType string
	ExpiresAt time.Time
	UsedAt    *time.Time
	IsRevoked bool
	IPAddress string
	UserAgent string
}

type Login struct {
	User    *User
	Token   *jwt.TokenPair
	Session *UserSession
}

// Token types
const (
	TokenTypeEmailVerification = "email_verification"
	TokenTypePasswordReset     = "password_reset"
	TokenTypeEmailChange       = "email_change"
	TokenTypeRefresh           = jwt.RefreshToken
	TokenTypeAccess            = jwt.AccessToken
)

// Device types
const (
	DeviceTypeMobile  = "mobile"
	DeviceTypeDesktop = "desktop"
	DeviceTypeTablet  = "tablet"
	DeviceTypeWeb     = "web"
)

// IsLocked checks if the user account is locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return utils.Now().Before(*u.LockedUntil)
}

// HasExpiredLock checks if the user has a lock that has expired
func (u *User) HasExpiredLock() bool {
	if u.LockedUntil == nil {
		return false
	}
	return utils.Now().After(*u.LockedUntil) || utils.Now().Equal(*u.LockedUntil)
}

// ShouldLockAccount checks if account should be locked based on failed attempts
func (u *User) ShouldLockAccount(maxAttempts int) bool {
	return u.FailedLoginAttempts >= maxAttempts
}

// IsAdmin checks if user has admin privileges
func (u *User) IsAdmin() bool {
	// You can implement your own role-based access control here
	// For now, we'll use a simple email-based check or add a role field
	return false // TODO: implement proper role system
}

// IsExpired checks if the session has expired
func (s *UserSession) IsExpired() bool {
	return utils.Now().After(s.ExpiresAt)
}

// IsValidSession checks if the session is valid and active
func (s *UserSession) IsValidSession() bool {
	return s.IsActive && !s.IsExpired()
}

// IsExpired checks if the token has expired
func (t *UserToken) IsExpired() bool {
	return utils.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used
func (t *UserToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid checks if the token is valid (not expired and not used)
func (t *UserToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// Login represents login request in domain layer
// type Login struct {
// 	Email      string
// 	Password   string
// 	RememberMe bool
// 	DeviceName string
// 	DeviceType string
// 	DeviceID   string
// 	UserAgent  string
// 	IPAddress  string
// }

// // ChangePassword represents change password request in domain layer
// type ChangePassword struct {
// 	CurrentPassword string
// 	NewPassword     string
// 	ConfirmPassword string
// }

// // ForgotPassword represents forgot password request in domain layer
// type ForgotPassword struct {
// 	Email     string
// 	UserAgent string
// 	IPAddress string
// }

// // ResetPassword represents reset password request in domain layer
// type ResetPassword struct {
// 	Token           string
// 	NewPassword     string
// 	ConfirmPassword string
// 	UserAgent       string
// 	IPAddress       string
// }

// // RefreshToken represents refresh token request in domain layer
// type RefreshToken struct {
// 	RefreshToken string
// 	DeviceID     string
// 	UserAgent    string
// 	IPAddress    string
// }

// // AuthResponse represents authentication response in domain layer
// type AuthResponse struct {
// 	User    *User
// 	Tokens  *TokenPair
// 	Session *UserSession
// }
