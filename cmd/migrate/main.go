package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"goilerplate/internal/bootstrap"
	"goilerplate/pkg/migration"
)

func main() {
	// All times are UTC regardless of the host time zone
	time.Local = time.UTC

	var (
		action        = flag.String("action", "", "Migration action: up, down, status, create")
		migrationName = flag.String("name", "", "Migration name (required for the create action)")
		migrationDir  = flag.String("dir", "internal/migrations", "Migration directory")
	)
	flag.Usage = printUsage
	flag.Parse()

	if *action == "" {
		printUsage()
		os.Exit(2)
	}

	if *action == "create" {
		createMigration(*migrationDir, *migrationName)
		return
	}

	// Every other action needs the database, so the config is loaded and validated first.
	app := bootstrap.Init()
	log := app.Log

	if app.DB == nil || app.DB.GDB == nil {
		log.Error("database connection not available")
		os.Exit(1)
	}

	// This binary usually runs as an init container, so its output is scraped by the same log
	// pipeline as the server's. That is why the operational lines go through slog rather than
	// fmt — an unstructured line in the middle of a JSON stream breaks whatever parses it.
	migrator := migration.NewMigrator(app.DB.GDB, log)

	switch *action {
	case "up":
		if err := migrator.Up(*migrationDir); err != nil {
			log.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
		log.Info("migrations completed")

	case "down":
		if err := migrator.Down(*migrationDir); err != nil {
			log.Error("failed to roll back migration", "error", err)
			os.Exit(1)
		}
		log.Info("migration rolled back")

	case "status":
		// Status prints a table for a human; the failure is still an operational event.
		if err := migrator.Status(*migrationDir); err != nil {
			log.Error("failed to read migration status", "error", err)
			os.Exit(1)
		}

	default:
		log.Error("unknown action", "action", *action)
		printUsage()
		os.Exit(2)
	}
}

// createMigration writes the migration pair. It runs before bootstrap.Init, so there is no
// application logger yet — and it needs no database, which is the point: a developer can
// scaffold a migration without one running.
func createMigration(migrationDir, migrationName string) {
	if migrationName == "" {
		fmt.Fprintln(os.Stderr, "migrate: -name is required for the create action")
		os.Exit(2)
	}

	if err := migration.CreateMigrationFiles(migrationDir, migrationName); err != nil {
		slog.Error("failed to create migration", "error", err)
		os.Exit(1)
	}
}

// printUsage writes to stderr, so `migrate -action=status > report.txt` captures the table and
// not the help text.
func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: migrate [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  migrate -action=create -name=create_users_table")
	fmt.Fprintln(os.Stderr, "  migrate -action=up")
	fmt.Fprintln(os.Stderr, "  migrate -action=down")
	fmt.Fprintln(os.Stderr, "  migrate -action=status")
}
