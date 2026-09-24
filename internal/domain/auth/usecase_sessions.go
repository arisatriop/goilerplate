package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
)

// MsgSessionNotFound is returned when a session to revoke does not exist, is not the caller's,
// or is already revoked. The three are deliberately indistinguishable: telling "not yours" apart
// from "does not exist" would confirm another user's session IDs.
const MsgSessionNotFound = "Session not found"

// ListSessions returns the user's active sessions, most recently used first.
func (uc *authUseCase) ListSessions(ctx context.Context, userID string) ([]UserSession, error) {
	sessions, err := uc.authRepo.ListActiveSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	return sessions, nil
}

// RevokeSession signs one of the user's own sessions out, typically another device. Ownership is
// enforced by the repository's conditional UPDATE, which matches on user_id as well as the
// session ID, so there is no read-then-write window in which the check could be raced.
func (uc *authUseCase) RevokeSession(ctx context.Context, userID, sessionID string) error {
	if err := uc.authRepo.RevokeSession(ctx, userID, sessionID, RevokedReasonLogout); err != nil {
		if errors.Is(err, ErrNotFound) {
			return utils.ClientErr(http.StatusNotFound, MsgSessionNotFound)
		}
		return fmt.Errorf("revoking session: %w", err)
	}

	uc.sessionService.Evict(ctx, sessionID)

	logger.Security(ctx, logger.SecurityEvent{
		Action:    logger.ActionSessionRevoked,
		Outcome:   logger.OutcomeSuccess,
		UserID:    userID,
		SessionID: sessionID,
		Reason:    RevokedReasonLogout,
	})

	return nil
}
