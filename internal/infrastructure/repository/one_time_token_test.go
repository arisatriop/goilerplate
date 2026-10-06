package repository_test

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"goilerplate/internal/domain/auth"
	"goilerplate/internal/domain/user"
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

	// One pool per test; without closing them they pile up until PostgreSQL refuses clients.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, migration.NewMigrator(db, nil).Up(context.Background(), migrationsDir))

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
	expired.NewEmail = "moved@example.test"
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
		_, err := repo.ConsumeOneTimeToken(ctx, active.TokenHash, auth.OneTimeTokenEmailChange)
		assert.ErrorIs(t, err, auth.ErrNotFound)
	})

	t.Run("unknown hash", func(t *testing.T) {
		_, err := repo.ConsumeOneTimeToken(ctx, "does-not-exist", auth.OneTimeTokenPasswordReset)
		assert.ErrorIs(t, err, auth.ErrNotFound)
	})

	t.Run("expired", func(t *testing.T) {
		_, err := repo.ConsumeOneTimeToken(ctx, expired.TokenHash, auth.OneTimeTokenPasswordReset)
		assert.ErrorIs(t, err, auth.ErrNotFound)
	})

	t.Run("success then reuse", func(t *testing.T) {
		consumed, err := repo.ConsumeOneTimeToken(ctx, active.TokenHash, auth.OneTimeTokenPasswordReset)
		require.NoError(t, err)
		assert.Equal(t, active.ID, consumed.ID)
		assert.Equal(t, userID, consumed.UserID, "the caller learns whose token it was")
		assert.True(t, consumed.IsUsed(), "the returned row is the updated one")
		assert.Equal(t, "203.0.113.7", consumed.IPAddress)

		_, err = repo.ConsumeOneTimeToken(ctx, active.TokenHash, auth.OneTimeTokenPasswordReset)
		assert.ErrorIs(t, err, auth.ErrNotFound)

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
			_, err := repo.ConsumeOneTimeToken(ctx, token.TokenHash, auth.OneTimeTokenPasswordReset)
			results <- err
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

	_, err = repo.ConsumeOneTimeToken(ctx, token.TokenHash, auth.OneTimeTokenPasswordReset)
	require.NoError(t, err)
	_, err = repo.IncrementOneTimeTokenAttempts(ctx, token.ID)
	assert.ErrorIs(t, err, auth.ErrNotFound, "a consumed token no longer counts attempts")
}

// Expiring ends the user's usable tokens of one type and nothing else: not another user's, not
// another type, and not the record that a token was actually redeemed.
func TestOneTimeToken_ExpireEndsOnlyThatUsersTokensOfThatType(t *testing.T) {
	// Arrange
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	otherUserID := createTestUser(t, db)

	first := newToken(userID, utils.GenerateUUID(), time.Hour)
	require.NoError(t, repo.CreateOneTimeToken(ctx, first))
	second := newToken(userID, utils.GenerateUUID(), time.Hour)
	require.NoError(t, repo.CreateOneTimeToken(ctx, second))
	redeemed := newToken(userID, utils.GenerateUUID(), time.Hour)
	require.NoError(t, repo.CreateOneTimeToken(ctx, redeemed))
	_, err := repo.ConsumeOneTimeToken(ctx, redeemed.TokenHash, auth.OneTimeTokenPasswordReset)
	require.NoError(t, err)

	otherType := newToken(userID, utils.GenerateUUID(), time.Hour)
	otherType.TokenType = auth.OneTimeTokenEmailVerification
	require.NoError(t, repo.CreateOneTimeToken(ctx, otherType))
	otherUser := newToken(otherUserID, utils.GenerateUUID(), time.Hour)
	require.NoError(t, repo.CreateOneTimeToken(ctx, otherUser))

	// Act
	require.NoError(t, repo.ExpireOneTimeTokens(ctx, userID, auth.OneTimeTokenPasswordReset))

	// Assert
	for _, token := range []*auth.OneTimeToken{first, second} {
		_, err := repo.ConsumeOneTimeToken(ctx, token.TokenHash, auth.OneTimeTokenPasswordReset)
		assert.ErrorIs(t, err, auth.ErrNotFound, "a superseded token can no longer be redeemed")

		var usedAt sql.NullTime
		require.NoError(t, db.Raw("SELECT used_at FROM one_time_tokens WHERE id = ?", token.ID).Row().Scan(&usedAt))
		assert.False(t, usedAt.Valid, "superseded is not the same as used")
	}

	var redeemedExpiry time.Time
	require.NoError(t, db.Raw("SELECT expires_at FROM one_time_tokens WHERE id = ?", redeemed.ID).Scan(&redeemedExpiry).Error)
	assert.WithinDuration(t, redeemed.ExpiresAt, redeemedExpiry, time.Millisecond, "a redeemed token keeps its own history")

	_, err = repo.ConsumeOneTimeToken(ctx, otherType.TokenHash, auth.OneTimeTokenEmailVerification)
	assert.NoError(t, err, "another token type is untouched")
	_, err = repo.ConsumeOneTimeToken(ctx, otherUser.TokenHash, auth.OneTimeTokenPasswordReset)
	assert.NoError(t, err, "another user's token is untouched")
}

// An OTP's hash is keyed by its token ID, so the ID the domain chose must be the one stored.
func TestOneTimeToken_CreateKeepsAGivenID(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)

	token := newToken(userID, utils.GenerateUUID(), time.Hour)
	token.ID = utils.GenerateUUID()
	chosen := token.ID
	require.NoError(t, repo.CreateOneTimeToken(ctx, token))

	stored, err := repo.GetLatestActiveOneTimeToken(ctx, userID, auth.OneTimeTokenPasswordReset)
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.Equal(t, chosen, stored.ID)
}

