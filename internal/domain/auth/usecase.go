package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"goilerplate/pkg/auth"
	"goilerplate/pkg/jwt"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
	"time"
)

const (
	sessionDuration    = 7 * 24 * time.Hour  // 7 days default
	rememberMeDuration = 30 * 24 * time.Hour // 30 days for remember me
)

type authUseCase struct {
	authRepo          Repository
	jwtService        *jwt.JWTService
	tokenService      *TokenService
	userValidator     *UserValidator
	tokenStorage      *TokenStorage
	menuService       *MenuService
	cacheService      *CacheService
	permissionService *PermissionService
}

// Usecase defines the authentication use case interface
type Usecase interface {
	Register(ctx context.Context, entity *User) error
	Login(ctx context.Context, credentials *LoginCredentials, deviceInfo *DeviceInfo) (*LoginResult, error)
	Logout(ctx context.Context, userID string, tokenHash string, sessionID string) error
	LogoutAll(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, userID string, sessionID string, tokenHash string, refreshToken string, refreshTokenExpiresAt time.Time, deviceInfo *DeviceInfo) (*LoginResult, error)
}

func NewUseCase(authRepo Repository, jwtService *jwt.JWTService, cacheService *CacheService) Usecase {
	tokenService := NewTokenService(jwtService, authRepo, cacheService)
	userValidator := NewUserValidator(authRepo)
	tokenStorage := NewTokenStorage(authRepo, cacheService)
	menuService := NewMenuService(authRepo)
	permissionService := NewPermissionService(authRepo, cacheService)

	return &authUseCase{
		authRepo:          authRepo,
		jwtService:        jwtService,
		tokenService:      tokenService,
		userValidator:     userValidator,
		tokenStorage:      tokenStorage,
		menuService:       menuService,
		cacheService:      cacheService,
		permissionService: permissionService,
	}
}

// Register creates a new user account
func (uc *authUseCase) Register(ctx context.Context, entity *User) error {
	existingUser, err := uc.authRepo.GetUserByEmail(ctx, entity.Email)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}
	if existingUser != nil {
		return utils.ErrEmailAlreadyExists
	}

	hashedPassword, err := auth.HashPassword(entity.PasswordHash)
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

	// Generate session and tokens
	sessionID := utils.GenerateUUID()
	tokenPair, err := uc.jwtService.GenerateTokenPair(
		user.ID,
		user.Name,
		user.Email,
		sessionID,
		deviceInfo.DeviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// Create user session
	session := uc.createUserSession(sessionID, user.ID, tokenPair.RefreshToken, deviceInfo, credentials.RememberMe)
	createdSession, err := uc.authRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Store tokens in database
	err = uc.tokenStorage.StoreTokenPair(ctx, user.ID, sessionID, tokenPair, deviceInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to store tokens: %w", err)
	}

	// Get parent menus from role_menus
	roleMenus, err := uc.authRepo.GetRoleMenus(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch role menus: %w", err)
	}

	// Get parent menus from user_menus
	userMenus, err := uc.authRepo.GetUserMenus(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user menus: %w", err)
	}

	// Build hierarchical menu trees (fetches children on-demand)
	roleMenuTree := uc.menuService.BuildMenuTree(ctx, roleMenus)
	userMenuTree := uc.menuService.BuildMenuTree(ctx, userMenus)

	// Cache session to Redis if enabled
	if uc.cacheService.IsEnabled() {
		if err := uc.cacheService.CacheSession(ctx, createdSession); err != nil {
			return nil, fmt.Errorf("failed to cache session to Redis: %w", err)
		}

		if err := uc.permissionService.CacheAllUserPermissions(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("failed to cache user permissions to Redis: %w", err)
		}
	}

	return &LoginResult{
		User:        user,
		Menus:       roleMenuTree,
		CustomMenus: userMenuTree,
		Tokens:      tokenPair,
		Session:     createdSession,
	}, nil
}

// Logout invalidates both access and refresh tokens for the current user session
// Note: Authentication is handled by middleware, userID, tokenHash, and sessionID come from context
func (uc *authUseCase) Logout(ctx context.Context, userID string, tokenHash string, sessionID string) error {
	// Delete tokens (no need to validate - already done in middleware)
	return uc.tokenService.DeleteTokens(ctx, tokenHash, userID, sessionID)
}

