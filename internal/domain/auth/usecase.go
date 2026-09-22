// Package auth owns authentication: who a caller is, which login they are using, and what that
// login may do. The use case is split by flow across four files, all on the same authUseCase
// type, because one 588-line file carrying every flow was the file most likely to be copied as
// the model for a new domain.
//
//	usecase.go              the type, the interface, and construction (this file)
//	usecase_signin.go       sign-in and sign-out
//	usecase_token.go        the refresh-token lifecycle
//	usecase_credentials.go  registration, password change, deactivation
//
// The collaborators each flow leans on live in their own files already: session_service.go,
// permission_service.go, menu_service.go, device_service.go, user_validator.go.
package auth

import (
	"context"
	"goilerplate/internal/domain/transaction"
	"goilerplate/pkg/jwt"
	"goilerplate/pkg/password"
	"time"
)

// SessionExpiry holds the absolute session lifetimes, from auth.session_expiry and
// auth.remember_me_expiry. A session never extends past it, however often it is refreshed.
type SessionExpiry struct {
	Default    time.Duration
	RememberMe time.Duration
}

// For reports the lifetime a session gets, honouring the remember-me choice.
func (e SessionExpiry) For(rememberMe bool) time.Duration {
	if rememberMe {
		return e.RememberMe
	}
	return e.Default
}

type authUseCase struct {
	authRepo          Repository
	jwtService        *jwt.JWTService
	userValidator     *UserValidator
	menuService       *MenuService
	sessionService    *SessionService
	permissionService *PermissionService
	sessionExpiry     SessionExpiry
	refreshReuseGrace time.Duration
	txManager         transaction.Transaction
	passwordPolicy    password.Policy
}

// Usecase defines the authentication use case interface
type Usecase interface {
	Register(ctx context.Context, entity *User, plaintextPassword string) error
	Login(ctx context.Context, credentials *LoginCredentials, deviceInfo *DeviceInfo) (*LoginResult, error)
	Logout(ctx context.Context, userID string, sessionID string) error
	LogoutAll(ctx context.Context, userID string) error
	ChangePassword(ctx context.Context, userID, sessionID, currentPassword, newPassword string) error
	DeactivateUser(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, userID, sessionID, refreshJTI string, deviceInfo *DeviceInfo) (*LoginResult, error)
}

func NewUseCase(
	authRepo Repository,
	jwtService *jwt.JWTService,
	sessionService *SessionService,
	permissionService *PermissionService,
	sessionExpiry SessionExpiry,
	refreshReuseGrace time.Duration,
	txManager transaction.Transaction,
	lockout Lockout,
	passwordPolicy password.Policy,
) Usecase {
	userValidator := NewUserValidator(authRepo, lockout)
	menuService := NewMenuService(authRepo)

	return &authUseCase{
		authRepo:          authRepo,
		jwtService:        jwtService,
		userValidator:     userValidator,
		menuService:       menuService,
		sessionService:    sessionService,
		permissionService: permissionService,
		sessionExpiry:     sessionExpiry,
		refreshReuseGrace: refreshReuseGrace,
		txManager:         txManager,
		passwordPolicy:    passwordPolicy,
	}
}
