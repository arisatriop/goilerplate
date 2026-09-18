package auth

import (
	"context"
	"errors"
	"fmt"
	"goilerplate/internal/domain/transaction"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/jwt"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
	"net/http"
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
}

// Usecase defines the authentication use case interface
type Usecase interface {
	Register(ctx context.Context, entity *User) error
	Login(ctx context.Context, credentials *LoginCredentials, deviceInfo *DeviceInfo) (*LoginResult, error)
	Logout(ctx context.Context, userID string, sessionID string) error
	LogoutAll(ctx context.Context, userID string) error
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
	}
}

// Register creates a new user account
func (uc *authUseCase) Register(ctx context.Context, entity *User) error {
	existingUser, err := uc.authRepo.GetUserByEmail(ctx, entity.Email)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}
	if existingUser != nil {
		return utils.ClientErr(http.StatusBadRequest, "User is already registered")
	}

	hashedPassword, err := utils.HashPassword(entity.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	entity.PasswordHash = hashedPassword
	entity.IsActive = true

	_, err = uc.authRepo.CreateUser(ctx, entity)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Login authenticates a user and creates a session
func (uc *authUseCase) Login(ctx context.Context, credentials *LoginCredentials, deviceInfo *DeviceInfo) (*LoginResult, error) {
	// Validate user credentials
	user, err := uc.userValidator.ValidateUserForLogin(ctx, credentials.Email, credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user for login: %w", err)
	}

	// Generate session and tokens. The refresh token expires with the session, so the
	// session's absolute lifetime is decided here and handed to the signer. Signing touches
	// nothing outside this function, so it is done before the transaction opens.
	sessionID := utils.GenerateUUID()
	expiry := uc.sessionExpiry.For(credentials.RememberMe)
	tokenPair, err := uc.jwtService.GenerateTokenPair(
		user.ID,
		user.Name,
		user.Email,
		sessionID,
		deviceInfo.DeviceID,
		expiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// Stamping the login and creating the session commit together: a failure between them
	// would otherwise leave a session the user was never told about, or a last_login_at with
	// no session behind it.
	session := uc.createUserSession(sessionID, user.ID, tokenPair.RefreshTokenID, deviceInfo, expiry)
	var createdSession *UserSession
	err = uc.txManager.Do(ctx, func(txCtx context.Context) error {
		repo := uc.authRepo.WithTx(txCtx)

		if err := repo.UpdateUserLoginInfo(txCtx, user.ID, true); err != nil {
			return fmt.Errorf("failed to update user login info: %w", err)
		}

		createdSession, err = repo.CreateSession(txCtx, session)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Cache writes happen only after the commit, so a rolled-back login cannot leave the
	// cache describing a session that does not exist.
	if err := uc.permissionService.InvalidateUserPermissions(ctx, user.ID); err != nil {
		logger.Error(ctx, err)
	}

	menu, permissions, err := uc.buildMenuAndPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:       user,
		Menu:       menu,
		Permission: permissions,
		Tokens:     tokenPair,
		Session:    createdSession,
	}, nil
}

// buildMenuAndPermissions resolves the user's effective permissions and the menu tree they
// may see. Login and refresh both return it, so the client never has to reconcile the two.
func (uc *authUseCase) buildMenuAndPermissions(ctx context.Context, userID string) ([]Menu, []string, error) {
	menus, err := uc.authRepo.GetParentMenus(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user menus: %w", err)
	}

	userRoles, err := uc.authRepo.GetUserRolesByUserID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	rolePermissions, err := uc.authRepo.GetRolePermissionsByRoleIDs(ctx, userRoles)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	userPermissionOverrides, err := uc.authRepo.GetUserPermissionOverrides(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user permission overrides: %w", err)
	}

	permissions := uc.mergePermissions(rolePermissions, userPermissionOverrides)
	menuTree := uc.menuService.BuildMenuTree(ctx, menus)

	return uc.filterMenuTreeByPermissions(menuTree, permissions), permissions, nil
}

// Logout revokes the caller's session, which is what invalidates both of its tokens: the
// access token stops passing the session check and the refresh token can no longer be
// exchanged. Other devices keep their own sessions.
// Note: Authentication is handled by middleware, userID and sessionID come from context.
func (uc *authUseCase) Logout(ctx context.Context, userID string, sessionID string) error {
	// Logging out twice is not an error: an already-revoked session is the state the caller
	// asked for. The cache is evicted either way, so a stale entry cannot outlive the call.
	if err := uc.authRepo.RevokeSession(ctx, userID, sessionID, RevokedReasonLogout); err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("revoking session: %w", err)
	}

	uc.sessionService.Evict(ctx, sessionID)

	return nil
}

// LogoutAll invalidates all tokens for a user (logout from all devices)
// Note: Authentication is handled by middleware, userID comes from context
func (uc *authUseCase) LogoutAll(ctx context.Context, userID string) error {
	// Sessions are deactivated, not deleted, to keep an audit trail
	if err := uc.authRepo.DeactivateUserSessions(ctx, userID, RevokedReasonLogoutAll); err != nil {
		return fmt.Errorf("failed to deactivate user sessions: %w", err)
	}

	uc.sessionService.EvictUser(ctx, userID)

	return nil
}

// RefreshToken generates new access token using refresh token
// Note: Token validation is handled by AuthenticateRefreshToken middleware
func (uc *authUseCase) RefreshToken(ctx context.Context, userID, sessionID, refreshJTI string, deviceInfo *DeviceInfo) (*LoginResult, error) {
	// Validate user is still allowed to refresh (not locked/disabled)
	user, err := uc.userValidator.ValidateUserForRefresh(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user for refresh: %w", err)
	}

	session, err := uc.rotateRefreshToken(ctx, userID, sessionID, refreshJTI)
	if err != nil {
		return nil, err
	}

	tokenPair, err := uc.issueTokensForSession(user, session, deviceInfo)
	if err != nil {
		return nil, err
	}

	// Refresh with fresh permissions; the cache refills on the next check
	if err := uc.permissionService.InvalidateUserPermissions(ctx, user.ID); err != nil {
		logger.Error(ctx, err)
	}

	menu, permissions, err := uc.buildMenuAndPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:       user,
		Menu:       menu,
		Permission: permissions,
		Tokens:     tokenPair,
		Session:    session,
	}, nil
}

// rotateRefreshToken claims the presented refresh token and returns the session the new tokens
// belong to. The returned session always carries the refresh_jti the caller must sign.
//
// The session was already checked by the refresh middleware, so a missing, revoked, or expired
// session has become a plain 401 before we get here and is never mistaken for token reuse.
func (uc *authUseCase) rotateRefreshToken(ctx context.Context, userID, sessionID, refreshJTI string) (*UserSession, error) {
	newJTI := utils.GenerateUUID()

	err := uc.authRepo.RotateRefreshJTI(ctx, sessionID, refreshJTI, newJTI)
	if err == nil {
		session, err := uc.authRepo.GetSessionByID(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("reading rotated session: %w", err)
		}
		if session == nil {
			return nil, utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
		}
		return session, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("rotating refresh token: %w", err)
	}

	return uc.resolveFailedRotation(ctx, userID, sessionID, refreshJTI)
}

// resolveFailedRotation decides what a rotation that matched no row actually means. The copy
// of the session loaded earlier may predate a concurrent rotation, so it is re-read from the
// repository (never the cache) after the failed UPDATE, which PostgreSQL made wait for the
// concurrent commit.
func (uc *authUseCase) resolveFailedRotation(ctx context.Context, userID, sessionID, refreshJTI string) (*UserSession, error) {
	session, err := uc.authRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("re-reading session after failed rotation: %w", err)
	}
	if session == nil || !session.IsValidSession() {
		return nil, utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	// The token we were handed is the one the previous rotation just replaced, and that
	// rotation is recent: this is a concurrent tab or a lost response, not an attack. Hand back
	// the session's current refresh token rather than rotating again, so both callers converge
	// on the same token.
	if uc.withinReuseGrace(session, refreshJTI) {
		return session, nil
	}

	// Anything else means a refresh token was presented that this session no longer accepts and
	// cannot explain. Revoke this login only; the user's other devices are untouched.
	if err := uc.authRepo.RevokeSession(ctx, userID, sessionID, RevokedReasonReuseDetected); err != nil && !errors.Is(err, ErrNotFound) {
		logger.Error(ctx, fmt.Errorf("revoking session after refresh token reuse: %w", err))
	}
	uc.sessionService.Evict(ctx, sessionID)

	logger.Warn(ctx, fmt.Sprintf(
		"refresh token reuse detected: session %s of user %s revoked", sessionID, userID))

	return nil, utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
}

// withinReuseGrace reports whether refreshJTI is the token the last rotation replaced, and
// that rotation happened recently enough to treat replaying it as harmless.
func (uc *authUseCase) withinReuseGrace(session *UserSession, refreshJTI string) bool {
	if session.PreviousRefreshJTI == "" || session.PreviousRefreshJTI != refreshJTI {
		return false
	}
	if session.RotatedAt == nil {
		return false
	}
	return utils.Now().Sub(*session.RotatedAt) <= uc.refreshReuseGrace
}

// issueTokensForSession mints an access token and a refresh token carrying the session's
// current refresh_jti. The refresh token expires with the session, so refreshing never
// extends the login.
func (uc *authUseCase) issueTokensForSession(user *User, session *UserSession, deviceInfo *DeviceInfo) (*jwt.TokenPair, error) {
	accessToken, accessExpiresAt, err := uc.jwtService.GenerateAccessToken(
		user.ID, user.Name, user.Email, session.ID, deviceInfo.DeviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := uc.jwtService.SignRefreshToken(
		user.ID, session.ID, deviceInfo.DeviceID, session.RefreshJTI, session.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("signing refresh token: %w", err)
	}

	return uc.buildTokenPair(accessToken, accessExpiresAt, refreshToken, session.ExpiresAt), nil
}

// createUserSession creates a new user session with device information.
// refreshJTI is the jti of the refresh token this session starts with; rotation (T3.3)
// replaces it on every refresh.
func (uc *authUseCase) createUserSession(sessionID, userID, refreshJTI string, deviceInfo *DeviceInfo, expiry time.Duration) *UserSession {
	now := utils.Now()

	return &UserSession{
		ID:         sessionID,
		UserID:     userID,
		RefreshJTI: refreshJTI,
		DeviceName: deviceInfo.DeviceName,
		DeviceType: deviceInfo.DeviceType,
		DeviceID:   deviceInfo.DeviceID,
		IPAddress:  deviceInfo.IPAddress,
		UserAgent:  deviceInfo.UserAgent,
		IsActive:   true,
		ExpiresAt:  now.Add(expiry),
		LastUsedAt: now,
	}
}

// buildTokenPair creates a jwt.TokenPair from access and refresh token details
func (uc *authUseCase) buildTokenPair(
	accessToken string,
	accessExpiry time.Time,
	refreshToken string,
	refreshExpiry time.Time,
) *jwt.TokenPair {
	return &jwt.TokenPair{
		AccessToken:           accessToken,
		AccessTokenType:       "Bearer",
		AccessTokenExpiresIn:  int64(time.Until(accessExpiry).Seconds()),
		AccessTokenExpiresAt:  accessExpiry,
		RefreshToken:          refreshToken,
		RefreshTokenType:      "Bearer",
		RefreshTokenExpiresIn: int64(time.Until(refreshExpiry).Seconds()),
		RefreshTokenExpiresAt: refreshExpiry,
	}
}

// filterMenuTreeByPermissions filters menu tree based on user permissions
func (uc *authUseCase) filterMenuTreeByPermissions(menuTree []Menu, userPermissions []string) []Menu {
	var filteredMenus []Menu

	// Create a map for faster permission lookup
	permissionMap := make(map[string]bool)
	for _, permission := range userPermissions {
		permissionMap[permission] = true
	}

	for _, menu := range menuTree {
		filteredMenu := uc.filterSingleMenu(menu, permissionMap)
		if filteredMenu != nil {
			filteredMenus = append(filteredMenus, *filteredMenu)
		}
	}

	return filteredMenus
}

// filterSingleMenu recursively filters a single menu and its children
func (uc *authUseCase) filterSingleMenu(menu Menu, permissionMap map[string]bool) *Menu {
	// Filter children recursively first
	var filteredChildren []Menu
	for _, child := range menu.Children {
		filteredChild := uc.filterSingleMenu(child, permissionMap)
		if filteredChild != nil {
			filteredChildren = append(filteredChildren, *filteredChild)
		}
	}

	// Check if this menu should be displayed
	var shouldDisplay bool

	if len(menu.Permissions) > 0 {
		// Menu has permissions, check if user has any of them
		hasAnyPermission := false
		for _, permissionID := range menu.Permissions {
			if permissionMap[permissionID] {
				hasAnyPermission = true
				break
			}
		}
		shouldDisplay = hasAnyPermission
	} else {
		// Menu doesn't have permissions (usually parent menu)
		// Display only if it has accessible children
		shouldDisplay = len(filteredChildren) > 0
	}

	if shouldDisplay {
		// Return the menu with filtered children
		filteredMenu := menu
		filteredMenu.Children = filteredChildren
		return &filteredMenu
	}

	return nil
}

// mergePermissions merges role permissions with user permission overrides
// Returns final permission list after applying user-specific grants and revocations
func (uc *authUseCase) mergePermissions(rolePermissions []string, userOverrides map[string]bool) []string {
	// Start with role permissions as a set for efficient lookup
	permissionSet := make(map[string]bool)
	for _, permission := range rolePermissions {
		permissionSet[permission] = true
	}

	// Apply user permission overrides
	for permission, isGranted := range userOverrides {
		if isGranted {
			// Grant permission (add to set)
			permissionSet[permission] = true
		} else {
			// Revoke permission (remove from set)
			delete(permissionSet, permission)
		}
	}

	// Convert set back to slice
	finalPermissions := make([]string, 0, len(permissionSet))
	for permission := range permissionSet {
		finalPermissions = append(finalPermissions, permission)
	}

	return finalPermissions
}
