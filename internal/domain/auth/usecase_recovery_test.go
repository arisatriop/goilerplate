package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"goilerplate/pkg/hash"
	"goilerplate/pkg/password"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testResetTTL      = 30 * time.Minute
	testResetCooldown = time.Minute
)

// recoveryRepo holds one user and the reset tokens issued to them, with the consumption and
// expiry semantics the SQL implements: a token is usable while unused and unexpired.
type recoveryRepo struct {
	Repository
	user         *User
	tokens       []*OneTimeToken
	lookupErr    error
	passwordHash string
	revokedKeep  *string
	revokeReason string
}

func (r *recoveryRepo) WithTx(context.Context) Repository { return r }

func (r *recoveryRepo) GetUserByEmail(_ context.Context, email string) (*User, error) {
	if r.lookupErr != nil {
		return nil, r.lookupErr
	}
	if r.user == nil || r.user.Email != email {
		return nil, nil
	}
	copied := *r.user
	return &copied, nil
}

func (r *recoveryRepo) GetUserByID(_ context.Context, userID string) (*User, error) {
	if r.user == nil || r.user.ID != userID {
		return nil, nil
	}
	copied := *r.user
	return &copied, nil
}

func usable(token *OneTimeToken) bool { return !token.IsUsed() && !token.IsExpired() }

func (r *recoveryRepo) GetLatestActiveOneTimeToken(_ context.Context, userID, tokenType string) (*OneTimeToken, error) {
	for i := len(r.tokens) - 1; i >= 0; i-- {
		token := r.tokens[i]
		if token.UserID == userID && token.TokenType == tokenType && usable(token) {
			copied := *token
			return &copied, nil
		}
	}
	return nil, nil
}

func (r *recoveryRepo) ExpireOneTimeTokens(_ context.Context, userID, tokenType string) error {
	for _, token := range r.tokens {
		if token.UserID == userID && token.TokenType == tokenType && usable(token) {
			token.ExpiresAt = utils.Now()
		}
	}
	return nil
}

func (r *recoveryRepo) CreateOneTimeToken(_ context.Context, token *OneTimeToken) error {
	copied := *token
	copied.CreatedAt = utils.Now()
	r.tokens = append(r.tokens, &copied)
	return nil
}

func (r *recoveryRepo) ConsumeOneTimeToken(_ context.Context, tokenHash, tokenType string) (*OneTimeToken, error) {
	for _, token := range r.tokens {
		if token.TokenHash == tokenHash && token.TokenType == tokenType && usable(token) {
			now := utils.Now()
			token.UsedAt = &now
			copied := *token
			return &copied, nil
		}
	}
	return nil, ErrNotFound
}

func (r *recoveryRepo) UpdateUserPassword(_ context.Context, _, hash string) error {
	r.passwordHash = hash
	return nil
}

func (r *recoveryRepo) RevokeOtherUserSessions(_ context.Context, _, keepSessionID, reason string) error {
	r.revokedKeep = &keepSessionID
	r.revokeReason = reason
	return nil
}

// recordingNotifier keeps every notice, so a test can read back the plaintext token the way the
// user would from their inbox.
type recordingNotifier struct {
	notices []PasswordResetNotice
	err     error
}

func (n *recordingNotifier) SendPasswordReset(_ context.Context, notice PasswordResetNotice) error {
	if n.err != nil {
		return n.err
	}
	n.notices = append(n.notices, notice)
	return nil
}

func newRecoveryUseCase(repo *recoveryRepo, notifier Notifier) (*authUseCase, *fakeSessionStore) {
	store := newFakeSessionStore()
	return &authUseCase{
		authRepo:       repo,
		sessionService: NewSessionService(repo, store, true),
		txManager:      inlineTx{},
		passwordPolicy: password.NewPolicy(nil),
		recovery: Recovery{
			Notifier:       notifier,
			ResetTTL:       testResetTTL,
			ResendCooldown: testResetCooldown,
		},
	}, store
}

func activeUser() *User {
	return &User{ID: "u1", Name: "Ana", Email: "ana@example.org", IsActive: true}
}