func TestMarkEmailVerified_KeepsTheFirstTimestamp(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)

	require.NoError(t, repo.MarkEmailVerified(ctx, userID))
	first, err := repo.GetUserByID(ctx, userID)
	require.NoError(t, err)
	require.True(t, first.EmailVerified)
	require.NotNil(t, first.EmailVerifiedAt)

	require.NoError(t, repo.MarkEmailVerified(ctx, userID), "marking twice is not an error")
	second, err := repo.GetUserByID(ctx, userID)
	require.NoError(t, err)
	assert.True(t, first.EmailVerifiedAt.Equal(*second.EmailVerifiedAt), "the original verification time is kept")

	assert.ErrorIs(t, repo.MarkEmailVerified(ctx, utils.GenerateUUID()), auth.ErrNotFound)
}

func TestOneTimeToken_NewEmailRoundTripsOnEmailChangeOnly(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)

	change := newToken(userID, utils.GenerateUUID(), time.Hour)
	change.TokenType = auth.OneTimeTokenEmailChange
	change.NewEmail = "moved@example.test"
	require.NoError(t, repo.CreateOneTimeToken(ctx, change))

	stored, err := repo.GetLatestActiveOneTimeToken(ctx, userID, auth.OneTimeTokenEmailChange)
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.Equal(t, "moved@example.test", stored.NewEmail)

	// The migration's CHECK: a change token needs a target, and no other token may carry one.
	missing := newToken(userID, utils.GenerateUUID(), time.Hour)
	missing.TokenType = auth.OneTimeTokenEmailChange
	assert.Error(t, repo.CreateOneTimeToken(ctx, missing), "an email change with nowhere to go")

	stray := newToken(userID, utils.GenerateUUID(), time.Hour)
	stray.NewEmail = "moved@example.test"
	assert.Error(t, repo.CreateOneTimeToken(ctx, stray), "a password reset carrying an address")
}

func TestUpdateUserEmail(t *testing.T) {
	repo, db := newTestRepository(t)
	ctx := context.Background()
	userID := createTestUser(t, db)
	otherID := createTestUser(t, db)
	newEmail := utils.GenerateUUID() + "@example.test"

	require.NoError(t, repo.UpdateUserEmail(ctx, userID, newEmail))

	moved, err := repo.GetUserByID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, newEmail, moved.Email)
	assert.True(t, moved.EmailVerified, "the new address was proved by the code sent to it")
	assert.NotNil(t, moved.EmailVerifiedAt)

	err = repo.UpdateUserEmail(ctx, otherID, newEmail)
	assert.ErrorIs(t, err, auth.ErrEmailAlreadyRegistered, "the unique constraint, not a 500")

	assert.ErrorIs(t, repo.UpdateUserEmail(ctx, utils.GenerateUUID(), "nobody@example.test"), auth.ErrNotFound)
}

// Two registrations for one address can both pass the lookup; the constraint decides, and the
// loser must get the domain's error so registration can answer it like any taken address.
func TestUserCreateUser_DuplicateEmailIsTheDomainError(t *testing.T) {
	_, db := newTestRepository(t)
	users := repository.NewUser(db)
	ctx := context.Background()
	email := utils.GenerateUUID() + "@example.test"

	created, err := users.CreateUser(ctx, &user.User{Name: "First", Email: email, PasswordHash: "x", IsActive: true})
	require.NoError(t, err)
	t.Cleanup(func() { db.Exec("DELETE FROM users WHERE id = ?", created.ID.String()) })

	_, err = users.CreateUser(ctx, &user.User{Name: "Second", Email: email, PasswordHash: "x", IsActive: true})

	assert.ErrorIs(t, err, user.ErrEmailAlreadyRegistered)
}
