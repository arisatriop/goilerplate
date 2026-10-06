package auth

import (
	"testing"
	"time"

	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const changeTestPassword = "the-current-password"

// newEmailChangeUseCase is a use case over one active user with a known password.
func newEmailChangeUseCase(t *testing.T) (*authUseCase, *recoveryRepo, *recordingNotifier) {
	t.Helper()
	user := activeUser()
	user.EmailVerified = true
	hash, err := utils.HashPassword(changeTestPassword)
	require.NoError(t, err)
	user.PasswordHash = hash

	repo := &recoveryRepo{user: user}
	notifier := &recordingNotifier{}
	return newVerificationUseCase(repo, notifier), repo, notifier
}

// requestChange asks to move to newEmail and returns the code sent there.
func requestChange(t *testing.T, uc *authUseCase, notifier *recordingNotifier, newEmail string) string {
	t.Helper()
	require.NoError(t, uc.RequestEmailChange(t.Context(), "u1", changeTestPassword, newEmail, RequestOrigin{}))
	require.NotEmpty(t, notifier.changeCodes)
	return notifier.changeCodes[len(notifier.changeCodes)-1].Code
}

func TestEmailChange_MovesTheAccountAndWarnsTheOldAddress(t *testing.T) {
	t.Parallel()
	uc, repo, notifier := newEmailChangeUseCase(t)

	code := requestChange(t, uc, notifier, "ana.new@example.org")

	assert.Equal(t, "ana.new@example.org", notifier.changeCodes[0].NewEmail, "the code goes to the new inbox")
	assert.Equal(t, "ana@example.org", repo.user.Email, "nothing changes before the code is confirmed")
	require.Len(t, repo.tokens, 1)
	assert.Equal(t, "ana.new@example.org", repo.tokens[0].NewEmail)

	require.NoError(t, uc.ConfirmEmailChange(t.Context(), "u1", code))

	assert.Equal(t, "ana.new@example.org", repo.user.Email)
	assert.True(t, repo.user.EmailVerified, "the code proved the user reads the new inbox")
	require.Len(t, notifier.changed, 1)
	assert.Equal(t, EmailChangedNotice{OldEmail: "ana@example.org", NewEmail: "ana.new@example.org", Name: "Ana"}, notifier.changed[0])
	assert.ErrorIs(t, uc.ConfirmEmailChange(t.Context(), "u1", code), ErrInvalidVerificationCode, "a code works once")
}

func TestRequestEmailChange_Refusals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		newEmail string
		wantErr  error
	}{
		{"wrong current password", "not-the-password", "ana.new@example.org", ErrInvalidCredentials},
		{"same address", changeTestPassword, "ana@example.org", ErrEmailUnchanged},
		{"same address in another case", changeTestPassword, " ANA@example.org", ErrEmailUnchanged},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo, notifier := newEmailChangeUseCase(t)

			err := uc.RequestEmailChange(t.Context(), "u1", tt.password, tt.newEmail, RequestOrigin{})

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Empty(t, notifier.changeCodes)
			assert.Empty(t, repo.tokens)
		})
	}
}

// A signed-in user must not be able to probe which addresses are registered: a taken address is
// answered like a free one, and no code is sent anywhere.
func TestRequestEmailChange_TakenAddressLooksFree(t *testing.T) {
	t.Parallel()
	uc, repo, notifier := newEmailChangeUseCase(t)
	repo.others = []*User{{ID: "u2", Email: "taken@example.org", IsActive: true}}

	err := uc.RequestEmailChange(t.Context(), "u1", changeTestPassword, "taken@example.org", RequestOrigin{})

	assert.NoError(t, err)
	assert.Empty(t, notifier.changeCodes)
	assert.Empty(t, repo.tokens)
}

func TestRequestEmailChange_CooldownThenSupersede(t *testing.T) {
	t.Parallel()
	uc, repo, notifier := newEmailChangeUseCase(t)

	first := requestChange(t, uc, notifier, "first@example.org")
	require.NoError(t, uc.RequestEmailChange(t.Context(), "u1", changeTestPassword, "second@example.org", RequestOrigin{}))
	assert.Len(t, notifier.changeCodes, 1, "a second request inside the cooldown sends nothing")

	repo.age(2 * time.Minute)
	second := requestChange(t, uc, notifier, "second@example.org")

	if first != second {
		assert.ErrorIs(t, uc.ConfirmEmailChange(t.Context(), "u1", first), ErrInvalidVerificationCode)
	}
	require.NoError(t, uc.ConfirmEmailChange(t.Context(), "u1", second))
	assert.Equal(t, "second@example.org", repo.user.Email, "only the newest request can complete")
}

func TestConfirmEmailChange_SharesTheAttemptCeiling(t *testing.T) {
	t.Parallel()
	uc, repo, notifier := newEmailChangeUseCase(t)
	code := requestChange(t, uc, notifier, "ana.new@example.org")

	for range testOTPMaxAttempts {
		assert.ErrorIs(t, uc.ConfirmEmailChange(t.Context(), "u1", wrongCode(code)), ErrInvalidVerificationCode)
	}

	assert.ErrorIs(t, uc.ConfirmEmailChange(t.Context(), "u1", code), ErrInvalidVerificationCode)
	assert.Equal(t, "ana@example.org", repo.user.Email)
	assert.Empty(t, notifier.changed)
}

// The address was free when the code was sent and taken before it was confirmed. The database
// constraint decides, and the caller is told: they own the new inbox, so this reveals nothing
// they could not learn from it.
func TestConfirmEmailChange_AddressTakenMeanwhile(t *testing.T) {
	t.Parallel()
	uc, repo, notifier := newEmailChangeUseCase(t)
	code := requestChange(t, uc, notifier, "ana.new@example.org")
	repo.updateErr = ErrEmailAlreadyRegistered

	err := uc.ConfirmEmailChange(t.Context(), "u1", code)

	assert.ErrorIs(t, err, ErrEmailAlreadyRegistered)
	assert.Equal(t, "ana@example.org", repo.user.Email)
	assert.Empty(t, notifier.changed)
}

func TestSendAccountExistsNotice(t *testing.T) {
	t.Parallel()
	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc := newVerificationUseCase(repo, notifier)

	require.NoError(t, uc.SendAccountExistsNotice(t.Context(), "ana@example.org"))
	require.NoError(t, uc.SendAccountExistsNotice(t.Context(), "gone@example.org"), "an address deleted meanwhile has nobody to tell")

	assert.Equal(t, []AccountExistsNotice{{Email: "ana@example.org", Name: "Ana"}}, notifier.exists)
}