// age moves every issued token's timestamps back by d, as if it had been issued that long ago.
func (r *recoveryRepo) age(d time.Duration) {
	for _, token := range r.tokens {
		token.CreatedAt = token.CreatedAt.Add(-d)
		token.ExpiresAt = token.ExpiresAt.Add(-d)
	}
}

func TestForgotPassword_EmailsAHashedSingleUseToken(t *testing.T) {
	t.Parallel()

	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc, _ := newRecoveryUseCase(repo, notifier)

	err := uc.ForgotPassword(t.Context(), "ana@example.org", RequestOrigin{IPAddress: "203.0.113.7", UserAgent: "curl/8"})

	require.NoError(t, err)
	require.Len(t, notifier.notices, 1)
	notice := notifier.notices[0]
	assert.Equal(t, "ana@example.org", notice.Email)
	assert.Equal(t, "Ana", notice.Name)
	assert.GreaterOrEqual(t, len(notice.Token), 43, "256 bits of entropy, base64url-encoded")

	require.Len(t, repo.tokens, 1)
	stored := repo.tokens[0]
	assert.Equal(t, hash.Token(notice.Token), stored.TokenHash, "only the hash is stored")
	assert.NotEqual(t, notice.Token, stored.TokenHash)
	assert.Equal(t, OneTimeTokenPasswordReset, stored.TokenType)
	assert.Equal(t, notice.ExpiresAt, stored.ExpiresAt)
	assert.WithinDuration(t, utils.Now().Add(testResetTTL), stored.ExpiresAt, time.Second)
	assert.Equal(t, "203.0.113.7", stored.IPAddress)
	assert.Equal(t, "curl/8", stored.UserAgent)
}

// The response must not say whether an address has an account. Each of these answers exactly
// as a registered address does, and sends nothing.
func TestForgotPassword_AnswersTheSameWhenNothingIsSent(t *testing.T) {
	t.Parallel()

	disabled := activeUser()
	disabled.IsActive = false

	tests := []struct {
		name  string
		user  *User
		email string
	}{
		{"unknown address", activeUser(), "nobody@example.org"},
		{"disabled account", disabled, "ana@example.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recoveryRepo{user: tt.user}
			notifier := &recordingNotifier{}
			uc, _ := newRecoveryUseCase(repo, notifier)

			err := uc.ForgotPassword(t.Context(), tt.email, RequestOrigin{})

			assert.NoError(t, err)
			assert.Empty(t, notifier.notices)
			assert.Empty(t, repo.tokens)
		})
	}
}

// A failure to queue the email is the operator's problem, not the client's: a 500 only for
// registered addresses would undo everything the endpoint does to hide them.
func TestForgotPassword_DeliveryFailureStillAnswersSuccess(t *testing.T) {
	t.Parallel()

	repo := &recoveryRepo{user: activeUser()}
	uc, _ := newRecoveryUseCase(repo, &recordingNotifier{err: errors.New("queue full")})

	err := uc.ForgotPassword(t.Context(), "ana@example.org", RequestOrigin{})

	assert.NoError(t, err)
}

// A database failure is not specific to registered addresses, so it is reported as one.
func TestForgotPassword_LookupFailureIsAnError(t *testing.T) {
	t.Parallel()

	repo := &recoveryRepo{user: activeUser(), lookupErr: errors.New("connection refused")}
	uc, _ := newRecoveryUseCase(repo, &recordingNotifier{})

	err := uc.ForgotPassword(t.Context(), "ana@example.org", RequestOrigin{})

	assert.Error(t, err)
}

// The cooldown stops the endpoint being used to flood someone's inbox; once it has passed, a
// new link replaces the old one, which stops working.
func TestForgotPassword_CooldownThenSupersede(t *testing.T) {
	t.Parallel()

	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc, _ := newRecoveryUseCase(repo, notifier)

	require.NoError(t, uc.ForgotPassword(t.Context(), "ana@example.org", RequestOrigin{}))

	repo.age(testResetCooldown - time.Second)
	require.NoError(t, uc.ForgotPassword(t.Context(), "ana@example.org", RequestOrigin{}))
	assert.Len(t, notifier.notices, 1, "a second request inside the cooldown sends nothing")

	repo.age(2 * time.Second)
	require.NoError(t, uc.ForgotPassword(t.Context(), "ana@example.org", RequestOrigin{}))
	require.Len(t, notifier.notices, 2)

	first, second := notifier.notices[0].Token, notifier.notices[1].Token
	assert.ErrorIs(t, uc.ResetPassword(t.Context(), first, "a-new-strong-password"), ErrInvalidResetToken,
		"the older link stops working once a new one is issued")
	assert.NoError(t, uc.ResetPassword(t.Context(), second, "a-new-strong-password"))
}

