package auth

import (
	"context"
	"errors"
	"fmt"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
	"net/http"
)

// TokenService handles token operations
type TokenService struct {
	authRepo Repository
}

// NewTokenService creates a new token service
func NewTokenService(authRepo Repository) *TokenService {
	return &TokenService{
		authRepo: authRepo,
	}
}

// DeleteTokens deletes the access token and the session's refresh token, and deactivates
// the session. This only affects the current session, not all of the user's tokens.
// Evicting the session from cache is the caller's responsibility.
func (ts *TokenService) DeleteTokens(ctx context.Context, tokenHash string, userID string, sessionID string) error {
	tokensDeleted := ts.deleteAccessToken(ctx, tokenHash)
	if sessionID != "" {
		if ts.deleteSessionTokens(ctx, userID, sessionID) {
			tokensDeleted = true
		}
	}

	if !tokensDeleted {
		return utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	return nil
}

// deleteAccessToken removes the access token from the database
func (ts *TokenService) deleteAccessToken(ctx context.Context, tokenHash string) bool {
	err := ts.authRepo.DeleteTokenByHash(ctx, tokenHash)
	if err != nil && !errors.Is(err, ErrNotFound) {
		logger.Error(ctx, fmt.Errorf("deleting access token: %w", err))
	}

	return err == nil
}

// deleteSessionTokens removes the session's refresh token and deactivates the session
func (ts *TokenService) deleteSessionTokens(ctx context.Context, userID, sessionID string) bool {
	err := ts.authRepo.DeleteTokensBySession(ctx, userID, sessionID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		logger.Error(ctx, fmt.Errorf("deleting session tokens: %w", err))
	}

	return err == nil
}