// LogoutAll invalidates all tokens for a user (logout from all devices)
// Note: Authentication is handled by middleware, userID comes from context
func (uc *authUseCase) LogoutAll(ctx context.Context, userID string) error {

	// Step 1: Blacklist all tokens atomically BEFORE deletion
	// This prevents race conditions where tokens might still be used during deletion
	if err := uc.blacklistAllUserTokens(ctx, userID); err != nil {
		// Critical: If Redis is enabled and blacklisting fails, we must abort
		// Otherwise tokens would be deleted from DB but still cached and usable
		if uc.cacheService.IsEnabled() {
			return fmt.Errorf("failed to blacklist user tokens: %w", err)
		}
		// If Redis is not enabled, continue (DB is source of truth)
	}

	// Step 2: Delete tokens from database (permanent removal)
	if err := uc.authRepo.DeleteUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user tokens: %w", err)
	}

	// Step 3: Deactivate sessions (preserve for audit trail)
	// Sessions are not deleted, just marked as inactive for compliance
	if err := uc.authRepo.DeactivateUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to deactivate user sessions: %w", err)
	}

	// Step 4: Clean up cache entries
	// Cache cleanup failures are not critical since DB is already updated
	if err := uc.deleteUserTokensFromCache(ctx, userID); err != nil {
		// Continue - cache will expire naturally, DB is source of truth
	}

	if err := uc.deleteUserSessionsFromCache(ctx, userID); err != nil {
		// Continue - cache will expire naturally, DB is source of truth
	}

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

	// Cache permissions if Redis enabled
	// This is critical when Redis is enabled because permissions will be read from Redis
	if uc.cacheService.IsEnabled() {
		if err := uc.permissionService.CacheAllUserPermissions(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("failed to cache user permissions to Redis: %w", err)
		}
	}

	// Create response
	tokenPair := uc.buildTokenPair(
		accessTokenString,
		expiresAt,
		refreshToken,
		refreshTokenExpiresAt,
	)

	session := uc.buildActiveSession(sessionID, user.ID, deviceInfo)

	return &LoginResult{
		User:    user,
		Tokens:  tokenPair,
		Session: session,
	}, nil
}

// createUserSession creates a new user session with device information
func (uc *authUseCase) createUserSession(sessionID, userID, refreshToken string, deviceInfo *DeviceInfo, rememberMe bool) *UserSession {
	expirationDuration := sessionDuration
	if rememberMe {
		expirationDuration = rememberMeDuration
	}

	return &UserSession{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: uc.hashToken(refreshToken),
		DeviceName:       deviceInfo.DeviceName,
		DeviceType:       deviceInfo.DeviceType,
		DeviceID:         deviceInfo.DeviceID,
		IPAddress:        deviceInfo.IPAddress,
		UserAgent:        deviceInfo.UserAgent,
		Location:         deviceInfo.Location,
		IsActive:         true,
		ExpiresAt:        time.Now().Add(expirationDuration),
		LastUsedAt:       time.Now(),
	}
}

// hashToken creates a SHA256 hash of the token for secure storage
func (uc *authUseCase) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
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
	go func(hash string, userCtx context.Context) {
		if err := uc.authRepo.MarkTokenAsUsed(userCtx, hash); err != nil {
			logger.Error(userCtx, err)
		}
	}(tokenHash, ctx)
}

// blacklistAllUserTokens adds all user tokens to blacklist atomically
func (uc *authUseCase) blacklistAllUserTokens(ctx context.Context, userID string) error {
	if !uc.cacheService.IsEnabled() {
		return nil // Redis not enabled, skip blacklisting
	}

	userTokens, err := uc.authRepo.GetUserTokens(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user tokens for blacklisting: %w", err)
	}

	var tokenHashes []string
	var maxTTL time.Duration

	for _, token := range userTokens {
		if ttl := time.Until(token.ExpiresAt); ttl > 0 {
			tokenHashes = append(tokenHashes, token.TokenHash)
			if ttl > maxTTL {
				maxTTL = ttl
			}
		}
	}

	if len(tokenHashes) > 0 {
		if err := uc.cacheService.AddMultipleTokensToBlacklist(ctx, tokenHashes, maxTTL); err != nil {
			return fmt.Errorf("failed to add tokens to blacklist: %w", err)
		}
	}

	return nil
}

// deleteUserTokensFromCache removes all user tokens from Redis
func (uc *authUseCase) deleteUserTokensFromCache(ctx context.Context, userID string) error {
	if !uc.cacheService.IsEnabled() {
		return nil // Redis not enabled, skip cache deletion
	}

	if err := uc.cacheService.DeleteUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user tokens from cache: %w", err)
	}

	return nil
}

func (uc *authUseCase) deleteUserSessionsFromCache(ctx context.Context, userID string) error {
	if !uc.cacheService.IsEnabled() {
		return nil // Redis not enabled, skip cache deletion
	}

	if err := uc.cacheService.DeleteUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user sessions from cache: %w", err)
	}

	return nil
}
