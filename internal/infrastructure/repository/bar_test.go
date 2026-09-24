package repository_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"goilerplate/internal/domain/bar"
	"goilerplate/internal/infrastructure/repository"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newBarRepository returns a bar repository on the test database and a context carrying the
// audit identity the repository stamps on every write.
func newBarRepository(t *testing.T) (bar.Repository, *gorm.DB, context.Context) {
	t.Helper()

	_, db := newTestRepository(t)
	ctx := context.WithValue(context.Background(), constants.ContextKeyUserID, "bar-repository-test")
	return repository.NewBar(db), db, ctx
}

// uniqueCode returns a code no other test uses, removing every bar that carries it afterwards.
func uniqueCode(t *testing.T, db *gorm.DB) string {
	t.Helper()

	code := "EXP-" + utils.GenerateUUID()
	t.Cleanup(func() { db.Exec("DELETE FROM bars WHERE code = ?", code) })
	return code
}

func TestBar_CreateDuplicateLiveCodeIsAConflict(t *testing.T) {
	// Arrange
	repo, db, ctx := newBarRepository(t)
	code := uniqueCode(t, db)
	_, err := repo.CreateBar(ctx, &bar.Bar{Code: code, Bar: "first"})
	require.NoError(t, err)

	// Act
	_, err = repo.CreateBar(ctx, &bar.Bar{Code: code, Bar: "second"})

	// Assert
	assert.ErrorIs(t, err, bar.ErrCodeAlreadyExists, "a duplicate must be a 409, not a 500")
}

// code is the business key that identifies a bar across environments, so it is never
// reassigned: deleting a bar does not free its code for another one.
func TestBar_DeletedBarsCodeIsNeverReassigned(t *testing.T) {
	// Arrange
	repo, db, ctx := newBarRepository(t)
	code := uniqueCode(t, db)
	first, err := repo.CreateBar(ctx, &bar.Bar{Code: code, Bar: "first"})
	require.NoError(t, err)
	require.NoError(t, repo.DeleteBar(ctx, first))

	// Act
	_, err = repo.CreateBar(ctx, &bar.Bar{Code: code, Bar: "second"})

	// Assert
	assert.ErrorIs(t, err, bar.ErrCodeAlreadyExists)
}

// The use case refuses a code change, and the repository does not write the column either, so
// no other caller can change a business key by accident.
func TestBar_UpdateNeverChangesTheCode(t *testing.T) {
	// Arrange
	repo, db, ctx := newBarRepository(t)
	code := uniqueCode(t, db)
	created, err := repo.CreateBar(ctx, &bar.Bar{Code: code, Bar: "before"})
	require.NoError(t, err)

	// Act
	err = repo.UpdateBar(ctx, &bar.Bar{ID: created.ID, Code: uniqueCode(t, db), Bar: "after"})

	// Assert
	require.NoError(t, err)
	reloaded, err := repo.GetBarByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, code, reloaded.Code)
	assert.Equal(t, "after", reloaded.Bar)
}

// The race a check-then-insert cannot close: every request passes the check, and only the index
// decides. Exactly one may win, and every loser must get the conflict error, not a 500.
func TestBar_ConcurrentCreatesOfOneCodeYieldOneWinner(t *testing.T) {
	// Arrange
	repo, db, ctx := newBarRepository(t)
	code := uniqueCode(t, db)
	const attempts = 8

	// Act
	var wg sync.WaitGroup
	errs := make([]error, attempts)
	start := make(chan struct{})
	for i := range attempts {
		wg.Go(func() {
			<-start
			_, errs[i] = repo.CreateBar(ctx, &bar.Bar{Code: code, Bar: "racer"})
		})
	}
	close(start)
	wg.Wait()

	// Assert
	winners := 0
	for _, err := range errs {
		switch {
		case err == nil:
			winners++
		case !errors.Is(err, bar.ErrCodeAlreadyExists):
			t.Errorf("a losing create failed with %v, want ErrCodeAlreadyExists", err)
		}
	}
	assert.Equal(t, 1, winners)
}

func TestBar_BulkCreateWithAnExistingCodeWritesNothing(t *testing.T) {
	// Arrange
	repo, db, ctx := newBarRepository(t)
	taken := uniqueCode(t, db)
	fresh := uniqueCode(t, db)
	_, err := repo.CreateBar(ctx, &bar.Bar{Code: taken, Bar: "holder"})
	require.NoError(t, err)

	// Act
	err = repo.BulkCreate(ctx, []*bar.Bar{{Code: fresh, Bar: "new"}, {Code: taken, Bar: "clash"}})

	// Assert
	assert.ErrorIs(t, err, bar.ErrCodeAlreadyExists)
	var count int64
	require.NoError(t, db.Table("bars").Where("code = ?", fresh).Count(&count).Error)
	assert.Zero(t, count, "the batch is one statement, so nothing from it may be written")
}

func TestBar_GetBarByIDUnknownIsNotFound(t *testing.T) {
	repo, _, ctx := newBarRepository(t)

	_, err := repo.GetBarByID(ctx, utils.GenerateUUID())

	assert.ErrorIs(t, err, bar.ErrNotFound)
}
