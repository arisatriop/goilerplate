package repository_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"goilerplate/internal/domain/auth"
	"goilerplate/internal/infrastructure/repository"
	"goilerplate/pkg/migration"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const migrationsDir = "../../migrations"

// newTestRepository connects to POSTGRES_TEST_DSN and applies the migrations.
// Database tests are skipped when POSTGRES_TEST_DSN is not set.
func newTestRepository(t *testing.T) (auth.Repository, *gorm.DB) {
	t.Helper()

	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN not set; skipping PostgreSQL integration test")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: utils.Now,
		Logger:  gormlogger.Discard,
	})
	require.NoError(t, err)
	require.NoError(t, migration.NewMigrator(db, nil).Up(migrationsDir))

	return repository.NewAuth(db), db
}

// createTestUser inserts a user and removes it (and its tokens, via ON DELETE CASCADE) afterwards.
func createTestUser(t *testing.T, db *gorm.DB) string {
	t.Helper()

	userID := utils.GenerateUUID()
	err := db.Exec(`INSERT INTO users (id, name, email, password_hash, created_by, updated_by)
		VALUES (?, 'One-time token test', ?, 'x', 'test', 'test')`,
		userID, userID+"@example.test").Error
	require.NoError(t, err)

	t.Cleanup(func() { db.Exec("DELETE FROM users WHERE id = ?", userID) })
	return userID
}

func newToken(userID, hash string, expiresIn time.Duration) *auth.OneTimeToken {
	return &auth.OneTimeToken{
		UserID:    userID,
		TokenType: auth.OneTimeTokenPasswordReset,
		TokenHash: hash,
		ExpiresAt: utils.Now().Add(expiresIn),
		IPAddress: "203.0.113.7",
		UserAgent: "test-agent",
	}
}

func TestOneTimeToken_CreateAndGetLatestActive(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)

	older := newToken(userID, utils.GenerateUUID(), time.Hour)
	require.NoError(t, repo.CreateOneTimeToken(ctx, older))
	time.Sleep(2 * time.Millisecond)
	newest := newToken(userID, utils.GenerateUUID(), time.Hour)
	require.NoError(t, repo.CreateOneTimeToken(ctx, newest))
	expired := newToken(userID, utils.GenerateUUID(), -time.Minute)
	expired.TokenType = auth.OneTimeTokenEmailChange
	require.NoError(t, repo.CreateOneTimeToken(ctx, expired))

	// Act
	got, err := repo.GetLatestActiveOneTimeToken(ctx, userID, auth.OneTimeTokenPasswordReset)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, newest.TokenHash, got.TokenHash)
	assert.NotEmpty(t, newest.ID, "create fills in the generated ID")
	assert.Equal(t, "203.0.113.7", got.IPAddress)
	assert.Equal(t, 0, got.Attempts)
	assert.False(t, got.IsUsed())

	expiredLookup, err := repo.GetLatestActiveOneTimeToken(ctx, userID, auth.OneTimeTokenEmailChange)
	require.NoError(t, err)
	assert.Nil(t, expiredLookup, "expired tokens are not returned")

	none, err := repo.GetLatestActiveOneTimeToken(ctx, userID, auth.OneTimeTokenEmailVerification)
	require.NoError(t, err)
	assert.Nil(t, none)
}

func TestOneTimeToken_ConsumeRejectsInvalidTokens(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)

	active := newToken(userID, utils.GenerateUUID(), time.Hour)
	require.NoError(t, repo.CreateOneTimeToken(ctx, active))
	expired := newToken(userID, utils.GenerateUUID(), -time.Minute)
	require.NoError(t, repo.CreateOneTimeToken(ctx, expired))

	t.Run("wrong type", func(t *testing.T) {
		err := repo.ConsumeOneTimeToken(ctx, active.TokenHash, auth.OneTimeTokenEmailChange)
		assert.ErrorIs(t, err, auth.ErrNotFound)
	})

	t.Run("unknown hash", func(t *testing.T) {
		err := repo.ConsumeOneTimeToken(ctx, "does-not-exist", auth.OneTimeTokenPasswordReset)
		assert.ErrorIs(t, err, auth.ErrNotFound)
	})

	t.Run("expired", func(t *testing.T) {
		err := repo.ConsumeOneTimeToken(ctx, expired.TokenHash, auth.OneTimeTokenPasswordReset)
		assert.ErrorIs(t, err, auth.ErrNotFound)
	})

	t.Run("success then reuse", func(t *testing.T) {
		require.NoError(t, repo.ConsumeOneTimeToken(ctx, active.TokenHash, auth.OneTimeTokenPasswordReset))
		assert.ErrorIs(t, repo.ConsumeOneTimeToken(ctx, active.TokenHash, auth.OneTimeTokenPasswordReset), auth.ErrNotFound)

		lookup, err := repo.GetLatestActiveOneTimeToken(ctx, userID, auth.OneTimeTokenPasswordReset)
		require.NoError(t, err)
		assert.Nil(t, lookup, "a consumed token is no longer active")
	})
}

func TestOneTimeToken_ConcurrentConsumeSucceedsOnce(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	token := newToken(userID, utils.GenerateUUID(), time.Hour)
	require.NoError(t, repo.CreateOneTimeToken(ctx, token))

	// Act: many requests consume the same token at once
	const callers = 20
	var wg sync.WaitGroup
	results := make(chan error, callers)
	start := make(chan struct{})
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results <- repo.ConsumeOneTimeToken(ctx, token.TokenHash, auth.OneTimeTokenPasswordReset)
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	// Assert
	successes := 0
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		assert.ErrorIs(t, err, auth.ErrNotFound)
	}
	assert.Equal(t, 1, successes, "exactly one caller consumes the token")

	var usedAt *time.Time
	require.NoError(t, db.Raw("SELECT used_at FROM one_time_tokens WHERE token_hash = ?", token.TokenHash).Scan(&usedAt).Error)
	assert.NotNil(t, usedAt)
}

func TestOneTimeToken_IncrementAttempts(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	token := newToken(userID, utils.GenerateUUID(), time.Hour)
	require.NoError(t, repo.CreateOneTimeToken(ctx, token))

	// Act
	first, err := repo.IncrementOneTimeTokenAttempts(ctx, token.ID)
	require.NoError(t, err)
	second, err := repo.IncrementOneTimeTokenAttempts(ctx, token.ID)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, 1, first)
	assert.Equal(t, 2, second)

	stored, err := repo.GetLatestActiveOneTimeToken(ctx, userID, auth.OneTimeTokenPasswordReset)
	require.NoError(t, err)
	assert.Equal(t, 2, stored.Attempts)

	require.NoError(t, repo.ConsumeOneTimeToken(ctx, token.TokenHash, auth.OneTimeTokenPasswordReset))
	_, err = repo.IncrementOneTimeTokenAttempts(ctx, token.ID)
	assert.ErrorIs(t, err, auth.ErrNotFound, "a consumed token no longer counts attempts")
}
