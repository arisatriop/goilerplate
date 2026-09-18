package auth

import (
	"context"
	"fmt"
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
	tokenService      *TokenService
	userValidator     *UserValidator
	tokenStorage      *TokenStorage
	menuService       *MenuService
	sessionService    *SessionService
	permissionService *PermissionService
	sessionExpiry     SessionExpiry
}

// Usecase defines the authentication use case interface
type Usecase interface {
	Register(ctx context.Context, entity *User) error
	Login(ctx context.Context, credentials *LoginCredentials, deviceInfo *DeviceInfo) (*LoginResult, error)
	Logout(ctx context.Context, userID string, tokenHash string, sessionID string) error
	LogoutAll(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, userID string, sessionID string, tokenHash string, refreshToken string, refreshTokenExpiresAt time.Time, deviceInfo *DeviceInfo) (*LoginResult, error)
}

func NewUseCase(
	authRepo Repository,
	jwtService *jwt.JWTService,
	sessionService *SessionService,
	permissionService *PermissionService,
	sessionExpiry SessionExpiry,
) Usecase {
	tokenService := NewTokenService(authRepo)
	userValidator := NewUserValidator(authRepo)
	tokenStorage := NewTokenStorage(authRepo)
	menuService := NewMenuService(authRepo)

	return &authUseCase{
		authRepo:          authRepo,
		jwtService:        jwtService,
		tokenService:      tokenService,
		userValidator:     userValidator,
		tokenStorage:      tokenStorage,
		menuService:       menuService,
		sessionService:    sessionService,
		permissionService: permissionService,
		sessionExpiry:     sessionExpiry,
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

	// Update user login info
	if err := uc.authRepo.UpdateUserLoginInfo(ctx, user.ID, true); err != nil {
		return nil, fmt.Errorf("failed to update user login info: %w", err)
	}

	// Generate session and tokens. The refresh token expires with the session, so the
	// session's absolute lifetime is decided here and handed to the signer.
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

	// Create user session
	session := uc.createUserSession(sessionID, user.ID, tokenPair.RefreshTokenID, deviceInfo, expiry)
	createdSession, err := uc.authRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Store tokens in database
	err = uc.tokenStorage.StoreTokenPair(ctx, user.ID, sessionID, tokenPair, deviceInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to store tokens: %w", err)
	}

	// Start the session with fresh permissions; the cache refills on the next check
	if err := uc.permissionService.InvalidateUserPermissions(ctx, user.ID); err != nil {
		logger.Error(ctx, err)
	}

	menus, err := uc.authRepo.GetParentMenus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user menus: %w", err)
	}

	// Get user roles
	userRoles, err := uc.authRepo.GetUserRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Get permissions for user roles
	rolePermissions, err := uc.authRepo.GetRolePermissionsByRoleIDs(ctx, userRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	// Get user permission overrides
	userPermissionOverrides, err := uc.authRepo.GetUserPermissionOverrides(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permission overrides: %w", err)
	}

	// Merge permissions (role permissions + user overrides)
	finalPermissions := uc.mergePermissions(rolePermissions, userPermissionOverrides)

	// Build complete menu tree
	menuTree := uc.menuService.BuildMenuTree(ctx, menus)

	// Filter menu tree based on final merged permissions
	filteredMenuTree := uc.filterMenuTreeByPermissions(menuTree, finalPermissions)

	return &LoginResult{
		User:       user,
		Menu:       filteredMenuTree,
		Permission: finalPermissions,
		Tokens:     tokenPair,
		Session:    createdSession,
	}, nil
}

// Logout invalidates both access and refresh tokens for the current user session
// Note: Authentication is handled by middleware, userID, tokenHash, and sessionID come from context
func (uc *authUseCase) Logout(ctx context.Context, userID string, tokenHash string, sessionID string) error {
	// Delete tokens (no need to validate - already done in middleware)
	if err := uc.tokenService.DeleteTokens(ctx, tokenHash, userID, sessionID); err != nil {
		return err
	}

	uc.sessionService.Evict(ctx, sessionID)

	return nil
}

// LogoutAll invalidates all tokens for a user (logout from all devices)
// Note: Authentication is handled by middleware, userID comes from context
func (uc *authUseCase) LogoutAll(ctx context.Context, userID string) error {
	if err := uc.authRepo.DeleteUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user tokens: %w", err)
	}

	// Sessions are deactivated, not deleted, to keep an audit trail
	if err := uc.authRepo.DeactivateUserSessions(ctx, userID, RevokedReasonLogoutAll); err != nil {
		return fmt.Errorf("failed to deactivate user sessions: %w", err)
	}

	uc.sessionService.EvictUser(ctx, userID)

	return nil
}

