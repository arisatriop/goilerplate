package migration

import (
	"context"
	"database/sql"
	"fmt"
	"goilerplate/pkg/utils"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Migration represents a single migration
type Migration struct {
	ID        string
	Name      string
	UpSQL     string
	DownSQL   string
	Timestamp time.Time
}

// advisoryLockKey names the lock every migration run takes before touching the schema. The
// value is arbitrary, but must be the same in every process.
const advisoryLockKey int64 = 7_263_845_190_113_002

// Migrator handles database migrations
type Migrator struct {
	db  *gorm.DB
	log *slog.Logger
}

// withAdvisoryLock runs fn while holding a PostgreSQL advisory lock, so two processes cannot
// migrate the same database at once.
//
// Without it, concurrent runs collide before the migration bookkeeping even starts: CREATE
// TABLE IF NOT EXISTS is not concurrency-safe in PostgreSQL, and two sessions creating the
// migrations table together fail with a duplicate key on pg_type_typname_nsp_index. Runs that
// got past that would each read the same pending list and apply the same migration twice.
//
// This is reached in ordinary use rather than in theory: `go test ./...` runs packages in
// parallel, and two test packages sharing one database both migrate it as they start.
//
// The lock is held on one dedicated connection, because an advisory lock belongs to the session
// that took it. That connection goes back to the pool afterwards, so the unlock has to be
// explicit rather than left to the session ending.
func (m *Migrator) withAdvisoryLock(ctx context.Context, fn func() error) error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("failed to access the connection pool: %w", err)
	}

	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to reserve a connection for the migration lock: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", advisoryLockKey); err != nil {
		return fmt.Errorf("failed to acquire the migration lock: %w", err)
	}
	defer func() {
		// Logged rather than returned: whatever fn did is the caller's answer, and the lock is
		// dropped when the connection's session ends regardless.
		if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", advisoryLockKey); err != nil {
			m.log.Error("failed to release the migration lock", "error", err)
		}
	}()

	return fn()
}

// NewMigrator creates a new migrator instance. It logs through slog like the rest of the
// process; it used to use stdlib log, which in a container interleaves unstructured lines with
// the JSON stream and breaks whatever is parsing it.
//
// A nil logger falls back to slog's default, so a test or a script can call this with one
// argument.
func NewMigrator(db *gorm.DB, log *slog.Logger) *Migrator {
	if log == nil {
		log = slog.Default()
	}
	return &Migrator{db: db, log: log}
}

