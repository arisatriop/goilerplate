package auth

import (
	"context"
	"regexp"
	"testing"
	"time"

	"goilerplate/pkg/hash"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testOTPMaxAttempts = 3

var testOTPSecret = []byte("unit-test-otp-secret-of-32-bytes!")

func newVerificationUseCase(repo *recoveryRepo, notifier Notifier) *authUseCase {
	uc, _ := newRecoveryUseCase(repo, notifier)
	uc.email.OTP = OTPPolicy{
		Secret:         testOTPSecret,
		TTL:            15 * time.Minute,
		MaxAttempts:    testOTPMaxAttempts,
		ResendCooldown: time.Minute,
	}
	return uc
}

// sendCode runs SendEmailVerification and returns the code the user would find in their inbox.
func sendCode(t *testing.T, uc *authUseCase, notifier *recordingNotifier) string {
	t.Helper()
	require.NoError(t, uc.SendEmailVerification(t.Context(), "ana@example.org", RequestOrigin{}))
	require.NotEmpty(t, notifier.codes)
	return notifier.codes[len(notifier.codes)-1].Code
}

// wrongCode is any 6-digit code other than code.
func wrongCode(code string) string {
	if code == "000000" {
		return "000001"
	}
	return "000000"
}

func TestSendEmailVerification_StoresOnlyAKeyedHashOfTheCode(t *testing.T) {
	t.Parallel()
	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc := newVerificationUseCase(repo, notifier)

	code := sendCode(t, uc, notifier)

	assert.Regexp(t, `^\d{6}$`, code)
	require.Len(t, repo.tokens, 1)
	stored := repo.tokens[0]
	assert.Equal(t, OneTimeTokenEmailVerification, stored.TokenType)
	assert.NotEmpty(t, stored.ID, "the ID is chosen before insert, because the hash is keyed by it")
	assert.Equal(t, hash.Keyed(testOTPSecret, stored.ID+":"+code), stored.TokenHash)
	assert.NotEqual(t, hash.Token(code), stored.TokenHash, "an unkeyed digest of six digits is reversible")
	assert.Equal(t, "ana@example.org", notifier.codes[0].Email)
}

// The same promise as forgot-password: nothing about the response says whether the address is
// registered, active, or already verified.
func TestSendEmailVerification_AnswersTheSameWhenNothingIsSent(t *testing.T) {
	t.Parallel()
	disabled := activeUser()
	disabled.IsActive = false
	verified := activeUser()
	verified.EmailVerified = true

	tests := []struct {
		name  string
		user  *User
		email string
	}{
		{"unknown address", activeUser(), "nobody@example.org"},
		{"disabled account", disabled, "ana@example.org"},
		{"already verified", verified, "ana@example.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &recoveryRepo{user: tt.user}
			notifier := &recordingNotifier{}
			uc := newVerificationUseCase(repo, notifier)

			err := uc.SendEmailVerification(t.Context(), tt.email, RequestOrigin{})

			assert.NoError(t, err)
			assert.Empty(t, notifier.codes)
			assert.Empty(t, repo.tokens)
		})
	}
}

func TestSendEmailVerification_CooldownThenSupersede(t *testing.T) {
	t.Parallel()
	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc := newVerificationUseCase(repo, notifier)

	first := sendCode(t, uc, notifier)
	require.NoError(t, uc.SendEmailVerification(t.Context(), "ana@example.org", RequestOrigin{}))
	assert.Len(t, notifier.codes, 1, "a second request inside the cooldown sends nothing")

	repo.age(2 * time.Minute)
	second := sendCode(t, uc, notifier)

	if first != second {
		assert.ErrorIs(t, uc.VerifyEmail(t.Context(), "ana@example.org", first), ErrInvalidVerificationCode,
			"the earlier code stops working once a new one is sent")
	}
	assert.NoError(t, uc.VerifyEmail(t.Context(), "ana@example.org", second))
}

func TestVerifyEmail_MarksTheAddressVerifiedOnce(t *testing.T) {
	t.Parallel()
	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc := newVerificationUseCase(repo, notifier)
	code := sendCode(t, uc, notifier)

	require.NoError(t, uc.VerifyEmail(t.Context(), "ana@example.org", code))

	assert.True(t, repo.verified)
	assert.True(t, repo.tokens[0].IsUsed())
	assert.ErrorIs(t, uc.VerifyEmail(t.Context(), "ana@example.org", code), ErrInvalidVerificationCode,
		"a code works once")
}

// The ceiling the previous roadmap deferred: once a code has taken MaxAttempts wrong guesses it
// stops working, and the right code no longer helps.
func TestVerifyEmail_CodeIsVoidAfterMaxAttempts(t *testing.T) {
	t.Parallel()
	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc := newVerificationUseCase(repo, notifier)
	code := sendCode(t, uc, notifier)

	for range testOTPMaxAttempts {
		assert.ErrorIs(t, uc.VerifyEmail(t.Context(), "ana@example.org", wrongCode(code)), ErrInvalidVerificationCode)
	}

	assert.ErrorIs(t, uc.VerifyEmail(t.Context(), "ana@example.org", code), ErrInvalidVerificationCode)
	assert.False(t, repo.verified)
	assert.True(t, repo.tokens[0].IsExpired(), "the exhausted code is expired, so the next step is asking for a new one")
}

