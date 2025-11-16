package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/jwt"
	"goilerplate/pkg/utils"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// TokenService handles token validation and operations
type TokenService struct {
	jwtService   *jwt.JWTService
	authRepo     Repository
	cacheService *CacheService
}

// NewTokenService creates a new token service
func NewTokenService(jwtService *jwt.JWTService, authRepo Repository, cacheService *CacheService) *TokenService {
	return &TokenService{
		jwtService:   jwtService,
		authRepo:     authRepo,
		cacheService: cacheService,
	}
}

// ValidateAndGetClaims validates a token and returns claims
// This performs FULL validation including blacklist and database checks
func (ts *TokenService) ValidateAndGetClaims(ctx context.Context, authHeader string) (*jwt.Claims, error) {
	// Extract Bearer token
	tokenString, err := ts.extractBearerToken(authHeader)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization header: %w", err)
	}

	// Validate JWT signature and claims
	claims, err := ts.jwtService.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Hash token for database/cache lookup
	tokenHash := ts.hashToken(tokenString)

	// Check if token is blacklisted (immediate revocation check)
	if ts.cacheService.IsEnabled() {
		isBlacklisted, err := ts.cacheService.IsTokenBlacklisted(ctx, tokenHash)
		if err != nil {
			return nil, fmt.Errorf("failed to check token blacklist: %w", err)
		}
		if isBlacklisted {
			return nil, utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
		}
	}

	// Verify token exists in storage
	var userToken *UserToken
	if ts.cacheService.IsEnabled() {
		userToken, err = ts.cacheService.GetToken(ctx, tokenHash)
	} else {
		userToken, err = ts.authRepo.GetTokenByHash(ctx, tokenHash)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	if userToken == nil {
		return nil, utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	// Check if token is expired
	if userToken.IsExpired() {
		return nil, utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	return claims, nil
}

// ValidateRefreshTokenAndGetUser validates refresh token and returns user token and claims
func (ts *TokenService) ValidateRefreshTokenAndGetUser(ctx context.Context, authHeader string) (*UserToken, *jwt.Claims, error) {
	// Extract refresh token
	refreshTokenString, err := ts.extractBearerToken(authHeader)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid authorization header: %w", err)
	}

	// Validate refresh token
	claims, err := ts.jwtService.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check token type
	if claims.Type != jwt.RefreshToken {
		return nil, nil, utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	// Verify token exists in database
	refreshTokenHash := ts.hashToken(refreshTokenString)

	// Check if token is blacklisted (immediate revocation check)
	if ts.cacheService.IsEnabled() {
		isBlacklisted, err := ts.cacheService.IsTokenBlacklisted(ctx, refreshTokenHash)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to check token blacklist: %w", err)
		}
		if isBlacklisted {
			return nil, nil, utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
		}
	}

	userToken, err := ts.authRepo.GetTokenByHash(ctx, refreshTokenHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify refresh token: %w", err)
	}

	if userToken == nil {
		return nil, nil, utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	// Check if token is expired
	if userToken.IsExpired() {
		return nil, nil, utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	// Note: We allow refresh tokens to be reused multiple times until expiration
	// This is a design decision - if you need single-use refresh tokens, uncomment below:
	// if userToken.IsUsed() {
	// 	return nil, nil, ErrTokenAlreadyUsed
	// }

	return userToken, claims, nil
}

// DeleteTokens deletes access token and session tokens (refresh token) from both database and Redis
// This only deletes tokens for the current session, not all user's tokens
// IMPORTANT: This also adds tokens to blacklist for immediate invalidation
func (ts *TokenService) DeleteTokens(ctx context.Context, tokenHash string, userID string, sessionID string) error {
	// Step 1: Collect all tokens for blacklisting
	tokenHashes, maxTTL := ts.collectTokensForBlacklist(ctx, tokenHash, sessionID)

	// Step 2: Blacklist all tokens atomically (prevents race conditions)
	ts.blacklistTokens(ctx, tokenHashes, maxTTL)

	// Step 3: Delete tokens from storage
	tokensDeleted := ts.deleteAccessToken(ctx, tokenHash)
	if sessionID != "" {
		if ts.deleteSessionTokens(ctx, userID, sessionID) {
			tokensDeleted = true
		}
	}

	if !tokensDeleted {
		return utils.Error(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	return nil
}

// extractBearerToken extracts JWT token from Authorization header
func (ts *TokenService) extractBearerToken(authHeader string) (string, error) {
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
func (ts *TokenService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// collectTokensForBlacklist gathers all tokens and their TTLs for batch blacklisting
func (ts *TokenService) collectTokensForBlacklist(ctx context.Context, accessTokenHash, sessionID string) ([]string, time.Duration) {
	var tokenHashes []string
	var maxTTL time.Duration

	// Collect access token
	if accessToken, err := ts.authRepo.GetTokenByHash(ctx, accessTokenHash); err == nil && accessToken != nil {
		if ttl := time.Until(accessToken.ExpiresAt); ttl > 0 {
			tokenHashes = append(tokenHashes, accessTokenHash)
			maxTTL = ttl
		}
	}

	// Collect refresh token if session exists
	if sessionID != "" {
		if session, err := ts.authRepo.GetSessionByID(ctx, sessionID); err == nil && session != nil {
			if refreshToken, err := ts.authRepo.GetTokenByHash(ctx, session.RefreshTokenHash); err == nil && refreshToken != nil {
				if ttl := time.Until(refreshToken.ExpiresAt); ttl > 0 {
					tokenHashes = append(tokenHashes, session.RefreshTokenHash)
					if ttl > maxTTL {
						maxTTL = ttl
					}
				}
			}
		}
	}

	return tokenHashes, maxTTL
}

// blacklistTokens adds multiple tokens to blacklist in a single atomic operation
func (ts *TokenService) blacklistTokens(ctx context.Context, tokenHashes []string, ttl time.Duration) {
	if !ts.cacheService.IsEnabled() || len(tokenHashes) == 0 {
		return
	}

	if err := ts.cacheService.AddMultipleTokensToBlacklist(ctx, tokenHashes, ttl); err != nil {
		fmt.Printf("Warning: failed to add tokens to blacklist: %v\n", err)
	}
}

// deleteAccessToken removes access token from database and cache
func (ts *TokenService) deleteAccessToken(ctx context.Context, tokenHash string) bool {
	err := ts.authRepo.DeleteTokenByHash(ctx, tokenHash)
	if err != nil && err != gorm.ErrRecordNotFound {
		fmt.Printf("Warning: failed to delete access token from DB: %v\n", err)
	}

	if ts.cacheService.IsEnabled() {
		if err := ts.cacheService.DeleteToken(ctx, tokenHash); err != nil {
			fmt.Printf("Warning: failed to delete access token from cache: %v\n", err)
		}
	}

	return err == nil
}

// deleteSessionTokens removes session and refresh token from database and cache
func (ts *TokenService) deleteSessionTokens(ctx context.Context, userID, sessionID string) bool {
	// Get refresh token hash before deletion
	var refreshTokenHash string
	if session, err := ts.authRepo.GetSessionByID(ctx, sessionID); err == nil && session != nil {
		refreshTokenHash = session.RefreshTokenHash
	}

	// Delete from database
	err := ts.authRepo.DeleteTokensBySession(ctx, userID, sessionID)
	if err != nil && err != gorm.ErrRecordNotFound {
		fmt.Printf("Warning: failed to delete session tokens from DB: %v\n", err)
	}

	// Delete from cache
	if ts.cacheService.IsEnabled() {
		if err := ts.cacheService.DeleteSession(ctx, sessionID); err != nil {
			fmt.Printf("Warning: failed to delete session from cache: %v\n", err)
		}
		if refreshTokenHash != "" {
			if err := ts.cacheService.DeleteToken(ctx, refreshTokenHash); err != nil {
				fmt.Printf("Warning: failed to delete refresh token from cache: %v\n", err)
			}
		}
	}

	return err == nil
}
