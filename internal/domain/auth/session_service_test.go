package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubRepository implements only the Repository methods these tests need;
// calling any other method panics on the nil embedded interface.
type stubRepository struct {
	Repository
	sessions        map[string]*UserSession
	sessionErr      error
	sessionLookups  int
	rolePermissions []string
	overrides       map[string]bool
	permissionReads int
}

func (r *stubRepository) GetSessionByID(_ context.Context, sessionID string) (*UserSession, error) {
	r.sessionLookups++
	if r.sessionErr != nil {
		return nil, r.sessionErr
	}
	session, ok := r.sessions[sessionID]
	if !ok {
		return nil, nil
	}
	copied := *session
	return &copied, nil
}

func (r *stubRepository) GetUserRolesByUserID(context.Context, string) ([]string, error) {
	r.permissionReads++
	return []string{"role-1"}, nil
}

func (r *stubRepository) GetRolePermissionsByRoleIDs(context.Context, []string) ([]string, error) {
	return r.rolePermissions, nil
}

func (r *stubRepository) GetUserPermissionOverrides(context.Context, string) (map[string]bool, error) {
	return r.overrides, nil
}

// fakeSessionStore is a map-backed SessionStore that can simulate cache failures.
type fakeSessionStore struct {
	sessions map[string]UserSession
	err      error
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{sessions: make(map[string]UserSession)}
}

func (s *fakeSessionStore) Get(_ context.Context, sessionID string) (*UserSession, bool, error) {
	if s.err != nil {
		return nil, false, s.err
	}
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, false, nil
	}
	return &session, true, nil
}

func (s *fakeSessionStore) Set(_ context.Context, session *UserSession) error {
	if s.err != nil {
		return s.err
	}
	s.sessions[session.ID] = *session
	return nil
}

func (s *fakeSessionStore) Delete(_ context.Context, sessionID string) error {
	if s.err != nil {
		return s.err
	}
	delete(s.sessions, sessionID)
	return nil
}

func (s *fakeSessionStore) DeleteByUser(_ context.Context, userID string) error {
	if s.err != nil {
		return s.err
	}
	for id, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, id)
		}
	}
	return nil
}

func activeSession(id, userID string) *UserSession {
	return &UserSession{ID: id, UserID: userID, IsActive: true, ExpiresAt: utils.Now().Add(time.Hour)}
}

func assertUnauthorized(t *testing.T, err error) {
	t.Helper()
	var clientErr *utils.ClientError
	require.ErrorAs(t, err, &clientErr)
	assert.Equal(t, http.StatusUnauthorized, clientErr.Code)
}

func TestSessionService_GetActive_CacheMissReadsRepositoryAndCaches(t *testing.T) {
	// Arrange
	repo := &stubRepository{sessions: map[string]*UserSession{"s1": activeSession("s1", "u1")}}
	store := newFakeSessionStore()
	service := NewSessionService(repo, store)
	ctx := context.Background()

	// Act
	first, err := service.GetActive(ctx, "s1")
	require.NoError(t, err)
	second, err := service.GetActive(ctx, "s1")
	require.NoError(t, err)

	// Assert
	assert.Equal(t, "u1", first.UserID)
	assert.Equal(t, "u1", second.UserID)
	assert.Equal(t, 1, repo.sessionLookups, "second call is served from cache")
	assert.Contains(t, store.sessions, "s1")
}

func TestSessionService_GetActive_RejectsInvalidSessions(t *testing.T) {
	inactive := activeSession("inactive", "u1")
	inactive.IsActive = false
	expired := activeSession("expired", "u1")
	expired.ExpiresAt = utils.Now().Add(-time.Minute)

	repo := &stubRepository{sessions: map[string]*UserSession{"inactive": inactive, "expired": expired}}
	service := NewSessionService(repo, newFakeSessionStore())

	for _, sessionID := range []string{"", "missing", "inactive", "expired"} {
		t.Run("session "+sessionID, func(t *testing.T) {
			_, err := service.GetActive(context.Background(), sessionID)
			assertUnauthorized(t, err)
		})
	}
}

func TestSessionService_GetActive_CachedRevocationIsHonored(t *testing.T) {
	// Arrange: the cache holds a revoked copy of a session
	revoked := activeSession("s1", "u1")
	revoked.IsActive = false
	store := newFakeSessionStore()
	store.sessions["s1"] = *revoked
	service := NewSessionService(&stubRepository{}, store)

	// Act
	_, err := service.GetActive(context.Background(), "s1")

	// Assert
	assertUnauthorized(t, err)
}

func TestSessionService_GetActive_CacheFailureFallsBackToRepository(t *testing.T) {
	// Arrange
	repo := &stubRepository{sessions: map[string]*UserSession{"s1": activeSession("s1", "u1")}}
	store := newFakeSessionStore()
	store.err = errors.New("redis down")
	service := NewSessionService(repo, store)

	// Act
	session, err := service.GetActive(context.Background(), "s1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "s1", session.ID)
	assert.Equal(t, 1, repo.sessionLookups)
}

func TestSessionService_GetActive_RepositoryErrorIsReturned(t *testing.T) {
	repo := &stubRepository{sessionErr: errors.New("db down")}
	service := NewSessionService(repo, newFakeSessionStore())

	_, err := service.GetActive(context.Background(), "s1")

	require.Error(t, err)
	var clientErr *utils.ClientError
	assert.False(t, errors.As(err, &clientErr), "infrastructure errors are not reported as 401")
}

func TestSessionService_Evict(t *testing.T) {
	// Arrange
	store := newFakeSessionStore()
	store.sessions["a1"] = *activeSession("a1", "alice")
	store.sessions["a2"] = *activeSession("a2", "alice")
	store.sessions["b1"] = *activeSession("b1", "bob")
	service := NewSessionService(&stubRepository{}, store)
	ctx := context.Background()

	// Act
	service.Evict(ctx, "a1")
	service.EvictUser(ctx, "alice")

	// Assert
	assert.NotContains(t, store.sessions, "a1")
	assert.NotContains(t, store.sessions, "a2")
	assert.Contains(t, store.sessions, "b1")
}

func TestSessionService_Evict_CacheFailureDoesNotPanic(t *testing.T) {
	store := newFakeSessionStore()
	store.err = errors.New("redis down")
	service := NewSessionService(&stubRepository{}, store)

	assert.NotPanics(t, func() {
		service.Evict(context.Background(), "s1")
		service.EvictUser(context.Background(), "u1")
	})
}
