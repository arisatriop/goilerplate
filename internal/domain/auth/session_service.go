package auth

import (
	"context"
	"fmt"
	"net/http"

	"goilerplate/pkg/constants"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
)

// SessionService checks sessions through the SessionStore, falling back to the repository.
// Cache failures are logged and never change the outcome: the repository is the source of truth.
type SessionService struct {
	repo Repository
	// checkEveryRequest reflects auth.revocation: true for strict, false for refresh_only.
	// Wiring resolves the mode once so no request path has to branch on configuration.
	checkEveryRequest bool
	store             SessionStore
}

// NewSessionService creates a SessionService. checkEveryRequest comes from auth.revocation:
// strict checks the session on every authenticated request, refresh_only only on refresh.
func NewSessionService(repo Repository, store SessionStore, checkEveryRequest bool) *SessionService {
	return &SessionService{repo: repo, store: store, checkEveryRequest: checkEveryRequest}
}

// GetActive returns the session when it exists, is active, and has not expired.
// Otherwise it returns a 401 client error.
func (s *SessionService) GetActive(ctx context.Context, sessionID string) (*UserSession, error) {
	if sessionID == "" {
		return nil, utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	session, found, err := s.store.Get(ctx, sessionID)
	if err != nil {
		logger.Error(ctx, fmt.Errorf("reading session cache: %w", err))
		found = false
	}

	if !found {
		session, err = s.repo.GetSessionByID(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("getting session: %w", err)
		}
		if session == nil {
			return nil, utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
		}
		if err := s.store.Set(ctx, session); err != nil {
			logger.Error(ctx, fmt.Errorf("caching session: %w", err))
		}
	}

	if !session.IsValidSession() {
		return nil, utils.ClientErr(http.StatusUnauthorized, constants.MsgUnauthorized)
	}

	return session, nil
}

// EnsureActiveForRequest is the per-request check. In strict mode it behaves like GetActive;
// in refresh_only it does nothing, so revocation takes effect when the access token expires
// rather than immediately. Refresh and logout always use GetActive regardless of the mode.
func (s *SessionService) EnsureActiveForRequest(ctx context.Context, sessionID string) error {
	if !s.checkEveryRequest {
		return nil
	}

	_, err := s.GetActive(ctx, sessionID)
	return err
}

// Evict removes one session from the cache after it was revoked in the repository.
// A failed eviction is logged: the revoked session may be served from cache until the
// cache TTL (auth.session_cache_ttl) elapses.
func (s *SessionService) Evict(ctx context.Context, sessionID string) {
	if err := s.store.Delete(ctx, sessionID); err != nil {
		logger.Error(ctx, fmt.Errorf("evicting session from cache: %w", err))
	}
}

// EvictUser removes every cached session of a user after they were revoked in the repository.
// Failures are handled like Evict.
func (s *SessionService) EvictUser(ctx context.Context, userID string) {
	if err := s.store.DeleteByUser(ctx, userID); err != nil {
		logger.Error(ctx, fmt.Errorf("evicting user sessions from cache: %w", err))
	}
}
