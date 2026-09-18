package middleware

import (
	"context"
	"errors"
	"fmt"
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

// refreshJTILocal carries the presented refresh token's jti from the middleware to the handler.
const refreshJTILocal = "refresh_jti"

type Auth struct {
	jwtService        *jwtService.JWTService
	authRepository    auth.Repository
	sessionService    *auth.SessionService
	permissionService *auth.PermissionService
	apikeys           map[string]string
}

func NewAuth(jwtService *jwtService.JWTService, authRepository auth.Repository, sessionService *auth.SessionService, permissionService *auth.PermissionService, apikeys map[string]string) *Auth {
	return &Auth{
		jwtService:        jwtService,
		authRepository:    authRepository,
		sessionService:    sessionService,
		permissionService: permissionService,
		apikeys:           apikeys,
	}
}

// Authenticate provides authentication for standard users (validates ACCESS tokens)
func (m *Auth) Authenticate() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Extract and validate token
		claims, err := m.validateAuthHeader(ctx)
		if err != nil {
			return response.HandleError(ctx, err)
		}

		// The access token is stateless: nothing about it is stored, so there is no per-request
		// read or write of it. Revocation is the session's job.
		if err := m.sessionService.EnsureActiveForRequest(ctx.UserContext(), claims.SessionID); err != nil {
			return response.HandleError(ctx, err)
		}

		// Set user context
		m.setUserContext(ctx, claims.UserID, claims.UserName, claims.SessionID)

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

		// Validate token and get claims. This accepts refresh tokens only, so an access
		// token presented here fails to verify.
		claims, err := m.jwtService.ValidateRefreshToken(token)
		if err != nil {
			return response.Unauthorized(ctx, "Invalid or expired token")
		}

		// Refresh always checks the session, whatever auth.revocation says: this is the point
		// where a revoked login must stop being able to mint new access tokens.
		if _, err := m.sessionService.GetActive(ctx.UserContext(), claims.SessionID); err != nil {
			if errors.As(err, &clientError) {
				return response.CustomError(ctx, clientError.Code, clientError.Error(), nil)
			}
			logger.Error(ctx.UserContext(), err)
			return response.InternalServerError(ctx, "")
		}

		// Set context for handler to use
		m.setUserContext(ctx, claims.UserID, claims.UserName, claims.SessionID)

		// The jti is what the session tracks and rotates; the raw token has no further use
		// now that refresh always issues a replacement.
		ctx.Locals(refreshJTILocal, claims.ID)

		return ctx.Next()
	}
}

// RequiredPermission checks if the authenticated user has the specified permission
// This middleware should be used after Authenticate() middleware
//
// Permissions are read through the permission cache (auth.session_cache mode) and fall back
// to the database on a miss.
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
		for name, key := range m.apikeys {
			if apiKey == key {
				isValid = true
				userID = key
				userName = name
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
func (m *Auth) validateAuthHeader(ctx *fiber.Ctx) (*jwtService.Claims, error) {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return nil, utils.ClientErr(http.StatusUnauthorized, "Unauthorized")
	}

	token, err := m.extractBearerToken(authHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract bearer token: %w", err)
	}

	// ValidateAccessToken pins the algorithm, issuer, audience, signing key and token type,
	// so a refresh token presented here is rejected by the parser rather than by a later check.
	claims, err := m.jwtService.ValidateAccessToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	return claims, nil
}

// setUserContext sets user information in both Fiber context and Go context
func (m *Auth) setUserContext(ctx *fiber.Ctx, userID, userName, sessionID string) {
	userIdCtx := context.WithValue(ctx.UserContext(), constants.ContextKeyUserID, userID)
	userNameCtx := context.WithValue(userIdCtx, constants.ContextKeyUserName, userName)
	sessionIDCtx := context.WithValue(userNameCtx, constants.ContextKeySessionID, sessionID)
	ctx.SetUserContext(sessionIDCtx)

	ctx.Locals(string(constants.ContextKeyUserID), userID)
	ctx.Locals(string(constants.ContextKeyUserName), userName)
	ctx.Locals(string(constants.ContextKeySessionID), sessionID)
}

// extractBearerToken extracts JWT token from Authorization header
func (m *Auth) extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	return token, nil
}
