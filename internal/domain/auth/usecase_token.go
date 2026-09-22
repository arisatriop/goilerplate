// The refresh-token lifecycle: rotating a refresh token, detecting reuse of a spent one, and
// minting the pair a session carries.

package auth

import (
	"context"
	"errors"
	"fmt"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/jwt"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
	"net/http"
	"time"
)

// RefreshToken generates new access token using refresh token
// Note: Token validation is handled by AuthenticateRefreshToken middleware
func (uc *authUseCase) RefreshToken(ctx context.Context, userID, sessionID, refreshJTI string, deviceInfo *DeviceInfo) (*LoginResult, error) {
	// Validate user is still allowed to refresh (not locked/disabled)
	user, err := uc.userValidator.ValidateUserForRefresh(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user for refresh: %w", err)
	}

	session, err := uc.rotateRefreshToken(ctx, userID, sessionID, refreshJTI)
	if err != nil {
		return nil, err
	}

	tokenPair, err := uc.issueTokensForSession(user, session, deviceInfo)
	if err != nil {
		return nil, err
	}

	// Refresh with fresh permissions; the cache refills on the next check
	if err := uc.permissionService.InvalidateUserPermissions(ctx, user.ID); err != nil {
		logger.Error(ctx, err)
	}

	menu, permissions, err := uc.buildMenuAndPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:       user,
		Menu:       menu,
		Permission: permissions,
		Tokens:     tokenPair,
		Session:    session,
	}, nil
}

// rotateRefreshToken claims the presented refresh token and returns the session the new tokens
// belong to. The returned session always carries the refresh_jti the caller must sign.
//
// The session was already checked by the refresh middleware, so a missing, revoked, or expired
// session has become a plain 401 before we get here and is never mistaken for token reuse.
func (uc *authUseCase) rotateRefreshToken(ctx context.Context, userID, sessionID, refreshJTI string) (*UserSession, error) {
	newJTI := utils.GenerateUUID()

	err := uc.authRepo.RotateRefreshJTI(ctx, sessionID, refreshJTI, newJTI)
	if err == nil {
		session, err := uc.authRepo.GetSessionByID(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("reading rotated session: %w", err)
		}
		if session == nil {
			return nil, utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
		}
		return session, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("rotating refresh token: %w", err)
	}

	return uc.resolveFailedRotation(ctx, userID, sessionID, refreshJTI)
}

// resolveFailedRotation decides what a rotation that matched no row actually means. The copy
// of the session loaded earlier may predate a concurrent rotation, so it is re-read from the
// repository (never the cache) after the failed UPDATE, which PostgreSQL made wait for the
// concurrent commit.
func (uc *authUseCase) resolveFailedRotation(ctx context.Context, userID, sessionID, refreshJTI string) (*UserSession, error) {
	session, err := uc.authRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("re-reading session after failed rotation: %w", err)
	}
	if session == nil || !session.IsValidSession() {
		return nil, utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	// The token we were handed is the one the previous rotation just replaced, and that
	// rotation is recent: this is a concurrent tab or a lost response, not an attack. Hand back
	// the session's current refresh token rather than rotating again, so both callers converge
	// on the same token.
	if uc.withinReuseGrace(session, refreshJTI) {
		return session, nil
	}

	// Anything else means a refresh token was presented that this session no longer accepts and
	// cannot explain. Revoke this login only; the user's other devices are untouched.
	if err := uc.authRepo.RevokeSession(ctx, userID, sessionID, RevokedReasonReuseDetected); err != nil && !errors.Is(err, ErrNotFound) {
		logger.Error(ctx, fmt.Errorf("revoking session after refresh token reuse: %w", err))
	}
	uc.sessionService.Evict(ctx, sessionID)

	// The jti that was replayed is deliberately not recorded: it is part of a bearer token, and
	// the audit trail outlives the token by far. The session it belongs to is what an
	// investigation needs anyway.
	logger.Security(ctx, logger.SecurityEvent{
		Action:    logger.ActionTokenReuseDetected,
		Outcome:   logger.OutcomeFailure,
		UserID:    userID,
		SessionID: sessionID,
		Reason:    RevokedReasonReuseDetected,
	})

	return nil, utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
}

// withinReuseGrace reports whether refreshJTI is the token the last rotation replaced, and
// that rotation happened recently enough to treat replaying it as harmless.
func (uc *authUseCase) withinReuseGrace(session *UserSession, refreshJTI string) bool {
	if session.PreviousRefreshJTI == "" || session.PreviousRefreshJTI != refreshJTI {
		return false
	}
	if session.RotatedAt == nil {
		return false
	}
	return utils.Now().Sub(*session.RotatedAt) <= uc.refreshReuseGrace
}

// issueTokensForSession mints an access token and a refresh token carrying the session's
// current refresh_jti. The refresh token expires with the session, so refreshing never
// extends the login.
func (uc *authUseCase) issueTokensForSession(user *User, session *UserSession, deviceInfo *DeviceInfo) (*jwt.TokenPair, error) {
	accessToken, accessExpiresAt, err := uc.jwtService.GenerateAccessToken(
		user.ID, user.Name, user.Email, session.ID, deviceInfo.DeviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := uc.jwtService.SignRefreshToken(
		user.ID, session.ID, deviceInfo.DeviceID, session.RefreshJTI, session.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("signing refresh token: %w", err)
	}

	return uc.buildTokenPair(accessToken, accessExpiresAt, refreshToken, session.ExpiresAt), nil
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
