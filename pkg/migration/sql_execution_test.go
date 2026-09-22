package migration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"goilerplate/pkg/migration"
	"goilerplate/pkg/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// writeMigration puts a single up/down pair in a fresh directory, so a test can run exactly the
// SQL it cares about rather than the whole project schema.
func writeMigration(t *testing.T, id, name, up, down string) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", id, name)), []byte(up), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", id, name)), []byte(down), 0o600))
	return dir
}

// isolatedDB gives each test its own schema, so tests can run in any order without seeing each
// other's tables.
//
// The schema is selected with pgx's search_path runtime parameter rather than a SET statement.
// SET applies to the one connection that ran it and GORM hands out connections from a pool, so
// a SET here would leave the migrator running against public on whichever other connection it
// drew.
//
// The connection is built through pgx.ParseConfig rather than by editing the DSN string,
// because POSTGRES_TEST_DSN comes in either shape: a URL locally, and libpq key/value in CI
// ("host=localhost port=5432 user=postgres ..."). An earlier version of this helper assumed a
// URL and used url.Parse, which does not fail on the key/value form — it quietly produces a
// path-shaped URL — so the mangled DSN only showed up as a DNS lookup for a hostname containing
// the entire connection string. pgx.ParseConfig understands both.
func isolatedDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN not set; skipping PostgreSQL integration test")
	}

	schema := fmt.Sprintf("mig_test_%d", hashName(t.Name()))

	admin := openDB(t)
	require.NoError(t, admin.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema)).Error)
	require.NoError(t, admin.Exec(fmt.Sprintf("CREATE SCHEMA %s", schema)).Error)
	t.Cleanup(func() {
		_ = admin.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema)).Error
	})

	config, err := pgx.ParseConfig(dsn)
	require.NoError(t, err, "POSTGRES_TEST_DSN must be a URL or a libpq key/value DSN")
	if config.RuntimeParams == nil {
		config.RuntimeParams = map[string]string{}
	}
	config.RuntimeParams["search_path"] = schema

	sqlDB := stdlib.OpenDB(*config)
	t.Cleanup(func() { _ = sqlDB.Close() })

	scoped, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		NowFunc: utils.Now,
		Logger:  gormlogger.Discard,
	})
	require.NoError(t, err)

	// Prove the scoping took before any test relies on it. Without this, a helper that silently
	// fell back to public would make every assertion below meaningless.
	var current string
	require.NoError(t, scoped.Raw("SELECT current_schema()").Scan(&current).Error)
	require.Equal(t, schema, current, "the test must be scoped to its own schema")

	return scoped
}

func hashName(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h % 100000
}

// A trigger that maintains updated_at is the most common thing a PostgreSQL migration does
// beyond CREATE TABLE, and its body is dollar-quoted. The body contains semicolons, so any
// splitter that treats a semicolon as a statement boundary tears the function apart.
func TestUp_DollarQuotedFunctionBody(t *testing.T) {
	// Arrange
	db := isolatedDB(t)
	dir := writeMigration(t, "20260101000000", "add_touch_function", `
CREATE OR REPLACE FUNCTION touch_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
`, `DROP FUNCTION IF EXISTS touch_updated_at();`)

	// Act
	err := migration.NewMigrator(db, nil).Up(context.Background(), dir)

	// Assert
	require.NoError(t, err)

	var exists bool
	require.NoError(t, db.Raw(
		`SELECT EXISTS (
		     SELECT 1 FROM pg_proc p
		     JOIN pg_namespace n ON n.oid = p.pronamespace
		     WHERE p.proname = 'touch_updated_at' AND n.nspname = current_schema()
		 )`).Scan(&exists).Error)
	assert.True(t, exists, "the function must actually exist after the migration reports success")
}

// A statement carrying an inline block comment must still be executed. The dangerous version of
// getting this wrong is not an error — it is the statement silently vanishing while the
// migration is recorded as applied, which leaves a schema that is missing a table nobody
// noticed was never created.
func TestUp_InlineBlockComment(t *testing.T) {
	// Arrange
	db := isolatedDB(t)
	dir := writeMigration(t, "20260101000001", "add_widgets", `
CREATE TABLE widgets (
    id   int PRIMARY KEY, /* the surrogate key */
    name text NOT NULL
);
`, `DROP TABLE IF EXISTS widgets;`)

	// Act
	err := migration.NewMigrator(db, nil).Up(context.Background(), dir)

	// Assert
	require.NoError(t, err)

	var exists bool
	require.NoError(t, db.Raw(
		`SELECT EXISTS (
		     SELECT 1 FROM information_schema.tables
		     WHERE table_name = 'widgets' AND table_schema = current_schema()
		 )`).Scan(&exists).Error)
	require.True(t, exists, "a migration reported as applied must have actually created the table")

	// The table existing is not enough. Dropping the whole line that carries the comment still
	// produces valid SQL — it just produces a table with a column missing, which no error
	// reports and no test that only checks for the table would catch.
	var columns []string
	require.NoError(t, db.Raw(
		`SELECT column_name FROM information_schema.columns
		 WHERE table_name = 'widgets' AND table_schema = current_schema()
		 ORDER BY ordinal_position`).Scan(&columns).Error)
	assert.Equal(t, []string{"id", "name"}, columns,
		"every column must survive, including the one on the line carrying the comment")
}

