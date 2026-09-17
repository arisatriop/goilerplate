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
	authRepo Repository
}

// NewTokenStorage creates a new token storage
func NewTokenStorage(authRepo Repository) *TokenStorage {
	return &TokenStorage{
		authRepo: authRepo,
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

	_, err := ts.authRepo.CreateToken(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
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

	_, err = ts.authRepo.CreateToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
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

	_, err := ts.authRepo.CreateToken(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	return nil
}

// MarkTokenAsUsedAsync marks token as used in background
func (ts *TokenStorage) MarkTokenAsUsedAsync(ctx context.Context, tokenID string) {
	bgCtx := context.WithoutCancel(ctx)
	go func() {
		err := ts.authRepo.MarkTokenAsUsed(bgCtx, tokenID)
		if err != nil {
			logger.Error(bgCtx, err)
		}
	}()
}

// hashToken creates a SHA256 hash of the token for secure storage
func (ts *TokenStorage) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
