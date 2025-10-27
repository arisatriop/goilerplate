package auth

import (
	"context"
	"time"
)

// Repository defines the authentication repository interface
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
	LockUser(ctx context.Context, userID string, lockedUntil *time.Time) error
	UpdateUserLoginInfo(ctx context.Context, userID string, resetFailedAttempts bool) error
	IncrementFailedLoginAttempts(ctx context.Context, userID string) error
	ResetExpiredLock(ctx context.Context, userID string) error

	// Session operations
	CreateSession(ctx context.Context, session *UserSession) (*UserSession, error)
	GetSessionByID(ctx context.Context, sessionID string) (*UserSession, error)
	DeleteUserSessions(ctx context.Context, userID string) error
	DeactivateUserSessions(ctx context.Context, userID string) error

	// Token operations
	CreateToken(ctx context.Context, token *UserToken) (*UserToken, error)
	GetTokenByHash(ctx context.Context, tokenHash string) (*UserToken, error)
	GetUserTokens(ctx context.Context, userID string) ([]UserToken, error)
	DeleteTokenByHash(ctx context.Context, tokenHash string) error
	DeleteUserTokens(ctx context.Context, userID string) error
	DeleteTokensBySession(ctx context.Context, userID, sessionID string) error
	MarkTokenAsUsed(ctx context.Context, token string) error

	// Menu operations
	GetRoleMenus(ctx context.Context, userID string) ([]Menu, error)
	GetUserMenus(ctx context.Context, userID string) ([]Menu, error)
	GetMenusByParentIDs(ctx context.Context, parentIDs []string) ([]Menu, error)

	// Permission operations
	GetUserPermissionOverride(ctx context.Context, userID string, permissionSlug string) (*bool, error)
	HasRolePermission(ctx context.Context, userID string, permissionSlug string) (bool, error)
	HasMenuPermissionByMenuIDs(ctx context.Context, menuIDs []string, permissionSlug string) (bool, error)

	// Get all permission slugs user has
	GetRolePermissionSlugs(ctx context.Context, userID string) ([]string, error)
	GetMenuPermissionSlugs(ctx context.Context, menuIDs []string) ([]string, error)
	GetAllUserPermissionOverrides(ctx context.Context, userID string) (map[string]bool, error)
}
