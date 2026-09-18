package repository_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const testMaxAttempts = 5

func attemptsAndLock(t *testing.T, db *gorm.DB, userID string) (int, bool) {
	t.Helper()

	var row struct {
		FailedLoginAttempts int
		LockedUntil         *time.Time
	}
	require.NoError(t, db.Raw(
		"SELECT failed_login_attempts, locked_until FROM users WHERE id = ?", userID,
	).Scan(&row).Error)

	return row.FailedLoginAttempts, row.LockedUntil != nil
}

// The lock must land on attempt N, not N+1. Counting the pre-increment value gives an attacker
// one extra guess for free, which is the whole margin a small threshold provides.
func TestRegisterFailedLogin_LocksExactlyOnAttemptN(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	lockUntil := utils.Now().Add(10 * time.Minute)

	// Act & Assert
	for attempt := 1; attempt < testMaxAttempts; attempt++ {
		locked, err := repo.RegisterFailedLogin(ctx, userID, testMaxAttempts, lockUntil)
		require.NoError(t, err)
		assert.False(t, locked, "attempt %d must not lock", attempt)

		count, hasLock := attemptsAndLock(t, db, userID)
		assert.Equal(t, attempt, count)
		assert.False(t, hasLock)
	}

	locked, err := repo.RegisterFailedLogin(ctx, userID, testMaxAttempts, lockUntil)
	require.NoError(t, err)
	assert.True(t, locked, "attempt %d must lock", testMaxAttempts)

	count, hasLock := attemptsAndLock(t, db, userID)
	assert.Equal(t, testMaxAttempts, count)
	assert.True(t, hasLock)
}

// Simultaneous guesses must not each read the same stale counter and walk past the threshold.
func TestRegisterFailedLogin_ConcurrentAttemptsAllCount(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	lockUntil := utils.Now().Add(10 * time.Minute)

	const attempts = 20
	var wg sync.WaitGroup
	lockedCount := make(chan bool, attempts)

	// Act
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locked, err := repo.RegisterFailedLogin(ctx, userID, testMaxAttempts, lockUntil)
			if err == nil {
				lockedCount <- locked
			}
		}()
	}
	wg.Wait()
	close(lockedCount)

	// Assert
	var reportedLocked int
	for locked := range lockedCount {
		if locked {
			reportedLocked++
		}
	}

	count, hasLock := attemptsAndLock(t, db, userID)
	assert.Equal(t, attempts, count, "every concurrent attempt must be counted exactly once")
	assert.True(t, hasLock)
	assert.Equal(t, attempts-testMaxAttempts+1, reportedLocked,
		"every attempt from the Nth onwards reports the account as locked")
}

func TestRegisterFailedLogin_UnknownUser(t *testing.T) {
	repo, _ := newTestRepository(t)

	_, err := repo.RegisterFailedLogin(
		context.Background(), utils.GenerateUUID(), testMaxAttempts, utils.Now().Add(time.Minute))

	assert.ErrorIs(t, err, auth.ErrNotFound)
}