// RefreshToken generates new access token using refresh token
// Note: Token validation is handled by AuthenticateRefreshToken middleware
func (uc *authUseCase) RefreshToken(ctx context.Context, userID string, sessionID string, tokenHash string, refreshToken string, refreshTokenExpiresAt time.Time, deviceInfo *DeviceInfo) (*LoginResult, error) {
	// Validate user is still allowed to refresh (not locked/disabled)
	user, err := uc.userValidator.ValidateUserForRefresh(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user for refresh: %w", err)
	}

	// Generate new access token
	accessTokenString, expiresAt, err := uc.jwtService.GenerateAccessToken(
		user.ID,
		user.Name,
		user.Email,
		sessionID,
		deviceInfo.DeviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	// Store new access token
	err = uc.tokenStorage.StoreAccessToken(ctx, user.ID, accessTokenString, expiresAt, deviceInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to store new access token: %w", err)
	}

	// Mark refresh token as used (async - for audit trail)
	uc.markTokenAsUsedAsync(ctx, tokenHash)

	// Refresh with fresh permissions; the cache refills on the next check
	if err := uc.permissionService.InvalidateUserPermissions(ctx, user.ID); err != nil {
		logger.Error(ctx, err)
	}

	// Create response
	tokenPair := uc.buildTokenPair(
		accessTokenString,
		expiresAt,
		refreshToken,
		refreshTokenExpiresAt,
	)

	session := uc.buildActiveSession(sessionID, user.ID, deviceInfo)

	// Get menus and permissions for consistency
	menus, err := uc.authRepo.GetParentMenus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user menus: %w", err)
	}

	// Get user roles
	userRoles, err := uc.authRepo.GetUserRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Get permissions for user roles
	rolePermissions, err := uc.authRepo.GetRolePermissionsByRoleIDs(ctx, userRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	// Get user permission overrides
	userPermissionOverrides, err := uc.authRepo.GetUserPermissionOverrides(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permission overrides: %w", err)
	}

	// Merge permissions (role permissions + user overrides)
	finalPermissions := uc.mergePermissions(rolePermissions, userPermissionOverrides)

	// Build complete menu tree
	menuTree := uc.menuService.BuildMenuTree(ctx, menus)

	// Filter menu tree based on final merged permissions
	filteredMenuTree := uc.filterMenuTreeByPermissions(menuTree, finalPermissions)

	return &LoginResult{
		User:       user,
		Menu:       filteredMenuTree,
		Permission: finalPermissions,
		Tokens:     tokenPair,
		Session:    session,
	}, nil
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

// buildActiveSession creates a UserSession from device info
// Used when we need to return session info without persisting it first
func (uc *authUseCase) buildActiveSession(sessionID, userID string, deviceInfo *DeviceInfo) *UserSession {
	return &UserSession{
		ID:         sessionID,
		UserID:     userID,
		DeviceID:   deviceInfo.DeviceID,
		DeviceName: deviceInfo.DeviceName,
		DeviceType: deviceInfo.DeviceType,
		IPAddress:  deviceInfo.IPAddress,
		UserAgent:  deviceInfo.UserAgent,
		IsActive:   true,
	}
}

// markTokenAsUsedAsync marks a token as used in background for audit trail
// Failures are logged but don't affect the main flow
func (uc *authUseCase) markTokenAsUsedAsync(ctx context.Context, tokenHash string) {
	bgCtx := context.WithoutCancel(ctx)
	go func() {
		if err := uc.authRepo.MarkTokenAsUsed(bgCtx, tokenHash); err != nil {
			logger.Error(bgCtx, err)
		}
	}()
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
