package repository_test

import (
	"context"
	"testing"
	"time"

	"goilerplate/internal/infrastructure/repository"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// insertSession writes one session with the exact state the cleanup predicate reads.
func insertSession(t *testing.T, db *gorm.DB, userID string, isActive bool, expiresAt time.Time, revokedAt *time.Time) string {
	t.Helper()

	id := utils.GenerateUUID()
	err := db.Exec(`
		INSERT INTO user_sessions (id, user_id, refresh_jti, is_active, expires_at, last_used_at, revoked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, userID, utils.GenerateUUID(), isActive, expiresAt, utils.Now(), revokedAt).Error
	require.NoError(t, err)

	return id
}

func sessionExists(t *testing.T, db *gorm.DB, id string) bool {
	t.Helper()

	var count int64
	require.NoError(t, db.Raw("SELECT count(*) FROM user_sessions WHERE id = ?", id).Scan(&count).Error)
	return count > 0
}

// This statement deletes login history, so the test that matters is not "does it delete" but
// "what does it refuse to delete".
func TestDeleteExpiredSessions_RemovesOnlyDeadRowsPastTheCutoff(t *testing.T) {
	_, db := newTestRepository(t)
	userID := createTestUser(t, db)
	cleanup := repository.NewCleanup(db)

	now := utils.Now()
	retention := 30 * 24 * time.Hour
	cutoff := now.Add(-retention)

	longAgo := now.Add(-60 * 24 * time.Hour)
	recently := now.Add(-time.Hour)

	// Eligible: dead, and dead for longer than the retention.
	oldExpired := insertSession(t, db, userID, true, longAgo, nil)
	oldRevoked := insertSession(t, db, userID, false, now.Add(24*time.Hour), &longAgo)

	// Not eligible, each for a different reason.
	live := insertSession(t, db, userID, true, now.Add(24*time.Hour), nil)
	recentlyExpired := insertSession(t, db, userID, true, recently, nil)
	recentlyRevoked := insertSession(t, db, userID, false, now.Add(24*time.Hour), &recently)
	// Revoked long ago but the row was never marked inactive — the coalesce still finds it.
	revokedLongAgoStillFlaggedActive := insertSession(t, db, userID, false, longAgo, &longAgo)

	removed, err := cleanup.DeleteExpiredSessions(context.Background(), cutoff, 100)

	require.NoError(t, err)
	assert.Equal(t, int64(3), removed)

	assert.False(t, sessionExists(t, db, oldExpired), "expired past the retention")
	assert.False(t, sessionExists(t, db, oldRevoked), "revoked past the retention")
	assert.False(t, sessionExists(t, db, revokedLongAgoStillFlaggedActive))

	assert.True(t, sessionExists(t, db, live), "a session still in use must never be deleted")
	assert.True(t, sessionExists(t, db, recentlyExpired), "still inside the retention window")
	assert.True(t, sessionExists(t, db, recentlyRevoked), "still inside the retention window")
}

// The second condition exists for exactly this: a retention long enough to put the cutoff in the
// future would otherwise make every live session eligible. It must delete nothing instead.
func TestDeleteExpiredSessions_AFutureCutoffDeletesNoLiveSessions(t *testing.T) {
	_, db := newTestRepository(t)
	userID := createTestUser(t, db)
	cleanup := repository.NewCleanup(db)

	now := utils.Now()
	live := insertSession(t, db, userID, true, now.Add(24*time.Hour), nil)
	alsoLive := insertSession(t, db, userID, true, now.Add(72*time.Hour), nil)

	removed, err := cleanup.DeleteExpiredSessions(context.Background(), now.Add(365*24*time.Hour), 100)

	require.NoError(t, err)
	assert.Zero(t, removed)
	assert.True(t, sessionExists(t, db, live))
	assert.True(t, sessionExists(t, db, alsoLive))
}

func TestDeleteExpiredSessions_RespectsTheBatchSize(t *testing.T) {
	_, db := newTestRepository(t)
	userID := createTestUser(t, db)
	cleanup := repository.NewCleanup(db)

	now := utils.Now()
	longAgo := now.Add(-60 * 24 * time.Hour)
	for range 5 {
		insertSession(t, db, userID, true, longAgo, nil)
	}

	cutoff := now.Add(-30 * 24 * time.Hour)

	first, err := cleanup.DeleteExpiredSessions(context.Background(), cutoff, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), first)

	second, err := cleanup.DeleteExpiredSessions(context.Background(), cutoff, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), second)

	third, err := cleanup.DeleteExpiredSessions(context.Background(), cutoff, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(1), third, "the short batch is what tells the caller to stop")
}

func TestDeleteFinishedOneTimeTokens_KeepsPendingTokens(t *testing.T) {
	_, db := newTestRepository(t)
	userID := createTestUser(t, db)
	cleanup := repository.NewCleanup(db)

	now := utils.Now()
	longAgo := now.Add(-60 * 24 * time.Hour)
	cutoff := now.Add(-30 * 24 * time.Hour)

	insert := func(expiresAt time.Time, usedAt *time.Time) string {
		id := utils.GenerateUUID()
		err := db.Exec(`
			INSERT INTO one_time_tokens (id, user_id, token_type, token_hash, expires_at, used_at)
			VALUES (?, ?, 'password_reset', ?, ?, ?)`,
			id, userID, utils.GenerateUUID(), expiresAt, usedAt).Error
		require.NoError(t, err)
		return id
	}
	exists := func(id string) bool {
		var count int64
		require.NoError(t, db.Raw("SELECT count(*) FROM one_time_tokens WHERE id = ?", id).Scan(&count).Error)
		return count > 0
	}

	oldUsed := insert(now.Add(time.Hour), &longAgo)
	oldExpired := insert(longAgo, nil)

	// A token that has neither been used nor expired is a password reset someone is still
	// waiting on. Deleting it would break the link in their inbox.
	pending := insert(now.Add(time.Hour), nil)
	recentlyUsed := insert(now.Add(time.Hour), &now)

	removed, err := cleanup.DeleteFinishedOneTimeTokens(context.Background(), cutoff, 100)

	require.NoError(t, err)
	assert.Equal(t, int64(2), removed)
	assert.False(t, exists(oldUsed))
	assert.False(t, exists(oldExpired))
	assert.True(t, exists(pending), "a pending reset must survive")
	assert.True(t, exists(recentlyUsed))
}