// A wrong guess short of the ceiling costs one attempt and leaves the code usable.
func TestVerifyEmail_RightCodeAfterAWrongOne(t *testing.T) {
	t.Parallel()
	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc := newVerificationUseCase(repo, notifier)
	code := sendCode(t, uc, notifier)

	assert.ErrorIs(t, uc.VerifyEmail(t.Context(), "ana@example.org", wrongCode(code)), ErrInvalidVerificationCode)
	assert.Equal(t, 1, repo.tokens[0].Attempts)

	assert.NoError(t, uc.VerifyEmail(t.Context(), "ana@example.org", code))
}

// Attempts are counted before the comparison. A code whose count is already past the ceiling —
// what concurrent guesses leave behind — refuses even the right code, so a burst of parallel
// guesses gets no more tries than a sequence of them.
func TestVerifyEmail_CountIsCheckedBeforeTheCode(t *testing.T) {
	t.Parallel()
	repo := &recoveryRepo{user: activeUser()}
	notifier := &recordingNotifier{}
	uc := newVerificationUseCase(repo, notifier)
	code := sendCode(t, uc, notifier)
	repo.tokens[0].Attempts = testOTPMaxAttempts

	err := uc.VerifyEmail(t.Context(), "ana@example.org", code)

	assert.ErrorIs(t, err, ErrInvalidVerificationCode)
	assert.False(t, repo.verified)
}

func TestVerifyEmail_NothingToVerifyAgainst(t *testing.T) {
	t.Parallel()
	disabled := activeUser()
	disabled.IsActive = false

	tests := []struct {
		name  string
		user  *User
		email string
	}{
		{"unknown address", activeUser(), "nobody@example.org"},
		{"no code was ever sent", activeUser(), "ana@example.org"},
		{"disabled account", disabled, "ana@example.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &recoveryRepo{user: tt.user}
			uc := newVerificationUseCase(repo, &recordingNotifier{})

			err := uc.VerifyEmail(t.Context(), tt.email, "123456")

			assert.ErrorIs(t, err, ErrInvalidVerificationCode)
			assert.False(t, repo.verified)
		})
	}
}

func TestEmailVerification_UnavailableWithoutANotifier(t *testing.T) {
	t.Parallel()
	uc := newVerificationUseCase(&recoveryRepo{user: activeUser()}, nil)
	uc.email.Notifier = nil

	assert.ErrorIs(t, uc.SendEmailVerification(t.Context(), "ana@example.org", RequestOrigin{}), errEmailFlowsUnavailable)
	assert.ErrorIs(t, uc.VerifyEmail(t.Context(), "ana@example.org", "123456"), errEmailFlowsUnavailable)
}

func TestGenerateOTP_IsSixDigitsIncludingLeadingZeros(t *testing.T) {
	t.Parallel()
	sixDigits := regexp.MustCompile(`^\d{6}$`)

	sawLeadingZero := false
	for range 2000 {
		code, err := generateOTP()
		require.NoError(t, err)
		require.Regexp(t, sixDigits, code)
		sawLeadingZero = sawLeadingZero || code[0] == '0'
	}

	// Roughly 10% of codes start with 0. Missing them all in 2000 draws (p ≈ 1e-92) would mean
	// the range or the padding is wrong.
	assert.True(t, sawLeadingZero)
}

// verificationRepo serves one user to the login validator.
type verificationRepo struct {
	Repository
	user *User
}

func (r *verificationRepo) GetUserByEmail(context.Context, string) (*User, error) {
	copied := *r.user
	return &copied, nil
}

func (r *verificationRepo) UpdateUserLoginInfo(context.Context, string, bool) error { return nil }

func (r *verificationRepo) RegisterFailedLogin(context.Context, string, int, time.Time) (bool, error) {
	return false, nil
}

// With verification required, an unverified address is refused — but only after the password,
// so the refusal tells nothing to someone who does not know it.
func TestValidateUserForLogin_UnverifiedEmail(t *testing.T) {
	t.Parallel()
	hashed, err := utils.HashPassword("the-right-password")
	require.NoError(t, err)

	tests := []struct {
		name     string
		required bool
		verified bool
		password string
		wantErr  error
	}{
		{"required, unverified, right password", true, false, "the-right-password", ErrEmailNotVerified},
		{"required, unverified, wrong password", true, false, "a-wrong-password", ErrInvalidCredentials},
		{"required, verified", true, true, "the-right-password", nil},
		{"not required, unverified", false, false, "the-right-password", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			user := activeUser()
			user.PasswordHash = hashed
			user.EmailVerified = tt.verified
			validator := NewUserValidator(&verificationRepo{user: user}, Lockout{MaxAttempts: 5, Duration: time.Minute})
			validator.requireVerifiedEmail = tt.required

			_, err := validator.ValidateUserForLogin(t.Context(), user.Email, tt.password)

			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
