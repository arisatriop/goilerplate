package auth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"goilerplate/pkg/password"
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

// stateRepo records what DeactivateUser and ChangePassword did to the user and their sessions.
type stateRepo struct {
	Repository
	user          *User
	activeSet     *bool
	passwordHash  string
	revokedKeep   string
	revokedReason string
	revokeCalls   int
}

func (r *stateRepo) WithTx(context.Context) Repository { return r }

func (r *stateRepo) GetUserByID(context.Context, string) (*User, error) {
	if r.user == nil {
		return nil, nil
	}
	copied := *r.user
	return &copied, nil
}

func (r *stateRepo) SetUserActive(_ context.Context, _ string, active bool) error {
	r.activeSet = &active
	return nil
}

func (r *stateRepo) UpdateUserPassword(_ context.Context, _, hash string) error {
	r.passwordHash = hash
	return nil
}

func (r *stateRepo) RevokeOtherUserSessions(_ context.Context, _, keepSessionID, reason string) error {
	r.revokeCalls++
	r.revokedKeep = keepSessionID
	r.revokedReason = reason
	return nil
}

// inlineTx runs the function without a real transaction; the repository stub ignores WithTx.
type inlineTx struct{}

func (inlineTx) Do(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

func newStateUseCase(repo *stateRepo) *authUseCase {
	return &authUseCase{
		authRepo:       repo,
		sessionService: NewSessionService(repo, newFakeSessionStore(), true),
		txManager:      inlineTx{},
		passwordPolicy: password.NewPolicy(nil),
	}
}

// Deactivating an account must revoke its sessions too. Flipping the flag alone would leave the
// user with API access until their sessions expired, because the per-request check reads the
// session, not the account.
func TestDeactivateUser_RevokesEverySession(t *testing.T) {
	repo := &stateRepo{}
	uc := newStateUseCase(repo)

	require.NoError(t, uc.DeactivateUser(context.Background(), "u1"))

	require.NotNil(t, repo.activeSet)
	assert.False(t, *repo.activeSet)
	assert.Equal(t, 1, repo.revokeCalls)
	assert.Empty(t, repo.revokedKeep, "no session is kept when the account is disabled")
	assert.Equal(t, RevokedReasonAdmin, repo.revokedReason)
}

func TestChangePassword_RequiresTheCurrentPassword(t *testing.T) {
	hash, err := utils.HashPassword("the-current-password")
	require.NoError(t, err)
	repo := &stateRepo{user: &User{ID: "u1", PasswordHash: hash, IsActive: true}}
	uc := newStateUseCase(repo)

	err = uc.ChangePassword(context.Background(), "u1", "s1", "not-the-password", "a-new-strong-password")

	assertUnauthorized(t, err)
	assert.Zero(t, repo.revokeCalls, "nothing is revoked when the change is refused")
	assert.Empty(t, repo.passwordHash)
}

func TestChangePassword_KeepsTheCallersSession(t *testing.T) {
	hash, err := utils.HashPassword("the-current-password")
	require.NoError(t, err)
	repo := &stateRepo{user: &User{ID: "u1", PasswordHash: hash, IsActive: true}}
	uc := newStateUseCase(repo)

	err = uc.ChangePassword(context.Background(), "u1", "s1", "the-current-password", "a-new-strong-password")

	require.NoError(t, err)
	assert.NotEmpty(t, repo.passwordHash)
	assert.NotEqual(t, hash, repo.passwordHash, "the stored hash is replaced")
	assert.Equal(t, "s1", repo.revokedKeep, "the device making the change stays signed in")
	assert.Equal(t, RevokedReasonPasswordChange, repo.revokedReason)
}

// A weak new password is a validation error, and must not revoke anything on the way out.
func TestChangePassword_RejectsWeakNewPassword(t *testing.T) {
	hash, err := utils.HashPassword("the-current-password")
	require.NoError(t, err)
	repo := &stateRepo{user: &User{ID: "u1", PasswordHash: hash, IsActive: true}}
	uc := newStateUseCase(repo)

	err = uc.ChangePassword(context.Background(), "u1", "s1", "the-current-password", "short")

	var clientErr *utils.ClientError
	require.ErrorAs(t, err, &clientErr)
	assert.Equal(t, http.StatusBadRequest, clientErr.Code)
	assert.Zero(t, repo.revokeCalls)
}
