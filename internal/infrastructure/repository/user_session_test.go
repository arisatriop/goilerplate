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

func newSession(userID string) *auth.UserSession {
	now := utils.Now()

	return &auth.UserSession{
		ID:         utils.GenerateUUID(),
		UserID:     userID,
		RefreshJTI: utils.GenerateUUID(),
		DeviceName: "iPhone",
		DeviceType: auth.DeviceTypeMobile,
		DeviceID:   "fp_0123456789ab",
		IPAddress:  "203.0.113.7",
		UserAgent:  "test-agent",
		IsActive:   true,
		ExpiresAt:  now.Add(7 * 24 * time.Hour),
		LastUsedAt: now,
		CreatedAt:  now,
	}
}

func TestUserSession_CreateAndGetByID(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	session := newSession(userID)

	// Act
	_, err := repo.CreateSession(ctx, session)
	require.NoError(t, err)

	got, err := repo.GetSessionByID(ctx, session.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, session.ID, got.ID)
	assert.Equal(t, session.UserID, got.UserID)
	assert.Equal(t, session.RefreshJTI, got.RefreshJTI)
	assert.Empty(t, got.PreviousRefreshJTI)
	assert.Nil(t, got.RotatedAt)
	assert.Equal(t, session.DeviceName, got.DeviceName)
	assert.Equal(t, session.DeviceType, got.DeviceType)
	assert.Equal(t, session.DeviceID, got.DeviceID)
	assert.Equal(t, session.IPAddress, got.IPAddress)
	assert.Equal(t, session.UserAgent, got.UserAgent)
	assert.True(t, got.IsActive)
	assert.Nil(t, got.RevokedAt)
	assert.Empty(t, got.RevokedReason)
	assert.True(t, got.IsValidSession())
	assert.WithinDuration(t, session.ExpiresAt, got.ExpiresAt, time.Second)
	assert.WithinDuration(t, session.CreatedAt, got.CreatedAt, time.Second)
}

// An empty IP address must be stored as NULL: the inet column rejects an empty string.
func TestUserSession_CreateWithoutIPAddress(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	session := newSession(createTestUser(t, db))
	session.IPAddress = ""

	// Act
	_, err := repo.CreateSession(ctx, session)
	require.NoError(t, err)

	got, err := repo.GetSessionByID(ctx, session.ID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Empty(t, got.IPAddress)
}

func TestUserSession_RevokeSession(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	session := newSession(userID)
	other := newSession(userID)
	_, err := repo.CreateSession(ctx, session)
	require.NoError(t, err)
	_, err = repo.CreateSession(ctx, other)
	require.NoError(t, err)

	// Act
	err = repo.RevokeSession(ctx, userID, session.ID, auth.RevokedReasonLogout)

	// Assert
	require.NoError(t, err)

	revoked, err := repo.GetSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.False(t, revoked.IsActive)
	assert.Equal(t, auth.RevokedReasonLogout, revoked.RevokedReason)
	require.NotNil(t, revoked.RevokedAt)
	assert.False(t, revoked.IsValidSession())

	// Logging out one device leaves the user's other sessions alone
	untouched, err := repo.GetSessionByID(ctx, other.ID)
	require.NoError(t, err)
	assert.True(t, untouched.IsActive)

	// Revoking again reports not found, so a concurrent logout cannot report success twice
	assert.ErrorIs(t, repo.RevokeSession(ctx, userID, session.ID, auth.RevokedReasonLogout), auth.ErrNotFound)
}

func TestUserSession_RevokeSessionOfAnotherUser(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	owner := createTestUser(t, db)
	attacker := createTestUser(t, db)
	session := newSession(owner)
	_, err := repo.CreateSession(ctx, session)
	require.NoError(t, err)

	// Act
	err = repo.RevokeSession(ctx, attacker, session.ID, auth.RevokedReasonLogout)

	// Assert
	assert.ErrorIs(t, err, auth.ErrNotFound)

	got, err := repo.GetSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.True(t, got.IsActive)
}

func TestUserSession_DeactivateUserSessions(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	first := newSession(userID)
	second := newSession(userID)
	_, err := repo.CreateSession(ctx, first)
	require.NoError(t, err)
	_, err = repo.CreateSession(ctx, second)
	require.NoError(t, err)

	// Act
	err = repo.DeactivateUserSessions(ctx, userID, auth.RevokedReasonLogoutAll)

	// Assert
	require.NoError(t, err)

	for _, id := range []string{first.ID, second.ID} {
		got, err := repo.GetSessionByID(ctx, id)
		require.NoError(t, err)
		assert.False(t, got.IsActive)
		assert.Equal(t, auth.RevokedReasonLogoutAll, got.RevokedReason)
		assert.NotNil(t, got.RevokedAt)
	}
}

func TestUserSession_GetSessionByIDUnknown(t *testing.T) {
	// Arrange
	repo, _ := newTestRepository(t)

	// Act
	got, err := repo.GetSessionByID(context.Background(), utils.GenerateUUID())

	// Assert
	require.NoError(t, err)
	assert.Nil(t, got)
}
