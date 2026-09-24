package auth

import (
	"context"
	"testing"
	"time"

	"goilerplate/pkg/apperr"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const correctPassword = "correct-horse-battery-staple"

// validatorRepo records what the login path did, so a test can tell "counted the attempt"
// apart from "never reached the password check".
type validatorRepo struct {
	Repository
	user            *User
	failedLogins    int
	lockResets      int
	registerReturns bool
}

func (r *validatorRepo) GetUserByEmail(context.Context, string) (*User, error) {
	if r.user == nil {
		return nil, nil
	}
	copied := *r.user
	return &copied, nil
}

func (r *validatorRepo) RegisterFailedLogin(context.Context, string, int, time.Time) (bool, error) {
	r.failedLogins++
	return r.registerReturns, nil
}

func (r *validatorRepo) ResetExpiredLock(context.Context, string) error {
	r.lockResets++
	return nil
}

func testUser(t *testing.T) *User {
	t.Helper()

	hash, err := utils.HashPassword(correctPassword)
	require.NoError(t, err)

	return &User{
		ID:           "u1",
		Email:        "user@example.test",
		Name:         "User",
		PasswordHash: hash,
		IsActive:     true,
	}
}

func newValidator(repo *validatorRepo) *UserValidator {
	return NewUserValidator(repo, Lockout{MaxAttempts: 5, Duration: 10 * time.Minute})
}

func assertClientErr(t *testing.T, err error, want *apperr.Error) {
	t.Helper()

	require.ErrorIs(t, err, want)
	assert.Equal(t, want.Message, err.Error(), "the message a client sees")
}

// An unregistered email and a wrong password must be indistinguishable: same status, same
// message. Anything else turns the login form into a directory of who has an account.
func TestValidateUserForLogin_UnknownEmailMatchesWrongPassword(t *testing.T) {
	// Arrange
	unknown := &validatorRepo{user: nil}
	known := &validatorRepo{user: testUser(t)}

	// Act
	_, unknownErr := newValidator(unknown).ValidateUserForLogin(context.Background(), "nobody@example.test", "whatever")
	_, wrongErr := newValidator(known).ValidateUserForLogin(context.Background(), "user@example.test", "wrong-password")

	// Assert
	assertClientErr(t, unknownErr, ErrInvalidCredentials)
	assertClientErr(t, wrongErr, ErrInvalidCredentials)
	assert.Equal(t, unknownErr.Error(), wrongErr.Error())

	assert.Zero(t, unknown.failedLogins, "there is no account to count an attempt against")
	assert.Equal(t, 1, known.failedLogins, "a wrong password against a real account is counted")
}

// While locked, the right password must look exactly like the wrong one. If the correct
// password produced a different answer, the lockout would just slow the attacker down instead
// of stopping them learning when they had guessed right.
func TestValidateUserForLogin_LockedHidesWhetherPasswordIsRight(t *testing.T) {
	lockedUntil := utils.Now().Add(5 * time.Minute)

	for _, password := range []string{correctPassword, "wrong-password"} {
		user := testUser(t)
		user.LockedUntil = &lockedUntil
		user.FailedLoginAttempts = 5
		repo := &validatorRepo{user: user}

		got, err := newValidator(repo).ValidateUserForLogin(context.Background(), user.Email, password)

		assert.Nil(t, got)
		assertClientErr(t, err, ErrAccountLocked)
		assert.Zero(t, repo.failedLogins, "a locked account does not evaluate or count the password")
	}
}

// Once the window passes the account is usable on this very attempt, not the next one.
func TestValidateUserForLogin_ExpiredLockIsCleared(t *testing.T) {
	expired := utils.Now().Add(-time.Minute)
	user := testUser(t)
	user.LockedUntil = &expired
	user.FailedLoginAttempts = 5
	repo := &validatorRepo{user: user}

	got, err := newValidator(repo).ValidateUserForLogin(context.Background(), user.Email, correctPassword)

	require.NoError(t, err)
	assert.Equal(t, "u1", got.ID)
	assert.Equal(t, 1, repo.lockResets)
}

// A disabled account is only revealed to someone who already proved they know the password,
// so the message cannot be used to enumerate accounts.
func TestValidateUserForLogin_DisabledOnlyAfterCorrectPassword(t *testing.T) {
	disabled := testUser(t)
	disabled.IsActive = false

	withWrong := &validatorRepo{user: disabled}
	_, err := newValidator(withWrong).ValidateUserForLogin(context.Background(), disabled.Email, "wrong-password")
	assertClientErr(t, err, ErrInvalidCredentials)

	withRight := &validatorRepo{user: disabled}
	_, err = newValidator(withRight).ValidateUserForLogin(context.Background(), disabled.Email, correctPassword)
	assertClientErr(t, err, ErrAccountDisabled)
}

func TestValidateUserForLogin_Success(t *testing.T) {
	repo := &validatorRepo{user: testUser(t)}

	got, err := newValidator(repo).ValidateUserForLogin(context.Background(), "user@example.test", correctPassword)

	require.NoError(t, err)
	assert.Equal(t, "u1", got.ID)
	assert.Zero(t, repo.failedLogins)
}
