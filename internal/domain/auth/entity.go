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
	PasswordHash        string
	IsActive            bool
	EmailVerified       bool
	EmailVerifiedAt     *time.Time
	PasswordChangedAt   time.Time
	LastLoginAt         *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
}

// UserSession represents one login (per device/browser). Every token issued for the login
// carries the session ID, so revoking the session revokes those tokens.
type UserSession struct {
	ID                 string
	UserID             string
	RefreshJTI         string
	PreviousRefreshJTI string
	RotatedAt          *time.Time
	DeviceName         string
	DeviceType         string
	DeviceID           string
	IPAddress          string
	UserAgent          string
	IsActive           bool
	ExpiresAt          time.Time
	LastUsedAt         time.Time
	RevokedAt          *time.Time
	RevokedReason      string
	CreatedAt          time.Time
}

// Session revocation reasons, stored in user_sessions.revoked_reason
const (
	RevokedReasonLogout         = "logout"
	RevokedReasonLogoutAll      = "logout_all"
	RevokedReasonPasswordChange = "password_change"
	RevokedReasonReuseDetected  = "reuse_detected"
	RevokedReasonAdmin          = "admin"
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

// IsExpired checks if the session has expired
func (s *UserSession) IsExpired() bool {
	return utils.Now().After(s.ExpiresAt)
}

// IsValidSession checks if the session is valid and active
func (s *UserSession) IsValidSession() bool {
	return s.IsActive && !s.IsExpired()
}
