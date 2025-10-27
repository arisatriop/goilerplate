package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"goilerplate/pkg/jwt"
	"goilerplate/pkg/logger"
	"time"
)

// TokenStorage handles token storage operations
type TokenStorage struct {
	authRepo     Repository
	cacheService *CacheService
}

// NewTokenStorage creates a new token storage
func NewTokenStorage(authRepo Repository, cacheService *CacheService) *TokenStorage {
	return &TokenStorage{
		authRepo:     authRepo,
		cacheService: cacheService,
	}
}

// StoreTokenPair stores both access and refresh tokens
func (ts *TokenStorage) StoreTokenPair(ctx context.Context, userID, sessionID string, tokenPair *jwt.TokenPair, deviceInfo *DeviceInfo) error {
	// Store access token in database
	accessToken := &UserToken{
		UserID:    userID,
		TokenHash: ts.hashToken(tokenPair.AccessToken),
		TokenType: TokenTypeAccess,
		ExpiresAt: tokenPair.AccessTokenExpiresAt,
		IPAddress: deviceInfo.IPAddress,
		UserAgent: deviceInfo.UserAgent,
	}

	createdAccessToken, err := ts.authRepo.CreateToken(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	// Cache access token to Redis if enabled
	if ts.cacheService.IsEnabled() {
		if err := ts.cacheService.CacheToken(ctx, createdAccessToken); err != nil {
			return fmt.Errorf("failed to cache access token: %w", err)
		}
	}

	// Store refresh token in database
	refreshToken := &UserToken{
		UserID:    userID,
		TokenHash: ts.hashToken(tokenPair.RefreshToken),
		TokenType: TokenTypeRefresh,
		ExpiresAt: tokenPair.RefreshTokenExpiresAt,
		IPAddress: deviceInfo.IPAddress,
		UserAgent: deviceInfo.UserAgent,
	}

	createdRefreshToken, err := ts.authRepo.CreateToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Cache refresh token to Redis if enabled
	if ts.cacheService.IsEnabled() {
		if err := ts.cacheService.CacheToken(ctx, createdRefreshToken); err != nil {
			return fmt.Errorf("failed to cache refresh token: %w", err)
		}
	}

	return nil
}

// StoreAccessToken stores only an access token
func (ts *TokenStorage) StoreAccessToken(ctx context.Context, userID string, accessTokenString string, expiresAt time.Time, deviceInfo *DeviceInfo) error {
	accessToken := &UserToken{
		UserID:    userID,
		TokenHash: ts.hashToken(accessTokenString),
		TokenType: TokenTypeAccess,
		ExpiresAt: expiresAt,
		IPAddress: deviceInfo.IPAddress,
		UserAgent: deviceInfo.UserAgent,
	}

	createdAccessToken, err := ts.authRepo.CreateToken(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	// Cache access token to Redis if enabled
	if ts.cacheService.IsEnabled() {
		if err := ts.cacheService.CacheToken(ctx, createdAccessToken); err != nil {
			return fmt.Errorf("failed to cache acess token: %w", err)
		}
	}

	return nil
}

// MarkTokenAsUsedAsync marks token as used in background
func (ts *TokenStorage) MarkTokenAsUsedAsync(ctx context.Context, tokenID string) {
	go func(tokenID string, userCtx context.Context) {
		err := ts.authRepo.MarkTokenAsUsed(userCtx, tokenID)
		if err != nil {
			logger.Error(userCtx, err)
		}
	}(tokenID, ctx)
}

// hashToken creates a SHA256 hash of the token for secure storage
func (ts *TokenStorage) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
