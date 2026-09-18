package auth

import (
	"context"
	"time"
)

// Repository defines the authentication repository interface
type Repository interface {

	// Session operations
	CreateSession(ctx context.Context, session *UserSession) (*UserSession, error)
	GetSessionByID(ctx context.Context, sessionID string) (*UserSession, error)
	DeleteUserSessions(ctx context.Context, userID string) error
	DeactivateUserSessions(ctx context.Context, userID, reason string) error
	RevokeSession(ctx context.Context, userID, sessionID, reason string) error

	// User operations
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
	LockUser(ctx context.Context, userID string, lockedUntil *time.Time) error
	UpdateUserLoginInfo(ctx context.Context, userID string, resetFailedAttempts bool) error
	IncrementFailedLoginAttempts(ctx context.Context, userID string) error
	ResetExpiredLock(ctx context.Context, userID string) error

	// One-time token operations (email verification, password reset, email change)
	CreateOneTimeToken(ctx context.Context, token *OneTimeToken) error
	GetLatestActiveOneTimeToken(ctx context.Context, userID, tokenType string) (*OneTimeToken, error)
	ConsumeOneTimeToken(ctx context.Context, tokenHash, tokenType string) error
	IncrementOneTimeTokenAttempts(ctx context.Context, tokenID string) (int, error)

	// Menu Operations
	GetParentMenus(ctx context.Context) ([]Menu, error)
	GetMenusByParentIDs(ctx context.Context, parentIDs []string) ([]Menu, error)
	GetMenuPermissionsByMenuID(ctx context.Context, menuID string) ([]string, error)

	// Role and Permission operations
	GetUserRolesByUserID(ctx context.Context, userID string) ([]string, error)
	GetRolePermissionsByRoleIDs(ctx context.Context, roleIDs []string) ([]string, error)
	GetUserPermissionOverrides(ctx context.Context, userID string) (map[string]bool, error)
}