// A semicolon inside a string literal is data, not a statement boundary.
func TestUp_SemicolonInsideAStringLiteral(t *testing.T) {
	// Arrange
	db := isolatedDB(t)
	dir := writeMigration(t, "20260101000002", "seed_settings", `
CREATE TABLE settings (key text PRIMARY KEY, value text NOT NULL);
INSERT INTO settings (key, value) VALUES ('delimiters', 'a;b;c');
`, `DROP TABLE IF EXISTS settings;`)

	// Act
	err := migration.NewMigrator(db, nil).Up(context.Background(), dir)

	// Assert
	require.NoError(t, err)

	var value string
	require.NoError(t, db.Raw("SELECT value FROM settings WHERE key = 'delimiters'").Scan(&value).Error)
	assert.Equal(t, "a;b;c", value, "the literal must survive intact")
}

// A comment marker inside a string literal is data too.
func TestUp_CommentMarkerInsideAStringLiteral(t *testing.T) {
	// Arrange
	db := isolatedDB(t)
	dir := writeMigration(t, "20260101000003", "seed_notes", `
CREATE TABLE notes (id int PRIMARY KEY, body text NOT NULL);
INSERT INTO notes (id, body) VALUES (1, 'see -- the dashes');
`, `DROP TABLE IF EXISTS notes;`)

	// Act
	require.NoError(t, migration.NewMigrator(db, nil).Up(context.Background(), dir))

	// Assert
	var body string
	require.NoError(t, db.Raw("SELECT body FROM notes WHERE id = 1").Scan(&body).Error)
	assert.Equal(t, "see -- the dashes", body)
}

// ── Failure and rollback ─────────────────────────────────────────────────────

// A migration that fails partway must leave nothing behind: not the half of the schema that
// succeeded, and not a row claiming it was applied. This is what makes a failed run safe to
// retry after fixing the SQL, and it is why this migrator has no "dirty" state to force —
// the transaction is the whole mechanism.
func TestUp_PartialFailureRollsBackEverything(t *testing.T) {
	// Arrange: the first statement is valid, the second is not.
	db := isolatedDB(t)
	dir := writeMigration(t, "20260101000004", "half_broken", `
CREATE TABLE good (id int PRIMARY KEY);
CREATE TABLE bad (id int PRIMARY KEY, oops NOT_A_REAL_TYPE);
`, `DROP TABLE IF EXISTS good;`)

	// Act
	err := migration.NewMigrator(db, nil).Up(context.Background(), dir)

	// Assert
	require.Error(t, err, "an invalid statement must fail the run")

	var tables []string
	require.NoError(t, db.Raw(
		`SELECT table_name FROM information_schema.tables WHERE table_schema = current_schema()`).
		Scan(&tables).Error)
	assert.NotContains(t, tables, "good",
		"the statement that succeeded must be rolled back with the one that failed")

	var recorded int64
	require.NoError(t, db.Raw("SELECT count(*) FROM migrations WHERE id = '20260101000004'").Scan(&recorded).Error)
	assert.Zero(t, recorded, "a migration that failed must not be recorded as applied")
}

