# Makefile for Go Boilerplate

# Build and run commands
.PHONY: build run test test-quick test-integration cover vuln clean \
	migrate-up migrate-down migrate-status migrate-create \
	swag dev-setup format lint lint-install \
	docker-build docker-run docker-build-local docker-run-local up help

# Application
build:
	@echo "Building application..."
	go build -o bin/server cmd/server/main.go
	go build -o bin/migrate cmd/migrate/main.go

run:
	@echo "Running application..."
	sh ./.scripts/run.sh

# -race is what CI runs, so it is what `make test` runs. A race that only the detector sees is
# still a race, and finding it here costs seconds where finding it in production costs a day.
# -shuffle=on catches tests that pass only because of the order they run in.
test:
	@echo "Running tests..."
	go test -race -shuffle=on ./...

# Without the detector, for a quick loop while iterating. Not what the gate runs, so a green
# `make test-quick` proves less than a green `make test`.
test-quick:
	@echo "Running tests without the race detector..."
	go test ./...

# Integration tests need PostgreSQL and Redis. Without POSTGRES_TEST_DSN and REDIS_TEST_ADDR
# they skip rather than fail, which is why `make test` alone can look green while the
# repository and cache-mode suites never ran.
TEST_POSTGRES_DSN ?= host=localhost port=5432 user=postgres dbname=goilerplate_test sslmode=disable timezone=UTC
TEST_REDIS_ADDR ?= localhost:6379

test-integration:
	@echo "Running tests with PostgreSQL and Redis..."
	POSTGRES_TEST_DSN="$(TEST_POSTGRES_DSN)" REDIS_TEST_ADDR="$(TEST_REDIS_ADDR)" go test -race -shuffle=on ./...

# Coverage against the real PostgreSQL and Redis: without them the repository, cache and
# integration suites skip, and the number that comes out is meaningless.
cover:
	@echo "Measuring coverage..."
	POSTGRES_TEST_DSN="$(TEST_POSTGRES_DSN)" REDIS_TEST_ADDR="$(TEST_REDIS_ADDR)" \
		go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | tail -1
	@echo "Per-package detail: go tool cover -html=coverage.out"

vuln:
	@echo "Checking dependencies for known vulnerabilities..."
	@command -v govulncheck >/dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/ coverage.out

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

# Swagger docs
swag:
	@echo "Generating Swagger docs..."
	swag init -g cmd/server/main.go -o .swagger/

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	go mod tidy

format:
	@echo "Formatting code..."
	go fmt ./...

lint-install:
	@echo "Installing golangci-lint v2.12.0..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew install golangci-lint || brew upgrade golangci-lint; \
	else \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.12.0; \
	fi

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

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests with -race (integration suites skip without a test DB)"
	@echo "  test-quick     - Run tests without the race detector (fast loop)"
	@echo "  test-integration - Run tests against PostgreSQL and Redis"
	@echo "  cover          - Coverage against a real PostgreSQL and Redis"
	@echo "  vuln           - Scan dependencies with govulncheck"
	@echo "  clean          - Clean build artifacts"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback last migration"
	@echo "  migrate-status - Check migration status"
	@echo "  migrate-create - Create new migration (usage: make migrate-create name=migration_name)"
	@echo "  dev-setup      - Setup development environment"
	@echo "  format         - Format code"
	@echo "  lint           - Run linter"
	@echo "  swag           - Generate Swagger docs"
	@echo "  help           - Show this help message"