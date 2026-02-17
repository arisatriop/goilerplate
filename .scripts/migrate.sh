#!/bin/bash

# Migration helper script for Go Boilerplate

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATION_DIR="$PROJECT_ROOT/internal/migrations"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ ${1}${NC}"
}

print_success() {
    echo -e "${GREEN}✅ ${1}${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  ${1}${NC}"
}

print_error() {
    echo -e "${RED}❌ ${1}${NC}"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 {up|down|status|create|reset} [options]"
    echo ""
    echo "Commands:"
    echo "  up              Run all pending migrations"
    echo "  down            Rollback the last migration"
    echo "  status          Show migration status"
    echo "  create <name>   Create a new migration"
    echo "  reset           Reset database (down all + up all)"
    echo ""
    echo "Options:"
    echo "  -h, --help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 up"
    echo "  $0 create add_users_table"
    echo "  $0 status"
    echo "  $0 reset"
}

# Function to run migrations
run_migration() {
    local action=$1
    local name=$2
    
    cd "$PROJECT_ROOT"
    
    case $action in
        "up")
            print_info "Running database migrations..."
            if go run cmd/migrate/main.go -action=up; then
                print_success "Migrations completed successfully!"
            else
                print_error "Migration failed!"
                exit 1
            fi
            ;;
        "down")
            print_warning "Rolling back last migration..."
            if go run cmd/migrate/main.go -action=down; then
                print_success "Migration rolled back successfully!"
            else
                print_error "Rollback failed!"
                exit 1
            fi
            ;;
        "status")
            print_info "Checking migration status..."
            go run cmd/migrate/main.go -action=status
            ;;
        "create")
            if [ -z "$name" ]; then
                print_error "Migration name is required for create command"
                echo "Usage: $0 create <migration_name>"
                exit 1
            fi
            print_info "Creating new migration: $name"
            if go run cmd/migrate/main.go -action=create -name="$name"; then
                print_success "Migration files created successfully!"
                echo ""
                print_info "Next steps:"
                echo "1. Edit the generated .up.sql file with your migration"
                echo "2. Edit the generated .down.sql file with the rollback"
                echo "3. Run: $0 up"
            else
                print_error "Failed to create migration!"
                exit 1
            fi
            ;;
        "reset")
            print_warning "This will reset your database (rollback all migrations and re-run them)"
            read -p "Are you sure? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                print_info "Resetting database..."
                # First, rollback all migrations
                while go run cmd/migrate/main.go -action=down 2>/dev/null; do
                    print_info "Rolled back a migration..."
                done
                # Then run all migrations
                if go run cmd/migrate/main.go -action=up; then
                    print_success "Database reset completed!"
                else
                    print_error "Database reset failed!"
                    exit 1
                fi
            else
                print_info "Database reset cancelled."
            fi
            ;;
        *)
            print_error "Unknown command: $action"
            show_usage
            exit 1
            ;;
    esac
}

# Main script logic
main() {
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi

    # Check if we're in the right directory
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        print_error "This script must be run from the project root or scripts directory"
        exit 1
    fi

    # Parse arguments
    case ${1:-""} in
        "-h"|"--help"|"help"|"")
            show_usage
            exit 0
            ;;
        "create")
            run_migration "create" "$2"
            ;;
        "up"|"down"|"status"|"reset")
            run_migration "$1"
            ;;
        *)
            print_error "Unknown command: $1"
            show_usage
            exit 1
            ;;
    esac
}

# Run the main function
main "$@"