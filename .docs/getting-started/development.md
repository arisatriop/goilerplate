# Development Guide

Guide to set up development environment and workflow for contributing to Goilerplate.

---

## 🛠️ Development Setup

### Prerequisites

- Go 1.24+
- PostgreSQL
- Redis (optional but recommended)
- Make (optional)
- Your favorite code editor (VS Code, GoLand, Vim, etc)

### Initial Setup

```bash
# Clone repository
git clone https://github.com/arisatriop/goilerplate.git
cd goilerplate

# Install dependencies
go mod download

# Setup configuration
cp config/config.example.yaml config/config.yaml

# Create database
createdb goilerplate

# Run migrations
go run cmd/migrate/main.go

# Start development server (with hot reload)
air
```

---

## 🚀 Development Workflow

### Option 1: Local Development (Recommended for Go purists)

**Advantages:**
- Direct Go development experience
- Fast feedback loop
- Easy debugging with `go run` or IDE

**Setup:**

1. Install air (hot reload):
   ```bash
   go install github.com/air-verse/air@latest
   ```

2. Run with air:
   ```bash
   air
   ```

3. Server restarts automatically when files change

### Option 2: Docker Development (Recommended for consistency)

**Advantages:**
- Consistent with team
- All services (DB, Redis) containerized
- Easy cleanup with `docker compose down`

**Setup:**

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f app

# Shell into container
docker compose exec app bash
```

---

## 📝 Making Code Changes

### Before Starting Development

1. **Create feature branch:**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Install development tools:**
   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

### Workflow

1. **Make changes** to code
2. **Server auto-restarts** (via air or compose)
3. **Test manually** via curl or API client
4. **Run tests:**
   ```bash
   go test ./...
   ```
5. **Lint code:**
   ```bash
   golangci-lint run ./...
   ```
6. **Commit changes:**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

---

## 🧪 Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Test

```bash
# Test specific package
go test ./internal/domain/user/...

# Run specific test function
go test -run TestUserCreate ./internal/domain/user/...
```

### Test Organization

```
internal/
├── domain/
│   └── user/
│       ├── entity.go
│       ├── usecase.go
│       ├── usecase_test.go      # Domain tests
│       └── repository.go
├── infrastructure/
│   └── repository/
│       ├── user.go
│       └── user_test.go          # Repository tests
└── delivery/http/handler/
    ├── user.go
    └── user_test.go              # Handler tests
```

---

## 📚 Creating CRUD Operations

To create new CRUD operations, follow [CRUD Operations Guide](../guides/crud-operations.md).

**Step by step:**

1. Define domain (entity, repository, usecase)
2. Create database model
3. Implement repository
4. Create DTOs & converters
5. Create handler
6. Setup Wire bindings
7. Register routes
8. Test everything

---

## 🐛 Debugging

### Using IDE Debugger

**VS Code:**

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch",
      "type": "go",
      "request": "launch",
      "mode": "local",
      "program": "${fileDirname}",
      "env": {},
      "args": []
    }
  ]
}
```

**GoLand:**

- Built-in debugging, press `Ctrl+F5` or `Run → Debug`

### Using Print Debugging

```go
import "log"

func MyHandler(ctx *fiber.Ctx) error {
    log.Println("Debug:", data) // Will print to console
    return nil
}
```

### Using Structured Logging

```go
import "github.com/your-logger"

func MyHandler(ctx *fiber.Ctx) error {
    log.Debug("Handler called", "data", data)
    return nil
}
```

---

## 🔍 Code Quality Tools

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run ./...
```

### Formatting

```bash
# Format all Go files
go fmt ./...

# Using goimports (organize imports)
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

### Testing & Coverage

