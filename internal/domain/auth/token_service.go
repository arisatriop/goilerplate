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

// DeleteTokens deletes the caller's access token and revokes its session, which is what
// invalidates the session's refresh token: every request checks that the session is still
// active. This only affects the current session, not the user's other devices.
// Evicting the session from cache is the caller's responsibility.
func (ts *TokenService) DeleteTokens(ctx context.Context, tokenHash string, userID string, sessionID string) error {
	revoked := ts.deleteAccessToken(ctx, tokenHash)
	if sessionID != "" {
		if ts.revokeSession(ctx, userID, sessionID) {
			revoked = true
		}
	}

	if !revoked {
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

// revokeSession deactivates the session, which invalidates its refresh token
func (ts *TokenService) revokeSession(ctx context.Context, userID, sessionID string) bool {
	err := ts.authRepo.RevokeSession(ctx, userID, sessionID, RevokedReasonLogout)
	if err != nil && !errors.Is(err, ErrNotFound) {
		logger.Error(ctx, fmt.Errorf("revoking session: %w", err))
	}

	return err == nil
}
