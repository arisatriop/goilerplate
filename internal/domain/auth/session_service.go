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
	repo  Repository
	store SessionStore
}

// NewSessionService creates a SessionService.
func NewSessionService(repo Repository, store SessionStore) *SessionService {
	return &SessionService{repo: repo, store: store}
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
