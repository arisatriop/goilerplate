package middleware

import (
	"context"
	"errors"
	"fmt"
	"goilerplate/config"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/apikey"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/hash"
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
	apikeys           *apikey.Registry
	// internalSecret is the digest of internal_auth.secret, or "" when the mode is none. Empty
	// means /internal is open to whatever reaches it, which is what D5 leaves to the gateway.
	internalSecret string
}

func NewAuth(jwtService *jwtService.JWTService, authRepository auth.Repository, sessionService *auth.SessionService, permissionService *auth.PermissionService, apikeys map[string]string, internalAuth config.InternalAuth) *Auth {
	internalSecret := ""
	if internalAuth.RequiresSecret() {
		internalSecret = hash.Token(internalAuth.Secret)
	}

	return &Auth{
		jwtService:        jwtService,
		authRepository:    authRepository,
		sessionService:    sessionService,
		permissionService: permissionService,
		// Built once, so the plaintext keys are reduced to digests at startup instead of being
		// held in memory for the life of the process.
		apikeys:        apikey.NewRegistry(apikeys),
		internalSecret: internalSecret,
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

		token, err := bearerToken(ctx)
		if err != nil {
			return response.Unauthorized(ctx, "")
		}

		// Validate token and get claims. This accepts refresh tokens only, so an access
		// token presented here fails to verify.
		claims, err := m.jwtService.ValidateRefreshToken(token)
		if err != nil {
			return response.Unauthorized(ctx, "")
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

// defaultInternalCaller is used when a caller does not name itself.
const defaultInternalCaller = "system"

// maxServiceNameLength bounds what an unverified header can put into every log line of a request.
const maxServiceNameLength = 40

// InternalAuthenticate guards the /internal routes, which are meant for pod-to-pod traffic only
// (D5). Keeping them off the public internet is the gateway's job: it forwards an explicit
// allowlist and never a catch-all.
//
// In shared_secret mode a caller must also present X-Internal-Secret. That is a second lock for
// when the gateway config belongs to another team, or changes often enough that one day it will
// be wrong — not a replacement for the allowlist, since the secret is shared by every caller and
// travels on every request.
func (m *Auth) InternalAuthenticate() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if m.internalSecret != "" && !hash.Equal(ctx.Get(constants.HeaderInternalSecret), m.internalSecret) {
			return response.Unauthorized(ctx, "")
		}

		caller := internalCallerName(ctx.Get(constants.HeaderServiceName))

		userIdCtx := context.WithValue(ctx.UserContext(), constants.ContextKeyUserID, caller)
		userNameCtx := context.WithValue(userIdCtx, constants.ContextKeyUserName, caller)
		ctx.SetUserContext(userNameCtx)

		// Set in Locals (for Fiber context usage)
		ctx.Locals(string(constants.ContextKeyUserID), caller)
		ctx.Locals(string(constants.ContextKeyUserName), caller)

		return ctx.Next()
	}
}

// internalCallerName sanitises the X-Service-Name header.
//
// The name is an unverified claim — in shared_secret mode every caller holds the same secret, so
// nothing distinguishes one from another. It is for attribution in logs, never for authorization.
// It is bounded and restricted to a plain character set so that a header cannot pad every log
// line a request produces.
func internalCallerName(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return defaultInternalCaller
	}

	cleaned := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '-', r == '_', r == '.':
			return r
		default:
			return -1
		}
	}, header)

	if cleaned == "" {
		return defaultInternalCaller
	}
	if len(cleaned) > maxServiceNameLength {
		cleaned = cleaned[:maxServiceNameLength]
	}

	return cleaned
}

// PartnerAuthenticate provides authentication for partner services.
//
// The partner's identity is its configured name. The key itself goes no further than this
// function: it used to be stored as the user ID, which put a live credential into every log
// line the request produced and into any cache keyed on the caller.
func (m *Auth) PartnerAuthenticate() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		apiKey := ctx.Get(constants.HeaderAPIKey)
		if apiKey == "" {
			return response.Unauthorized(ctx, "")
		}

		name, ok := m.apikeys.Lookup(apiKey)
		if !ok {
			return response.Unauthorized(ctx, "")
		}

		// A partner has no identifier separate from its name, so both carry the name. What
		// matters is that neither carries the key.
		userIdCtx := context.WithValue(ctx.UserContext(), constants.ContextKeyUserID, name)
		userNameCtx := context.WithValue(userIdCtx, constants.ContextKeyUserName, name)
		ctx.SetUserContext(userNameCtx)

		// Set in Locals (for Fiber context usage)
		ctx.Locals(string(constants.ContextKeyUserID), name)
		ctx.Locals(string(constants.ContextKeyUserName), name)

		return ctx.Next()
	}
}

// validateAuthHeader extracts and validates the authorization header
func (m *Auth) validateAuthHeader(ctx *fiber.Ctx) (*jwtService.Claims, error) {
	token, err := bearerToken(ctx)
	if err != nil {
		return nil, err
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

// bearerToken reads the token out of the Authorization header.
//
// Every failure answers the same way. Telling a caller whether the header was missing, not a
// Bearer scheme, or empty after the scheme describes our parser, not their mistake, and the one
// thing it reliably tells an attacker is which of their guesses got further.
func bearerToken(ctx *fiber.Ctx) (string, error) {
	parts := strings.SplitN(ctx.Get(fiber.HeaderAuthorization), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	return token, nil
}
