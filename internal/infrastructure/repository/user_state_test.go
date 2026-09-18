package repository_test

import (
	"context"
	"testing"
	"time"

	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A password change must turn out every other device while leaving the one making the change
// signed in — otherwise users get logged out of the browser they are standing in front of.
func TestRevokeOtherUserSessions_KeepsTheCallersSession(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	otherUserID := createTestUser(t, db)

	current := newSession(userID)
	otherDevice := newSession(userID)
	thirdDevice := newSession(userID)
	bystander := newSession(otherUserID)
	for _, s := range []*auth.UserSession{current, otherDevice, thirdDevice, bystander} {
		_, err := repo.CreateSession(ctx, s)
		require.NoError(t, err)
	}

	// Act
	err := repo.RevokeOtherUserSessions(ctx, userID, current.ID, auth.RevokedReasonPasswordChange)

	// Assert
	require.NoError(t, err)

	kept, err := repo.GetSessionByID(ctx, current.ID)
	require.NoError(t, err)
	assert.True(t, kept.IsActive, "the session making the change stays signed in")
	assert.Empty(t, kept.RevokedReason)

	for _, id := range []string{otherDevice.ID, thirdDevice.ID} {
		revoked, err := repo.GetSessionByID(ctx, id)
		require.NoError(t, err)
		assert.False(t, revoked.IsActive)
		assert.Equal(t, auth.RevokedReasonPasswordChange, revoked.RevokedReason)
	}

	untouched, err := repo.GetSessionByID(ctx, bystander.ID)
	require.NoError(t, err)
	assert.True(t, untouched.IsActive, "another user's sessions are none of this call's business")
}

// An empty keepSessionID revokes everything, which is what deactivation needs.
func TestRevokeOtherUserSessions_EmptyKeepRevokesAll(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)

	first, second := newSession(userID), newSession(userID)
	for _, s := range []*auth.UserSession{first, second} {
		_, err := repo.CreateSession(ctx, s)
		require.NoError(t, err)
	}

	require.NoError(t, repo.RevokeOtherUserSessions(ctx, userID, "", auth.RevokedReasonAdmin))

	for _, id := range []string{first.ID, second.ID} {
		got, err := repo.GetSessionByID(ctx, id)
		require.NoError(t, err)
		assert.False(t, got.IsActive)
		assert.Equal(t, auth.RevokedReasonAdmin, got.RevokedReason)
	}
}

// Changing the password also clears a lockout: the person proved they own the account, so
// making them wait out someone else's failed guesses would be punishing the wrong party.
func TestUpdateUserPassword_StampsAndClearsLockout(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	lockUntil := utils.Now().Add(10 * time.Minute)
	for range 5 {
		_, err := repo.RegisterFailedLogin(ctx, userID, 5, lockUntil)
		require.NoError(t, err)
	}

	var before time.Time
	require.NoError(t, db.Raw("SELECT password_changed_at FROM users WHERE id = ?", userID).Scan(&before).Error)

	// Act
	require.NoError(t, repo.UpdateUserPassword(ctx, userID, "new-hash"))

	// Assert
	var row struct {
		PasswordHash        string
		PasswordChangedAt   time.Time
		FailedLoginAttempts int
		LockedUntil         *time.Time
	}
	require.NoError(t, db.Raw(
		"SELECT password_hash, password_changed_at, failed_login_attempts, locked_until FROM users WHERE id = ?",
		userID).Scan(&row).Error)

	assert.Equal(t, "new-hash", row.PasswordHash)
	assert.True(t, row.PasswordChangedAt.After(before), "password_changed_at is re-stamped")
	assert.Zero(t, row.FailedLoginAttempts)
	assert.Nil(t, row.LockedUntil)
}

func TestUpdateUserPassword_UnknownUser(t *testing.T) {
	repo, _ := newTestRepository(t)

	err := repo.UpdateUserPassword(context.Background(), utils.GenerateUUID(), "hash")

	assert.ErrorIs(t, err, auth.ErrNotFound)
}

func TestSetUserActive(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)

	require.NoError(t, repo.SetUserActive(ctx, userID, false))

	var active bool
	require.NoError(t, db.Raw("SELECT is_active FROM users WHERE id = ?", userID).Scan(&active).Error)
	assert.False(t, active)

	assert.ErrorIs(t, repo.SetUserActive(ctx, utils.GenerateUUID(), false), auth.ErrNotFound)
}
