# Installation & Getting Started

Quick guide to set up and run Goilerplate on your local machine.

---

## 📋 Prerequisites

Before starting, make sure you have installed:

- **Go 1.24+** — [Download Go](https://golang.org/doc/install)
- **PostgreSQL** — [Download PostgreSQL](https://www.postgresql.org/download/)
- **Redis** (optional but recommended) — [Download Redis](https://redis.io/download)
- **Make** (optional, for Makefile commands)

---

## 🚀 Quick Start (5 minutes)

### 1. Clone Repository

```bash
git clone https://github.com/arisatriop/goilerplate.git
cd goilerplate
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Setup Configuration

```bash
cp config/config.example.yaml config/config.yaml
```

Edit `config/config.yaml` with your database credentials:

```yaml
db:
  driver: postgres
  host: localhost
  port: 5432
  name: goilerplate      # Adjust to your database name
  username: postgres     # Adjust to your username
  password: postgres     # Adjust to your password

redis:
  enabled: true
  host: localhost:6379
```

### 4. Create Database

```bash
createdb goilerplate
```

### 5. Run Migrations

```bash
go run cmd/migrate/main.go
```

### 6. Start Server

```bash
go run cmd/server/main.go
```

Server will start at `http://localhost:3000` 🎉

---

## ✅ Verify Installation

### Check Server Health

```bash
curl http://localhost:3000/health
```

Response:
```json
{
  "status": "OK",
  "database": "connected",
  "cache": "connected"
}
```

### Test Authentication

**Register:**
```bash
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "SecurePass123!",
    "avatar": "https://example.com/avatar.jpg"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!",
    "remember_me": false
  }'
```

---

## 📁 Project Structure

```
goilerplate/
├── cmd/               # Application entry points
│   ├── migrate/       # Database migrations
│   ├── server/        # HTTP server
│   └── worker/        # Background worker
├── config/            # Configuration files
├── internal/          # Application code
│   ├── bootstrap/     # App initialization
│   ├── delivery/      # HTTP layer
│   ├── domain/        # Business logic
│   ├── application/   # Services
│   ├── infrastructure/# Database & cache
│   └── wire/          # Dependency injection
├── pkg/               # Shared packages
├── docs/              # Documentation
├── deploy/            # Deployment configs
└── README.md          # Main README
```

See [Architecture Guide](../guides/architecture.md) for details.

---

## 🐳 Docker Setup (Alternative)

If you prefer using Docker:

```bash
# Start with Docker Compose (hot reload enabled)
make up
# or
docker compose up -d
```

Server will be accessible at `http://localhost:3000`

Source code is mounted as a volume, so changes will auto-rebuild.

See [Docker Guide](./docker.md) for details.

---

## 🔧 Development Setup

### Install Development Tools

```bash
# Air (hot reload)
go install github.com/air-verse/air@latest

# golangci-lint (linting)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Run with Hot Reload

```bash
air
```

Server will restart automatically every time a file changes.

See [Development Guide](./development.md) for detailed workflow.

---

## 🗄️ Database Setup

### PostgreSQL

**Install PostgreSQL:**

macOS:
```bash
brew install postgresql
```

Linux (Ubuntu):
```bash
sudo apt-get install postgresql postgresql-contrib
```

Windows:
- Download dari [postgresql.org](https://www.postgresql.org/download/windows/)

**Start PostgreSQL:**

macOS:
```bash
brew services start postgresql
```

Linux:
```bash
sudo systemctl start postgresql
```

**Create Database:**

```bash
createdb goilerplate
```

### MySQL (Alternative)

Edit `config/config.yaml`:

```yaml
db:
  driver: mysql
  host: localhost
  port: 3306
  name: goilerplate
  username: root
  password: your_password
```

Create database:
```bash
mysql -u root -p -e "CREATE DATABASE goilerplate;"
```

---

## ⚙️ Configuration

### Minimal Config

For quick start, minimum config required:

```yaml
db:
  driver: postgres
  host: localhost
  port: 5432
  name: goilerplate
  username: postgres
  password: postgres

jwt:
  secret_key: local-dev-key-only
```

### Full Configuration

See [Configuration Guide](../deployment/configuration.md) for all options.

---

## 🧪 Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/domain/user/...
```

---

## 🐛 Troubleshooting

### Port Already in Use

If port 3000 is already in use:

```yaml
# config/config.yaml
server:
  port: 3001  # Change port
```

### Database Connection Error

Check credentials in `config/config.yaml`:

```bash
psql -h localhost -U postgres -d goilerplate
```

### Redis Connection Error

If Redis error, disable redis in config:

```yaml
redis:
  enabled: false
```

### Migration Error

Run migrations:

```bash
go run cmd/migrate/main.go up
```

Rollback if needed:

```bash
go run cmd/migrate/main.go down
```

---

## 📚 Next Steps

- [Setup Development Environment](./development.md)
- [Learn Clean Architecture](../guides/architecture.md)
- [Create Your First CRUD](../guides/crud-operations.md)
- [Deploy to Production](../deployment/kubernetes.md)

---

## 🆘 Need Help?

- Check [Main README](../../README.md)
- See [Documentation Index](../README.md)
- Open [GitHub Issues](https://github.com/arisatriop/goilerplate/issues)