```bash
# Test dengan coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 📖 Project Structure Deep Dive

```
internal/
├── bootstrap/           # App initialization & setup
├── delivery/           # HTTP layer (Presentation)
│   └── http/
│       ├── dto/        # Data Transfer Objects
│       ├── handler/    # HTTP handlers (thin layer)
│       ├── middleware/ # HTTP middleware
│       ├── presenter/  # Domain → DTO conversion
│       ├── request/    # HTTP → Domain conversion
│       └── router/     # Route definitions
├── domain/             # Business logic (Core)
│   ├── user/
│   ├── product/
│   └── order/
├── application/        # Services (Orchestration)
│   └── use_cases/
└── infrastructure/     # External integrations
    ├── model/          # GORM models
    ├── repository/     # Database operations
    ├── cache/          # Redis integration
    └── transaction/    # DB transaction handling
```

**Key principle:** Dependencies point inward (Delivery → Application → Domain → Infrastructure)

---

## 🚀 Making Your First Contribution

### Create Feature

1. Create branch: `git checkout -b feature/users-api`
2. Follow [CRUD Operations Guide](../guides/crud-operations.md)
3. Write tests
4. Test manually
5. Run `go test ./...` & linter

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git commit -m "feat: add user delete endpoint"
git commit -m "fix: resolve nil pointer in auth middleware"
git commit -m "docs: update API documentation"
```

### Push & Create PR

```bash
git push origin feature/users-api
```

Then open PR on GitHub.

---

## 📋 Development Checklist

- [ ] Setup environment (Go, DB, Redis)
- [ ] Clone repository
- [ ] Install dependencies (`go mod download`)
- [ ] Setup config (`cp config/config.example.yaml config/config.yaml`)
- [ ] Create database
- [ ] Run migrations
- [ ] Start server (`air` or `docker compose up`)
- [ ] Test endpoints (curl or Postman)
- [ ] Read architecture docs
- [ ] Start building features!

---

## 🔗 Related Documentation

- [Installation Guide](./installation.md) - Initial setup
- [Docker Setup](./docker.md) - Docker development
- [CRUD Operations](../guides/crud-operations.md) - Create new endpoints
- [Architecture](../guides/architecture.md) - Understand the structure
- [Router Guide](../api/router.md) - API route organization

---

## 🔭 Observability (Local)

To enable distributed tracing and log aggregation locally, start the observability stack from the `docker-image` repo:

```bash
cd ../docker-image/observability
docker compose up -d
```

Then enable OTel in `config/config.yaml`:

```yaml
otel:
  enabled: true
  endpoint: localhost:4317
  insecure: true
```

Run the app with log redirect so Promtail can scrape logs into Loki:

```bash
# Using air
make run   # .air.toml already configured to tee logs to storage/logs/app.log

# Using nodemon
mkdir -p ./storage/logs && nodemon -e go --signal SIGINT --exec "sh -c 'go run cmd/server/main.go 2>&1 | tee ./storage/logs/app.log'"
```

Open Grafana at `http://localhost:3001` → Explore → select **Tempo** for traces or **Loki** for logs.

See [Observability Guide](../guides/observability.md) for full details.

---

## 💡 Pro Tips

1. **Use IDE features** - Most IDEs have excellent Go support
2. **Terminal multiplexing** - Use tmux or screen for multiple terminals
3. **Hot reload** - Use `air` or Docker for instant feedback
4. **Watch tests** - Some tools watch & re-run tests on file changes
5. **Git hooks** - Setup pre-commit hooks for formatting

---

## 🆘 Common Issues

### "go: no matching versions"

```bash
go get -u ./...
go mod tidy
```

### Air not working

```bash
# Reinstall air
go install github.com/air-verse/air@latest

# Or restart manually
go run cmd/server/main.go
```

### Database connection failed

```bash
# Check PostgreSQL running
psql -l

# Check config/config.yaml settings
cat config/config.yaml
```

### Tests failing

```bash
# Run specific test for debugging
go test -v -run TestName ./...

# Check if migrations ran
go run cmd/migrate/main.go
```

---

## 📚 Learning Resources

- [Go Official Docs](https://golang.org/doc/)
- [Fiber Web Framework](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Project Layout](https://github.com/golang-standards/project-layout)
