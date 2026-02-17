# Makefile for Go Boilerplate

# Build and run commands
.PHONY: build run test clean migrate-up migrate-down migrate-status migrate-create

# Application
build:
	@echo "Building application..."
	go build -o bin/server cmd/server/main.go
	go build -o bin/migrate cmd/migrate/main.go

run:
	@echo "Running application..."
	go run cmd/server/main.go

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	go run cmd/migrate/main.go -action=up

migrate-down:
	@echo "Rolling back last migration..."
	go run cmd/migrate/main.go -action=down

migrate-status:
	@echo "Checking migration status..."
	go run cmd/migrate/main.go -action=status

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(name)"
	go run cmd/migrate/main.go -action=create -name=$(name)

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	go mod tidy

format:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	golangci-lint run

# Docker commands (if you use Docker)
docker-build:
	@echo "Building Docker image..."
	docker build -t goilerplate .

docker-run:
	@echo "Running Docker container..."
	docker run -p 3000:3000 goilerplate

docker-build-local:
	@echo "Building Local Docker image..."
	docker build -f Dockerfile.local -t goilerplate .

docker-run-local:
	@echo "Running Local Docker container..."
	docker run -p 3000:3000 -v $(shell pwd):/app goilerplate

up:
	@echo "Starting development environment..."
	docker-compose up --build

# Database helpers
db-reset: migrate-down migrate-up
	@echo "Database reset complete"

db-seed:
	@echo "Seeding database..."
	go run cmd/seed/main.go

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback last migration"
	@echo "  migrate-status - Check migration status"
	@echo "  migrate-create - Create new migration (usage: make migrate-create name=migration_name)"
	@echo "  dev-setup      - Setup development environment"
	@echo "  format         - Format code"
	@echo "  lint           - Run linter"
	@echo "  db-reset       - Reset database (down + up)"
	@echo "  help           - Show this help message"