// CreateMigrationsTable creates the migrations tracking table
func (m *Migrator) CreateMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS migrations (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		executed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	)`

	return m.db.Exec(query).Error
}

// GetExecutedMigrations returns list of executed migrations
func (m *Migrator) GetExecutedMigrations() ([]string, error) {
	var migrations []string

	err := m.db.Raw("SELECT id FROM migrations ORDER BY id").Scan(&migrations).Error
	if err != nil {
		return nil, err
	}

	return migrations, nil
}

// LoadMigrations loads migration files from directory
func (m *Migrator) LoadMigrations(migrationDir string) ([]Migration, error) {
	var migrations []Migration

	err := filepath.Walk(migrationDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(info.Name(), ".sql") {
			return nil
		}

		// Parse filename: 001_create_users_table.up.sql or 001_create_users_table.down.sql
		parts := strings.Split(info.Name(), "_")
		if len(parts) < 2 {
			return nil
		}

		id := parts[0]
		namePart := strings.Join(parts[1:], "_")

		var isUp bool
		var name string

		if strings.HasSuffix(namePart, ".up.sql") {
			isUp = true
			name = strings.TrimSuffix(namePart, ".up.sql")
		} else if strings.HasSuffix(namePart, ".down.sql") {
			isUp = false
			name = strings.TrimSuffix(namePart, ".down.sql")
		} else {
			return nil
		}

		// #nosec G304 G122 -- path comes from walking migrationDir, a developer-controlled directory
		// of source files, not from request data. Symlink TOCTOU there is not a threat model.
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Find existing migration or create new one
		var migration *Migration
		for i := range migrations {
			if migrations[i].ID == id && migrations[i].Name == name {
				migration = &migrations[i]
				break
			}
		}

		if migration == nil {
			timestamp, _ := time.Parse("20060102150405", id)
			migrations = append(migrations, Migration{
				ID:        id,
				Name:      name,
				Timestamp: timestamp,
			})
			migration = &migrations[len(migrations)-1]
		}

		if isUp {
			migration.UpSQL = string(content)
		} else {
			migration.DownSQL = string(content)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// runScript executes one migration file inside tx.
//
// The whole file goes to PostgreSQL in a single Exec, deliberately. pgx forces the simple
// query protocol whenever a query carries no arguments (conn.go: "Always use simple protocol
// when there are no arguments"), and the simple protocol accepts several statements in one
// message — so the server parses the SQL, which is the only thing that can parse SQL correctly.
//
// This replaces a line-based splitter that cut the file on any line ending in a semicolon.
// That splitter was wrong in two ways that matter, both proved by tests in
// sql_execution_test.go:
//
//   - A dollar-quoted body — which is every PL/pgSQL function, including the updated_at
//     trigger that most schemas have — was torn into fragments at the semicolons inside it.
//     That fails loudly: "unterminated dollar-quoted string".
//   - A line carrying an inline block comment was dropped whole. `id int PRIMARY KEY, /* ... */`
//     simply disappeared, leaving valid SQL that creates the table without its primary key.
//     Nothing errors. The migration is recorded as applied. That is a half-applied schema that
//     reports success, which is the failure mode this package exists to avoid.
//
// A caller gets one error for the file rather than one per statement, but PostgreSQL reports
// the position of the failure, which is better than the statement text the splitter produced.
func runScript(ctx context.Context, tx *sql.Tx, script string) error {
	if _, err := tx.ExecContext(ctx, script); err != nil {
		return fmt.Errorf("executing migration SQL: %w", err)
	}
	return nil
}

// begin starts a transaction on the pool underneath GORM.
//
// The raw *sql.DB rather than m.db.Transaction, because GORM is configured with
// PrepareStmt: true, and a prepared statement goes over the extended protocol, which accepts
// exactly one statement per message. Going through database/sql directly keeps the no-argument
// path that pgx maps to the simple protocol.
func (m *Migrator) begin(ctx context.Context) (*sql.Tx, error) {
	sqlDB, err := m.db.DB()
	if err != nil {
		return nil, fmt.Errorf("accessing the connection pool: %w", err)
	}

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	return tx, nil
}

// applyInTx runs fn in a transaction, rolling back on any error or panic.
func (m *Migrator) applyInTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := m.begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// A no-op once the transaction has been committed; it matters on the error and panic
		// paths, where leaving it open would hold the schema locks it took.
		_ = tx.Rollback()
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing migration: %w", err)
	}
	return nil
}

// Up runs pending migrations. Cancelling ctx aborts the run: the statement in flight is
// cancelled server-side, its transaction rolls back, and the advisory lock is released, so a
// deploy that times out does not leave the next one locked out.
func (m *Migrator) Up(ctx context.Context, migrationDir string) error {
	return m.withAdvisoryLock(ctx, func() error { return m.up(ctx, migrationDir) })
}

func (m *Migrator) up(ctx context.Context, migrationDir string) error {
	if err := m.CreateMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	executed, err := m.GetExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	migrations, err := m.LoadMigrations(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	executedMap := make(map[string]bool)
	for _, id := range executed {
		executedMap[id] = true
	}

	for _, migration := range migrations {
		if executedMap[migration.ID] {
			m.log.Debug("migration already applied, skipping", "id", migration.ID)
			continue
		}

		if migration.UpSQL == "" {
			m.log.Warn("no up migration found, skipping", "id", migration.ID)
			continue
		}

		m.log.Info("applying migration", "id", migration.ID, "name", migration.Name)

		// The SQL and the bookkeeping row commit together, so a failure halfway through leaves
		// neither the schema change nor a record claiming it was applied.
		if err := m.applyInTx(ctx, func(tx *sql.Tx) error {
			if err := runScript(ctx, tx, migration.UpSQL); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migration.ID, err)
			}

			if _, err := tx.ExecContext(ctx, "INSERT INTO migrations (id, name) VALUES ($1, $2)", migration.ID, migration.Name); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
			}

			return nil
		}); err != nil {
			return err
		}

		m.log.Info("migration applied", "id", migration.ID)
	}

	return nil
}

// Down rolls back the last applied migration.
func (m *Migrator) Down(ctx context.Context, migrationDir string) error {
	return m.withAdvisoryLock(ctx, func() error { return m.down(ctx, migrationDir) })
}

func (m *Migrator) down(ctx context.Context, migrationDir string) error {
	executed, err := m.GetExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	if len(executed) == 0 {
		m.log.Info("no migrations to roll back")
		return nil
	}

	lastMigrationID := executed[len(executed)-1]

	migrations, err := m.LoadMigrations(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	var targetMigration *Migration
	for _, migration := range migrations {
		if migration.ID == lastMigrationID {
			targetMigration = &migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration file for %s not found", lastMigrationID)
	}

	if targetMigration.DownSQL == "" {
		return fmt.Errorf("no down migration found for %s", lastMigrationID)
	}

	m.log.Info("rolling back migration", "id", targetMigration.ID, "name", targetMigration.Name)

	if err := m.applyInTx(ctx, func(tx *sql.Tx) error {
		if err := runScript(ctx, tx, targetMigration.DownSQL); err != nil {
			return fmt.Errorf("failed to execute rollback %s: %w", targetMigration.ID, err)
		}

		if _, err := tx.ExecContext(ctx, "DELETE FROM migrations WHERE id = $1", targetMigration.ID); err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", targetMigration.ID, err)
		}

		return nil
	}); err != nil {
		return err
	}

	m.log.Info("migration rolled back", "id", targetMigration.ID)
	return nil
}

// Status prints which migrations are applied and which are pending.
func (m *Migrator) Status(migrationDir string) error {
	executed, err := m.GetExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	migrations, err := m.LoadMigrations(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	executedMap := make(map[string]bool)
	for _, id := range executed {
		executedMap[id] = true
	}

	// stdout, for the same reason as CreateMigrationFiles: a table for a human, not a log event.
	fmt.Println("Migration Status:")
	fmt.Println("ID\t\tName\t\t\t\tStatus")
	fmt.Println("--\t\t----\t\t\t\t------")

	for _, migration := range migrations {
		status := "Pending"
		if executedMap[migration.ID] {
			status = "Applied"
		}
		fmt.Printf("%s\t\t%s\t\t\t%s\n", migration.ID, migration.Name, status)
	}

	return nil
}

// GenerateMigrationID generates a new migration ID based on timestamp
func GenerateMigrationID() string {
	return utils.Now().Format("20060102150405")
}

// CreateMigrationFiles creates up and down migration files
func CreateMigrationFiles(migrationDir, name string) error {
	id := GenerateMigrationID()

	if err := os.MkdirAll(migrationDir, 0750); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}

	upFile := filepath.Join(migrationDir, fmt.Sprintf("%s_%s.up.sql", id, name))
	downFile := filepath.Join(migrationDir, fmt.Sprintf("%s_%s.down.sql", id, name))

	upContent := fmt.Sprintf("-- Migration: %s\n-- Created at: %s\n\n-- Add your up migration here\n", name, utils.Now().Format(time.RFC3339))
	downContent := fmt.Sprintf("-- Rollback: %s\n-- Created at: %s\n\n-- Add your down migration here\n", name, utils.Now().Format(time.RFC3339))

	if err := os.WriteFile(upFile, []byte(upContent), 0600); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	if err := os.WriteFile(downFile, []byte(downContent), 0600); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	// stdout rather than the logger: this is a CLI result a developer reads at a terminal and
	// may pipe into an editor, not an operational event a log pipeline should index.
	fmt.Println("Created migration files:")
	fmt.Printf("  %s\n", upFile)
	fmt.Printf("  %s\n", downFile)

	return nil
}
