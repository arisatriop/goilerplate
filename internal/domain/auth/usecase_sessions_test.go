package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sessionsRepo records what the session-management flows asked the repository to do.
type sessionsRepo struct {
	Repository
	listed      []UserSession
	listErr     error
	revokeErr   error
	revokedUser string
	revokedID   string
	revokedWhy  string
	keptID      string
	revokeAll   int
}

func (r *sessionsRepo) ListActiveSessions(context.Context, string) ([]UserSession, error) {
	return r.listed, r.listErr
}

func (r *sessionsRepo) RevokeSession(_ context.Context, userID, sessionID, reason string) error {
	r.revokedUser, r.revokedID, r.revokedWhy = userID, sessionID, reason
	return r.revokeErr
}

func (r *sessionsRepo) RevokeOtherUserSessions(_ context.Context, _, keepSessionID, reason string) error {
	r.revokeAll++
	r.keptID, r.revokedWhy = keepSessionID, reason
	return nil
}

func newSessionsUseCase(repo *sessionsRepo, store *fakeSessionStore) *authUseCase {
	return &authUseCase{authRepo: repo, sessionService: NewSessionService(repo, store, true)}
}

func TestAuthUseCase_ListSessions(t *testing.T) {
	// Arrange
	repo := &sessionsRepo{listed: []UserSession{{ID: "s1"}, {ID: "s2"}}}

	// Act
	got, err := newSessionsUseCase(repo, newFakeSessionStore()).ListSessions(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestAuthUseCase_ListSessions_RepositoryErrorIsWrapped(t *testing.T) {
	repo := &sessionsRepo{listErr: errors.New("connection reset")}

	_, err := newSessionsUseCase(repo, newFakeSessionStore()).ListSessions(context.Background(), "u1")

	require.Error(t, err)
	var clientErr *utils.ClientError
	assert.False(t, errors.As(err, &clientErr), "a database failure is a 500, not a client error")
}

// A revoked session must also leave the cache, or it keeps authenticating until the cache TTL.
func TestAuthUseCase_RevokeSession_RevokesAndEvicts(t *testing.T) {
	// Arrange
	repo := &sessionsRepo{}
	store := newFakeSessionStore()
	store.sessions["s2"] = UserSession{ID: "s2", UserID: "u1", IsActive: true}

	// Act
	err := newSessionsUseCase(repo, store).RevokeSession(context.Background(), "u1", "s2")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "u1", repo.revokedUser, "ownership is enforced by passing the caller's user ID")
	assert.Equal(t, "s2", repo.revokedID)
	assert.Equal(t, RevokedReasonLogout, repo.revokedWhy)
	assert.NotContains(t, store.sessions, "s2", "the revoked session is still cached")
}

// Not found covers "not yours" and "already revoked" too, because the repository's UPDATE
// matches on all three at once.
func TestAuthUseCase_RevokeSession_NotFoundIs404(t *testing.T) {
	repo := &sessionsRepo{revokeErr: ErrNotFound}

	err := newSessionsUseCase(repo, newFakeSessionStore()).RevokeSession(context.Background(), "u1", "s2")

	var clientErr *utils.ClientError
	require.ErrorAs(t, err, &clientErr)
	assert.Equal(t, http.StatusNotFound, clientErr.Code)
	assert.Equal(t, MsgSessionNotFound, clientErr.Error())
}

func TestAuthUseCase_RevokeSession_RepositoryErrorIsNotAClientError(t *testing.T) {
	repo := &sessionsRepo{revokeErr: errors.New("connection reset")}

	err := newSessionsUseCase(repo, newFakeSessionStore()).RevokeSession(context.Background(), "u1", "s2")

	require.Error(t, err)
	var clientErr *utils.ClientError
	assert.False(t, errors.As(err, &clientErr))
}

func TestAuthUseCase_LogoutAll_KeepsTheGivenSession(t *testing.T) {
	repo := &sessionsRepo{}

	require.NoError(t, newSessionsUseCase(repo, newFakeSessionStore()).LogoutAll(context.Background(), "u1", "s1"))

	assert.Equal(t, 1, repo.revokeAll)
	assert.Equal(t, "s1", repo.keptID)
	assert.Equal(t, RevokedReasonLogoutAll, repo.revokedWhy)
}

func TestAuthUseCase_LogoutAll_EmptyKeepRevokesEverySession(t *testing.T) {
	repo := &sessionsRepo{}

	require.NoError(t, newSessionsUseCase(repo, newFakeSessionStore()).LogoutAll(context.Background(), "u1", ""))

	assert.Equal(t, 1, repo.revokeAll)
	assert.Empty(t, repo.keptID)
}
