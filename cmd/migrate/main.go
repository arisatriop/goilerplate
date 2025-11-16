package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"goilerplate/internal/bootstrap"
	"goilerplate/pkg/migration"
)

func main() {
	var (
		action        = flag.String("action", "", "Migration action: up, down, status, create")
		migrationName = flag.String("name", "", "Migration name (required for create action)")
		migrationDir  = flag.String("dir", "internal/migrations", "Migration directory")
	)
	flag.Parse()

	if *action == "" {
		printUsage()
		os.Exit(1)
	}

	if *action == "create" {
		if *migrationName == "" {
			log.Fatal("Migration name is required for create action")
		}

		if err := migration.CreateMigrationFiles(*migrationDir, *migrationName); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
		return
	}

	// For other actions, we need database connection
	app := bootstrap.Init()

	if app.DB == nil || app.DB.GDB == nil {
		log.Fatal("Database connection not available")
	}

	migrator := migration.NewMigrator(app.DB.GDB)

	switch *action {
	case "up":
		if err := migrator.Up(*migrationDir); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully!")

	case "down":
		if err := migrator.Down(*migrationDir); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		fmt.Println("Migration rolled back successfully!")

	case "status":
		if err := migrator.Status(*migrationDir); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

	default:
		fmt.Printf("Unknown action: %s\n", *action)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -action string")
	fmt.Println("        Migration action: up, down, status, create")
	fmt.Println("  -name string")
	fmt.Println("        Migration name (required for create action)")
	fmt.Println("  -config string")
	fmt.Println("        Config file path (default: config/config.yaml)")
	fmt.Println("  -dir string")
	fmt.Println("        Migration directory (default: internal/migrations)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  migrate -action=create -name=create_users_table")
	fmt.Println("  migrate -action=up")
	fmt.Println("  migrate -action=down")
	fmt.Println("  migrate -action=status")
}