// The retry after fixing the SQL has to work, which is the point of the rollback above.
func TestUp_SucceedsOnRetryAfterAFailure(t *testing.T) {
	// Arrange
	db := isolatedDB(t)
	dir := writeMigration(t, "20260101000005", "fixable", `
CREATE TABLE fixable (id int PRIMARY KEY, oops NOT_A_REAL_TYPE);
`, `DROP TABLE IF EXISTS fixable;`)

	migrator := migration.NewMigrator(db, nil)
	require.Error(t, migrator.Up(context.Background(), dir))

	// Act: the operator fixes the file and runs again.
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "20260101000005_fixable.up.sql"),
		[]byte("CREATE TABLE fixable (id int PRIMARY KEY, ok text);"), 0o600))

	// Assert
	require.NoError(t, migrator.Up(context.Background(), dir))

	var exists bool
	require.NoError(t, db.Raw(
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables
		 WHERE table_name = 'fixable' AND table_schema = current_schema())`).Scan(&exists).Error)
	assert.True(t, exists)
}

// ── Down ─────────────────────────────────────────────────────────────────────

func TestDown_RollsBackTheLastMigrationOnly(t *testing.T) {
	// Arrange: two migrations, applied in order.
	db := isolatedDB(t)
	dir := t.TempDir()
	write := func(id, name, up, down string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", id, name)), []byte(up), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", id, name)), []byte(down), 0o600))
	}
	write("20260101000010", "first", "CREATE TABLE first (id int PRIMARY KEY);", "DROP TABLE first;")
	write("20260101000011", "second", "CREATE TABLE second (id int PRIMARY KEY);", "DROP TABLE second;")

	migrator := migration.NewMigrator(db, nil)
	require.NoError(t, migrator.Up(context.Background(), dir))

	// Act
	require.NoError(t, migrator.Down(context.Background(), dir))

	// Assert
	var tables []string
	require.NoError(t, db.Raw(
		`SELECT table_name FROM information_schema.tables WHERE table_schema = current_schema()`).
		Scan(&tables).Error)
	assert.Contains(t, tables, "first", "only the most recent migration rolls back")
	assert.NotContains(t, tables, "second")

	var ids []string
	require.NoError(t, db.Raw("SELECT id FROM migrations ORDER BY id").Scan(&ids).Error)
	assert.Equal(t, []string{"20260101000010"}, ids)
}

// Down on an empty database is a no-op, not an error: it is what a rollback script hits when
// it runs twice.
func TestDown_WithNothingAppliedIsANoOp(t *testing.T) {
	db := isolatedDB(t)
	dir := writeMigration(t, "20260101000012", "unused", "CREATE TABLE unused (id int);", "DROP TABLE unused;")

	migrator := migration.NewMigrator(db, nil)
	require.NoError(t, migrator.CreateMigrationsTable())

	assert.NoError(t, migrator.Down(context.Background(), dir))
}

// A failing down migration must leave the record in place, so the schema and the bookkeeping
// still agree and the rollback can be retried.
func TestDown_FailureKeepsTheMigrationRecorded(t *testing.T) {
	// Arrange
	db := isolatedDB(t)
	dir := writeMigration(t, "20260101000013", "bad_down",
		"CREATE TABLE keeper (id int PRIMARY KEY);",
		"DROP TABLE keeper; DROP TABLE does_not_exist;")

	migrator := migration.NewMigrator(db, nil)
	require.NoError(t, migrator.Up(context.Background(), dir))

	// Act
	err := migrator.Down(context.Background(), dir)

	// Assert
	require.Error(t, err)

	var recorded int64
	require.NoError(t, db.Raw("SELECT count(*) FROM migrations WHERE id = '20260101000013'").Scan(&recorded).Error)
	assert.Equal(t, int64(1), recorded, "a failed rollback must not un-record the migration")

	var exists bool
	require.NoError(t, db.Raw(
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables
		 WHERE table_name = 'keeper' AND table_schema = current_schema())`).Scan(&exists).Error)
	assert.True(t, exists, "the dropped table must come back with the rollback of the rollback")
}

// ── Ordering ─────────────────────────────────────────────────────────────────

// Migrations apply in ID order, not in whatever order the directory is read, and a failure
// stops the run rather than skipping ahead.
func TestUp_AppliesInOrderAndStopsAtAFailure(t *testing.T) {
	// Arrange
	db := isolatedDB(t)
	dir := t.TempDir()
	write := func(id, name, up string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", id, name)), []byte(up), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", id, name)), []byte("SELECT 1;"), 0o600))
	}
	write("20260101000020", "a", "CREATE TABLE step_a (id int);")
	write("20260101000021", "b", "CREATE TABLE step_b (id int, bad NOT_A_TYPE);")
	write("20260101000022", "c", "CREATE TABLE step_c (id int);")

	// Act
	err := migration.NewMigrator(db, nil).Up(context.Background(), dir)

	// Assert
	require.Error(t, err)

	var ids []string
	require.NoError(t, db.Raw("SELECT id FROM migrations ORDER BY id").Scan(&ids).Error)
	assert.Equal(t, []string{"20260101000020"}, ids,
		"a runs, b fails, and c must not be reached — applying it would skip over the gap b left")
}