func TestForgotPassword_UnavailableWithoutANotifier(t *testing.T) {
	t.Parallel()

	uc, _ := newRecoveryUseCase(&recoveryRepo{user: activeUser()}, nil)

	assert.ErrorIs(t, uc.ForgotPassword(t.Context(), "ana@example.org", RequestOrigin{}), errRecoveryUnavailable)
	assert.ErrorIs(t, uc.ResetPassword(t.Context(), "token", "a-new-strong-password"), errRecoveryUnavailable)
}

// issueToken runs ForgotPassword and returns the token the user would find in their inbox.
func issueToken(t *testing.T, uc *authUseCase, notifier *recordingNotifier) string {
	t.Helper()
	require.NoError(t, uc.ForgotPassword(t.Context(), "ana@example.org", RequestOrigin{}))
	require.NotEmpty(t, notifier.notices)
	return notifier.notices[len(notifier.notices)-1].Token
}

func TestResetPassword_ReplacesThePasswordAndRevokesEverySession(t *testing.T) {
	t.Parallel()

	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc, store := newRecoveryUseCase(repo, notifier)
	token := issueToken(t, uc, notifier)
	store.sessions["s1"] = *activeSession("s1", "u1")

	err := uc.ResetPassword(t.Context(), token, "a-new-strong-password")

	require.NoError(t, err)
	require.NoError(t, utils.CheckPassword("a-new-strong-password", repo.passwordHash))
	require.NotNil(t, repo.revokedKeep)
	assert.Empty(t, *repo.revokedKeep, "no session survives a reset, not even one on the device asking")
	assert.Equal(t, RevokedReasonPasswordReset, repo.revokeReason)
	assert.Empty(t, store.sessions, "the user's cached sessions are evicted, so revocation is immediate")

	assert.ErrorIs(t, uc.ResetPassword(t.Context(), token, "another-strong-password"), ErrInvalidResetToken,
		"a link works once")
}

func TestResetPassword_RefusesATokenItCannotRedeem(t *testing.T) {
	t.Parallel()

	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc, _ := newRecoveryUseCase(repo, notifier)
	token := issueToken(t, uc, notifier)

	tests := []struct {
		name  string
		token string
		setup func()
	}{
		{"empty token", "", func() {}},
		{"token never issued", "not-a-real-token", func() {}},
		{"issued token after expiry", token, func() { repo.age(testResetTTL) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := uc.ResetPassword(t.Context(), tt.token, "a-new-strong-password")

			assert.ErrorIs(t, err, ErrInvalidResetToken)
			assert.Empty(t, repo.passwordHash)
			assert.Nil(t, repo.revokedKeep)
		})
	}
}

// A password the policy refuses is the user's typo to fix, not a reason to make them request
// another email: the link must still work afterwards.
func TestResetPassword_WeakPasswordDoesNotUseUpTheLink(t *testing.T) {
	t.Parallel()

	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc, _ := newRecoveryUseCase(repo, notifier)
	token := issueToken(t, uc, notifier)

	err := uc.ResetPassword(t.Context(), token, "short")

	assert.ErrorIs(t, err, password.ErrTooShort)
	assert.NoError(t, uc.ResetPassword(t.Context(), token, "a-new-strong-password"))
}

// Deactivation must hold even against a link sent before it.
func TestResetPassword_DisabledAccountStaysDisabled(t *testing.T) {
	t.Parallel()

	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc, _ := newRecoveryUseCase(repo, notifier)
	token := issueToken(t, uc, notifier)
	repo.user.IsActive = false

	err := uc.ResetPassword(t.Context(), token, "a-new-strong-password")

	assert.ErrorIs(t, err, ErrInvalidResetToken)
	assert.Empty(t, repo.passwordHash)
}
