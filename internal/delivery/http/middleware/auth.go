package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"goilerplate/config"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/constants"
	jwtService "goilerplate/pkg/jwt"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/response"
	"goilerplate/pkg/utils"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Auth struct {
	jwtService        *jwtService.JWTService
	authRepository    auth.Repository
	cacheService      *auth.CacheService
	permissionService *auth.PermissionService
	apikeys           []config.Apikey
}

func NewAuth(jwtService *jwtService.JWTService, authRepository auth.Repository, cacheService *auth.CacheService, permissionService *auth.PermissionService, apikeys []config.Apikey) *Auth {
	return &Auth{
		jwtService:        jwtService,
		authRepository:    authRepository,
		cacheService:      cacheService,
		permissionService: permissionService,
		apikeys:           apikeys,
	}
}

// Authenticate provides authentication for standard users (validates ACCESS tokens)
func (m *Auth) Authenticate() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Extract and validate token
		token, claims, err := m.validateAuthHeader(ctx)
		if err != nil {
			return response.HandleError(ctx, err)
		}

		// Verify token is not blacklisted and exists in storage
		tokenHash := m.hashToken(token)
		_, err = m.verifyTokenValidity(ctx, tokenHash)
		if err != nil {
			return response.HandleError(ctx, err)
		}

		// Mark token as used for audit trail (async)
		m.markTokenAsUsedAsync(tokenHash, ctx.UserContext())

		// Set user context
		m.setUserContext(ctx, claims.UserID, claims.UserName, claims.SessionID, tokenHash)

		return ctx.Next()
	}
}

// AuthenticateRefreshToken provides authentication specifically for refresh token endpoint
// This validates REFRESH tokens, not access tokens
func (m *Auth) AuthenticateRefreshToken() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var clientError *utils.ClientError

		// Extract token from Authorization header
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(ctx, "Authorization header missing")
		}

		token, err := m.extractBearerToken(authHeader)
		if err != nil {
			return response.Unauthorized(ctx, "Invalid authorization format")
		}

		// Validate token and get claims
		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			return response.Unauthorized(ctx, "Invalid or expired token")
		}

		// IMPORTANT: Check that this is a REFRESH token, not an access token
		if claims.Type != jwtService.RefreshToken {
			return response.Unauthorized(ctx, "Invalid token type: expected refresh token")
		}

		// Hash token for storage lookup
		tokenHash := m.hashToken(token)

		// Verify token exists in storage and is not blacklisted
		userToken, err := m.verifyTokenValidity(ctx, tokenHash)
		if err != nil {
			if errors.As(err, &clientError) {
				return response.CustomError(ctx, clientError.Code, clientError.Error(), nil)
			}
			logger.Error(ctx.UserContext(), err)
			return response.InternalServerError(ctx, "")
		}

		// Set context for handler to use
		m.setUserContext(ctx, claims.UserID, claims.UserName, claims.SessionID, tokenHash)

		ctx.Locals("refresh_token", token)
		ctx.Locals("refresh_token_expires_at", userToken.ExpiresAt)

		return ctx.Next()
	}
}

// RequiredPermission checks if the authenticated user has the specified permission
// This middleware should be used after Authenticate() middleware
//
// Cache Strategy (like previous implementation):
// - If Redis is ENABLED: Check individual permission cache first
// - If not found in cache or Redis disabled: Use GetUserFinalPermissions (which checks cache for full list)
//
// Permission check priority:
// 1. User-specific permission override (user_permissions)
//   - is_granted = true: custom grant (user has permission)
//   - is_granted = false: revoked (user doesn't have permission)
//
// 2. Role-based permissions (user -> roles -> role_permissions)
// 3. Menu-based permissions (user -> roles -> role_menus -> menus + children -> menu_permissions)
func (m *Auth) RequiredPermission(permission string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Get user ID from context (set by Authenticate middleware)
		userIDStr, ok := ctx.Locals(string(constants.ContextKeyUserID)).(string)
		if !ok || userIDStr == "" {
			return response.Unauthorized(ctx, constants.MsgUnauthorized)
		}

		// Use PermissionService to check permission (handles cache + database fallback automatically)
		// This will try cache first, then fallback to database if cache miss
		hasPermission, err := m.permissionService.HasPermission(ctx.UserContext(), userIDStr, permission)
		if err != nil {
			logger.Error(ctx.UserContext(), err)
			return response.InternalServerError(ctx, "")
		}

		if !hasPermission {
			return response.Forbidden(ctx, fmt.Sprintf("'%s' permission required", permission))
		}

		return ctx.Next()
	}
}

