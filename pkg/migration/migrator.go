package migration

import (
	"fmt"
	"goilerplate/pkg/utils"
	"log"
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

// Migrator handles database migrations
type Migrator struct {
	db *gorm.DB
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// CreateMigrationsTable creates the migrations tracking table
func (m *Migrator) CreateMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS migrations (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

// splitSQLStatements splits SQL content into individual statements
// This function properly handles:
// - Comments (-- and /* */)
// - Multi-line statements
// - Semicolon separators
func (m *Migrator) splitSQLStatements(sql string) []string {
	var statements []string
	var currentStatement strings.Builder
	var inBlockComment bool

	lines := strings.Split(sql, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines
		if trimmedLine == "" {
			continue
		}

		// Handle line comments
		if strings.HasPrefix(trimmedLine, "--") {
			continue
		}

		// Handle block comments
		if strings.Contains(trimmedLine, "/*") {
			inBlockComment = true
		}
		if strings.Contains(trimmedLine, "*/") {
			inBlockComment = false
			continue
		}
		if inBlockComment {
			continue
		}

		// Add line to current statement
		currentStatement.WriteString(line)
		currentStatement.WriteString(" ")

		// Check if statement ends with semicolon
		if strings.HasSuffix(trimmedLine, ";") {
			stmt := strings.TrimSpace(currentStatement.String())
			// Remove trailing semicolon and clean up
			stmt = strings.TrimSuffix(stmt, ";")
			stmt = strings.TrimSpace(stmt)

			if stmt != "" {
				statements = append(statements, stmt)
			}
			currentStatement.Reset()
		}
	}

	// Handle any remaining statement without semicolon
	if currentStatement.Len() > 0 {
		stmt := strings.TrimSpace(currentStatement.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}

// executeSQL executes multiple SQL statements
func (m *Migrator) executeSQL(tx *gorm.DB, sql string) error {
	statements := m.splitSQLStatements(sql)

	for _, statement := range statements {
		if err := tx.Exec(statement).Error; err != nil {
			return fmt.Errorf("failed to execute statement: %s\nError: %w", statement, err)
		}
	}

	return nil
}

// Up runs pending migrations
func (m *Migrator) Up(migrationDir string) error {
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
			log.Printf("Migration %s already executed, skipping", migration.ID)
			continue
		}

		if migration.UpSQL == "" {
			log.Printf("No up migration found for %s, skipping", migration.ID)
			continue
		}

		log.Printf("Running migration %s: %s", migration.ID, migration.Name)

		// Execute migration in transaction
		err := m.db.Transaction(func(tx *gorm.DB) error {
			if err := m.executeSQL(tx, migration.UpSQL); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migration.ID, err)
			}

			if err := tx.Exec("INSERT INTO migrations (id, name) VALUES (?, ?)", migration.ID, migration.Name).Error; err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
			}

			return nil
		})

		if err != nil {
			return err
		}

		log.Printf("Migration %s completed successfully", migration.ID)
	}

	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down(migrationDir string) error {
	executed, err := m.GetExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	if len(executed) == 0 {
		log.Println("No migrations to rollback")
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

	log.Printf("Rolling back migration %s: %s", targetMigration.ID, targetMigration.Name)

	// Execute rollback in transaction
	err = m.db.Transaction(func(tx *gorm.DB) error {
		if err := m.executeSQL(tx, targetMigration.DownSQL); err != nil {
			return fmt.Errorf("failed to execute rollback %s: %w", targetMigration.ID, err)
		}

		if err := tx.Exec("DELETE FROM migrations WHERE id = ?", targetMigration.ID).Error; err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", targetMigration.ID, err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Printf("Migration %s rolled back successfully", targetMigration.ID)
	return nil
}

// Status shows migration status
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

	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}

	upFile := filepath.Join(migrationDir, fmt.Sprintf("%s_%s.up.sql", id, name))
	downFile := filepath.Join(migrationDir, fmt.Sprintf("%s_%s.down.sql", id, name))

	upContent := fmt.Sprintf("-- Migration: %s\n-- Created at: %s\n\n-- Add your up migration here\n", name, utils.Now().Format(time.RFC3339))
	downContent := fmt.Sprintf("-- Rollback: %s\n-- Created at: %s\n\n-- Add your down migration here\n", name, utils.Now().Format(time.RFC3339))

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	log.Printf("Created migration files:")
	log.Printf("  %s", upFile)
	log.Printf("  %s", downFile)

	return nil
}
