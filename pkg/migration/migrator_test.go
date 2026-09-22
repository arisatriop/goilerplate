package migration_test

import (
	"context"
	"os"
	"sync"
	"testing"

	"goilerplate/pkg/migration"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const migrationsDir = "../../internal/migrations"

func openDB(t *testing.T) *gorm.DB {
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

	return db
}

// Two processes migrating the same database at once has to be safe, because it happens without
// anyone arranging it: `go test ./...` runs packages in parallel and every package that needs a
// schema migrates on start.
//
// Before the advisory lock this failed on the migrations table itself — CREATE TABLE IF NOT
// EXISTS is not concurrency-safe in PostgreSQL, so the losers got
// "duplicate key value violates unique constraint pg_type_typname_nsp_index". A run that got
// past that would have applied the same pending migration more than once.
func TestUp_ConcurrentRunsDoNotCollide(t *testing.T) {
	const runners = 4

	errs := make([]error, runners)
	var wg sync.WaitGroup

	for i := range runners {
		db := openDB(t)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = migration.NewMigrator(db, nil).Up(context.Background(), migrationsDir)
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "migrator %d failed", i)
	}
}

// Migrating a database that is already up to date must be a no-op rather than an error, which
// is what lets every test package call Up on start.
func TestUp_IsIdempotent(t *testing.T) {
	migrator := migration.NewMigrator(openDB(t), nil)

	require.NoError(t, migrator.Up(context.Background(), migrationsDir))
	require.NoError(t, migrator.Up(context.Background(), migrationsDir))
}

// The lock must be released, or the next run would block on it forever. Up returning at all on
// the second call is the evidence; a leaked lock would hang this test instead.
func TestUp_ReleasesTheLock(t *testing.T) {
	db := openDB(t)

	require.NoError(t, migration.NewMigrator(db, nil).Up(context.Background(), migrationsDir))

	var locks int64
	require.NoError(t, db.Raw(
		"SELECT count(*) FROM pg_locks WHERE locktype = 'advisory'").Scan(&locks).Error)
	require.Zero(t, locks, "the migration lock outlived the run")
}