// InternalAuthenticate provides authentication for internal services
func (m *Auth) InternalAuthenticate() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID := "system"
		userName := "system"

		userIdCtx := context.WithValue(ctx.UserContext(), constants.ContextKeyUserID, userID)
		userNameCtx := context.WithValue(userIdCtx, constants.ContextKeyUserName, userName)
		ctx.SetUserContext(userNameCtx)

		// Set in Locals (for Fiber context usage)
		ctx.Locals(string(constants.ContextKeyUserID), userID)
		ctx.Locals(string(constants.ContextKeyUserName), userName)

		return ctx.Next()
	}
}

// PartnerAuthenticate provides authentication for partner services
func (m *Auth) PartnerAuthenticate() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		apiKey := ctx.Get("x-api-key")
		if apiKey == "" {
			return response.Unauthorized(ctx, "")
		}

		isValid := false
		userID := ""
		userName := ""
		for _, k := range m.apikeys {
			if apiKey == k.Key {
				isValid = true
				userID = k.Key
				userName = k.Name
				break
			}
		}

		if !isValid {
			return response.Unauthorized(ctx, "")
		}

		userIdCtx := context.WithValue(ctx.UserContext(), constants.ContextKeyUserID, userID)
		userNameCtx := context.WithValue(userIdCtx, constants.ContextKeyUserName, userName)
		ctx.SetUserContext(userNameCtx)

		// Set in Locals (for Fiber context usage)
		ctx.Locals(string(constants.ContextKeyUserID), userID)
		ctx.Locals(string(constants.ContextKeyUserName), userName)

		return ctx.Next()
	}
}

// validateAuthHeader extracts and validates the authorization header
func (m *Auth) validateAuthHeader(ctx *fiber.Ctx) (string, *jwtService.Claims, error) {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return "", nil, utils.Error(http.StatusUnauthorized, "Unauthorized")
	}

	token, err := m.extractBearerToken(authHeader)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract bearer token: %w", err)
	}

	claims, err := m.jwtService.ValidateToken(token)
	if err != nil {
		return "", nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if claims.Type != jwtService.AccessToken {
		return "", nil, utils.Error(http.StatusUnauthorized, "Invalid token")
	}

	return token, claims, nil
}

// verifyTokenValidity checks if token is blacklisted and exists in storage
func (m *Auth) verifyTokenValidity(ctx *fiber.Ctx, tokenHash string) (*auth.UserToken, error) {
	// Check blacklist first to prevent race conditions
	if m.cacheService.IsEnabled() {
		isBlacklisted, err := m.cacheService.IsTokenBlacklisted(ctx.UserContext(), tokenHash)
		if err != nil {
			return nil, fmt.Errorf("failed to check token blacklist: %w", err)
		}
		if isBlacklisted {
			return nil, utils.Error(http.StatusUnauthorized, "Unauthorized")
		}
	}

	// Get token from cache or database
	var userToken *auth.UserToken
	var err error

	if m.cacheService.IsEnabled() {
		userToken, err = m.cacheService.GetToken(ctx.UserContext(), tokenHash)
	} else {
		userToken, err = m.authRepository.GetTokenByHash(ctx.UserContext(), tokenHash)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}

	if userToken == nil {
		return nil, utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	// Check expiration if not using Redis cache
	if !m.cacheService.IsEnabled() && userToken.IsExpired() {
		return nil, utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	return userToken, nil
}

// markTokenAsUsedAsync marks the token as used in the background for audit trail
func (m *Auth) markTokenAsUsedAsync(tokenHash string, userCtx context.Context) {
	go func(hash string, ctx context.Context) {
		if err := m.authRepository.MarkTokenAsUsed(ctx, hash); err != nil {
			logger.Error(ctx, err)
		}
	}(tokenHash, userCtx)
}

// setUserContext sets user information in both Fiber context and Go context
func (m *Auth) setUserContext(ctx *fiber.Ctx, userID, userName, sessionID, tokenHash string) {
	userIdCtx := context.WithValue(ctx.UserContext(), constants.ContextKeyUserID, userID)
	userNameCtx := context.WithValue(userIdCtx, constants.ContextKeyUserName, userName)
	sessionIDCtx := context.WithValue(userNameCtx, constants.ContextKeySessionID, sessionID)
	userTokenHashCtx := context.WithValue(sessionIDCtx, constants.ContextTokenHash, tokenHash)
	ctx.SetUserContext(userTokenHashCtx)

	ctx.Locals(string(constants.ContextKeyUserID), userID)
	ctx.Locals(string(constants.ContextKeyUserName), userName)
	ctx.Locals(string(constants.ContextKeySessionID), sessionID)
	ctx.Locals(string(constants.ContextTokenHash), tokenHash)
}

// extractBearerToken extracts JWT token from Authorization header
func (m *Auth) extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	return token, nil
}

// hashToken creates a SHA256 hash of the token for secure storage
func (m *Auth) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
