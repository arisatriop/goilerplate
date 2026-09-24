package repository_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"goilerplate/internal/domain/auth"
	"goilerplate/internal/infrastructure/transaction"
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
	// The column means "NULL while active", so an active row must not store an empty string:
	// a WHERE revoked_reason IS NULL filter has to find it.
	var activeRowsWithNullReason int
	require.NoError(t, db.Raw(
		"SELECT count(*) FROM user_sessions WHERE id = ? AND revoked_reason IS NULL", session.ID,
	).Scan(&activeRowsWithNullReason).Error)
	assert.Equal(t, 1, activeRowsWithNullReason)
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

// The list is what a user acts on when signing a device out, so it must show exactly the
// sessions that can still be used: not revoked ones, not expired ones, and never someone else's.
func TestUserSession_ListActiveSessions(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	otherUserID := createTestUser(t, db)
	now := utils.Now()

	older := newSession(userID)
	older.LastUsedAt = now.Add(-time.Hour)
	recent := newSession(userID)
	revoked := newSession(userID)
	expired := newSession(userID)
	expired.ExpiresAt = now.Add(-time.Minute)
	someoneElses := newSession(otherUserID)

	for _, session := range []*auth.UserSession{older, recent, revoked, expired, someoneElses} {
		_, err := repo.CreateSession(ctx, session)
		require.NoError(t, err)
	}
	require.NoError(t, repo.RevokeSession(ctx, userID, revoked.ID, auth.RevokedReasonLogout))

	// Act
	got, err := repo.ListActiveSessions(ctx, userID)

	// Assert
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, recent.ID, got[0].ID, "most recently used first")
	assert.Equal(t, older.ID, got[1].ID)
	assert.Equal(t, "iPhone", got[0].DeviceName)
	assert.Equal(t, "203.0.113.7", got[0].IPAddress)
}

func TestUserSession_ListActiveSessionsEmpty(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)

	// Act
	got, err := repo.ListActiveSessions(context.Background(), createTestUser(t, db))

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, got, "an empty list, not nil, so the handler can return [] without a check")
	assert.Empty(t, got)
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

// Only one of many concurrent refreshes may claim the same refresh token, otherwise two
// clients would walk away believing they each own the session's next token.
func TestUserSession_ConcurrentRotateSucceedsOnce(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	session := newSession(createTestUser(t, db))
	_, err := repo.CreateSession(ctx, session)
	require.NoError(t, err)

	const racers = 20
	var wg sync.WaitGroup
	results := make(chan error, racers)

	// Act
	for range racers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- repo.RotateRefreshJTI(ctx, session.ID, session.RefreshJTI, utils.GenerateUUID())
		}()
	}
	wg.Wait()
	close(results)

	// Assert
	var succeeded int
	for err := range results {
		if err == nil {
			succeeded++
			continue
		}
		assert.ErrorIs(t, err, auth.ErrNotFound)
	}
	assert.Equal(t, 1, succeeded, "exactly one concurrent rotation may win")

	rotated, err := repo.GetSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.RefreshJTI, rotated.PreviousRefreshJTI, "the claimed token is kept for the grace window")
	assert.NotEqual(t, session.RefreshJTI, rotated.RefreshJTI)
	require.NotNil(t, rotated.RotatedAt)
	assert.True(t, rotated.IsActive)
	assert.WithinDuration(t, session.ExpiresAt, rotated.ExpiresAt, time.Second, "rotation must not extend the session")
}

func TestUserSession_RotateRejectsStaleOrUnusableSessions(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)

	revoked := newSession(userID)
	expired := newSession(userID)
	expired.ExpiresAt = utils.Now().Add(-time.Minute)
	live := newSession(userID)
	for _, s := range []*auth.UserSession{revoked, expired, live} {
		_, err := repo.CreateSession(ctx, s)
		require.NoError(t, err)
	}
	require.NoError(t, repo.RevokeSession(ctx, userID, revoked.ID, auth.RevokedReasonLogout))

	tests := []struct {
		name    string
		session *auth.UserSession
		jti     string
	}{
		{"revoked session", revoked, revoked.RefreshJTI},
		{"expired session", expired, expired.RefreshJTI},
		{"wrong jti", live, utils.GenerateUUID()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.RotateRefreshJTI(ctx, tt.session.ID, tt.jti, utils.GenerateUUID())
			assert.ErrorIs(t, err, auth.ErrNotFound)
		})
	}
}

// A login that fails after the session insert must leave nothing behind: the caller never
// received the tokens, so a session row would be one no client can ever present.
func TestUserSession_RollbackLeavesNoOrphanedSession(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	session := newSession(userID)
	txManager := transaction.NewGormTransaction(db)
	wantErr := errors.New("failure after the session was created")

	// Act
	err := txManager.Do(ctx, func(txCtx context.Context) error {
		txRepo := repo.WithTx(txCtx)

		if err := txRepo.UpdateUserLoginInfo(txCtx, userID, true); err != nil {
			return err
		}
		if _, err := txRepo.CreateSession(txCtx, session); err != nil {
			return err
		}

		// The session is visible inside the transaction...
		inTx, err := txRepo.GetSessionByID(txCtx, session.ID)
		require.NoError(t, err)
		require.NotNil(t, inTx)

		return wantErr
	})

	// Assert
	require.ErrorIs(t, err, wantErr)

	// ...but nothing survives the rollback, and last_login_at is unstamped with it
	orphan, err := repo.GetSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Nil(t, orphan, "a rolled-back login must leave no session")

	var stamped int
	require.NoError(t, db.Raw(
		"SELECT count(*) FROM users WHERE id = ? AND last_login_at IS NOT NULL", userID,
	).Scan(&stamped).Error)
	assert.Equal(t, 0, stamped, "the login stamp rolls back with the session")
}

// WithTx must fall back to the plain repository when the context carries no transaction,
// otherwise every non-transactional call would silently do nothing.
func TestUserSession_WithTxWithoutTransaction(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	session := newSession(createTestUser(t, db))

	_, err := repo.WithTx(ctx).CreateSession(ctx, session)
	require.NoError(t, err)

	got, err := repo.GetSessionByID(ctx, session.ID)
	require.NoError(t, err)
	assert.NotNil(t, got)
}
