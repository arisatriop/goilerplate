package auth

import (
	"context"
	"testing"
	"time"

	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testReuseGrace = 10 * time.Second

// refreshRepo records the revocations the refresh path performs, so a test can tell a plain
// 401 (no security event) apart from reuse detection (session revoked).
type refreshRepo struct {
	Repository
	session *UserSession
	revoked []revokeCall
}

type revokeCall struct {
	userID    string
	sessionID string
	reason    string
}

func (r *refreshRepo) GetSessionByID(context.Context, string) (*UserSession, error) {
	if r.session == nil {
		return nil, nil
	}
	copied := *r.session
	return &copied, nil
}

func (r *refreshRepo) RevokeSession(_ context.Context, userID, sessionID, reason string) error {
	r.revoked = append(r.revoked, revokeCall{userID, sessionID, reason})
	return nil
}

func newRefreshUseCase(repo *refreshRepo) *authUseCase {
	return &authUseCase{
		authRepo:          repo,
		sessionService:    NewSessionService(repo, newFakeSessionStore(), true),
		refreshReuseGrace: testReuseGrace,
	}
}

// rotatedSession is a session whose last rotation replaced previousJTI, `ago` in the past.
func rotatedSession(previousJTI string, ago time.Duration) *UserSession {
	rotatedAt := utils.Now().Add(-ago)
	return &UserSession{
		ID:                 "s1",
		UserID:             "u1",
		RefreshJTI:         "current-jti",
		PreviousRefreshJTI: previousJTI,
		RotatedAt:          &rotatedAt,
		IsActive:           true,
		ExpiresAt:          utils.Now().Add(time.Hour),
	}
}

// A client whose refresh response was lost, or a second tab racing the first, replays the
// token the last rotation just replaced. That must succeed and hand back the session's
// current token, so both callers converge instead of fighting.
func TestResolveFailedRotation_WithinGraceIsIdempotent(t *testing.T) {
	// Arrange
	repo := &refreshRepo{session: rotatedSession("replaced-jti", time.Second)}
	uc := newRefreshUseCase(repo)

	// Act
	session, err := uc.resolveFailedRotation(context.Background(), "u1", "s1", "replaced-jti")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "current-jti", session.RefreshJTI, "both callers end up on the session's current token")
	assert.Empty(t, repo.revoked, "a concurrent refresh is not an attack")
}

// The same token replayed once the grace window has passed is a stolen-token signal.
func TestResolveFailedRotation_AfterGraceRevokesThatSessionOnly(t *testing.T) {
	// Arrange
	repo := &refreshRepo{session: rotatedSession("replaced-jti", testReuseGrace+time.Second)}
	uc := newRefreshUseCase(repo)

	// Act
	session, err := uc.resolveFailedRotation(context.Background(), "u1", "s1", "replaced-jti")

	// Assert
	assert.Nil(t, session)
	assertUnauthorized(t, err)
	require.Len(t, repo.revoked, 1)
	assert.Equal(t, revokeCall{"u1", "s1", RevokedReasonReuseDetected}, repo.revoked[0],
		"only the session that presented the token is revoked")
}

// A token this session never issued cannot be explained by a race, whenever it arrives.
func TestResolveFailedRotation_UnknownTokenIsReuse(t *testing.T) {
	repo := &refreshRepo{session: rotatedSession("replaced-jti", time.Second)}
	uc := newRefreshUseCase(repo)

	_, err := uc.resolveFailedRotation(context.Background(), "u1", "s1", "never-issued-jti")

	assertUnauthorized(t, err)
	require.Len(t, repo.revoked, 1)
	assert.Equal(t, RevokedReasonReuseDetected, repo.revoked[0].reason)
}

// Refreshing after logout is an ordinary expired-credential case, not an attack: the session
// is already revoked, so there is nothing to revoke and no security event to raise.
func TestResolveFailedRotation_UnusableSessionIsPlain401(t *testing.T) {
	revoked := rotatedSession("replaced-jti", time.Second)
	revoked.IsActive = false

	expired := rotatedSession("replaced-jti", time.Second)
	expired.ExpiresAt = utils.Now().Add(-time.Minute)

	tests := []struct {
		name    string
		session *UserSession
	}{
		{"after logout", revoked},
		{"expired session", expired},
		{"missing session", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &refreshRepo{session: tt.session}
			uc := newRefreshUseCase(repo)

			_, err := uc.resolveFailedRotation(context.Background(), "u1", "s1", "replaced-jti")

			assertUnauthorized(t, err)
			assert.Empty(t, repo.revoked, "no security event for an already-dead session")
		})
	}
}

func TestWithinReuseGrace(t *testing.T) {
	rotatedAt := utils.Now().Add(-time.Second)
	neverRotated := &UserSession{RefreshJTI: "current-jti"}

	tests := []struct {
		name    string
		session *UserSession
		jti     string
		want    bool
	}{
		{"replaced token, just rotated", rotatedSession("replaced-jti", time.Second), "replaced-jti", true},
		{"replaced token, grace elapsed", rotatedSession("replaced-jti", testReuseGrace+time.Second), "replaced-jti", false},
		{"current token, not the replaced one", rotatedSession("replaced-jti", time.Second), "current-jti", false},
		{"session never rotated", neverRotated, "", false},
		{"previous jti set but no rotated_at", &UserSession{PreviousRefreshJTI: "replaced-jti"}, "replaced-jti", false},
		{"rotated exactly at the boundary", &UserSession{
			PreviousRefreshJTI: "replaced-jti",
			RotatedAt:          &rotatedAt,
		}, "replaced-jti", true},
	}

	uc := &authUseCase{refreshReuseGrace: testReuseGrace}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, uc.withinReuseGrace(tt.session, tt.jti))
		})
	}
}

func TestSessionExpiry_For(t *testing.T) {
	expiry := SessionExpiry{Default: 168 * time.Hour, RememberMe: 720 * time.Hour}

	assert.Equal(t, 168*time.Hour, expiry.For(false))
	assert.Equal(t, 720*time.Hour, expiry.For(true), "remember_me must use the longer lifetime")
}
