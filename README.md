# 🚀 Goilerplate

**Production-ready Go backend boilerplate** with authentication, authorization, and best practices built-in.

A clean, scalable Go REST API template featuring JWT authentication, role-based access control (RBAC), Redis caching, and clean architecture patterns.

---

## ✨ Features

### 🔐 Authentication & Authorization

- **JWT-based authentication** with access & refresh tokens
- **Token blacklisting** for immediate logout
- **Session management** with device tracking
- **Role-based access control (RBAC)** with permissions
- **Multi-device login support**

### 🏗️ Architecture

- **Clean Architecture** (Handler → Usecase → Repository)
- **Separation of Concerns** with proper layering
- **Dependency Injection** using Wire
- **Middleware pattern** for cross-cutting concerns
- **Global error handling** with consistent responses

### 🛠️ Technical Features

- **Redis caching** for sessions, tokens, and permissions
- **Database support** for PostgreSQL and MySQL
- **Request validation** with go-playground/validator
- **Structured logging** with slog
- **Configuration management** with Viper
- **Database migrations** support
- **Password hashing** with bcrypt

### 📦 Code Quality

- **DRY principle** applied throughout
- **Reusable helpers** and utilities
- **Consistent error handling**
- **Type-safe context passing**
- **Comprehensive comments** and documentation

---

## 🛠️ Tech Stack

### Core

- **[Go 1.24](https://golang.org/)** - Programming language
- **[Fiber v2](https://gofiber.io/)** - Fast HTTP web framework
- **[GORM](https://gorm.io/)** - ORM for database operations

### Authentication & Security

- **[JWT](https://github.com/golang-jwt/jwt)** - JSON Web Tokens
- **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** - Password hashing
- **[Validator](https://github.com/go-playground/validator)** - Request validation

### Database

- **[PostgreSQL](https://www.postgresql.org/)** - Primary database (also supports MySQL)
- **[Redis](https://redis.io/)** - Caching and session storage
- **[go-redis](https://github.com/redis/go-redis)** - Redis client

### Configuration & Tools

- **[Viper](https://github.com/spf13/viper)** - Configuration management
- **[Wire](https://github.com/google/wire)** - Dependency injection
- **[UUID](https://github.com/google/uuid)** - Unique ID generation

---

## 📁 Project Structure

```
be-boilerplate-go/
├── cmd/
│   ├── migrate/          # Database migration commands
│   ├── server/           # HTTP server entry point
│   └── worker/           # Background worker entry point
├── config/               # Configuration files
│   └── config.example.yaml
├── internal/
│   ├── bootstrap/        # Application bootstrap
│   ├── delivery/         # Delivery layer (HTTP handlers)
│   │   └── http/
│   │       ├── dto/      # Data Transfer Objects
│   │       ├── handler/  # HTTP handlers
│   │       ├── middleware/ # HTTP middleware
│   │       └── router/   # Route definitions
│   ├── domain/           # Business logic layer
│   │   └── auth/         # Auth domain (usecase, repository, services)
│   ├── infrastructure/   # Infrastructure layer
│   │   ├── model/        # Database models
│   │   └── repository/   # Repository implementations
│   └── wire/             # Wire dependency injection
├── pkg/                  # Shared packages
│   ├── auth/             # Auth utilities
│   ├── constants/        # Application constants
│   ├── jwt/              # JWT utilities
│   ├── logger/           # Logging utilities
│   ├── response/         # HTTP response helpers
│   └── utils/            # Common utilities
├── go.mod
├── go.sum
└── README.md
```

---

## 🚀 Getting Started

### Prerequisites

- **Go 1.24+** installed
- **PostgreSQL** or **MySQL** database
- **Redis** server (optional, but recommended)
- **Make** (optional, for Makefile commands)

### Installation

1. **Clone the repository**

```bash
git clone https://github.com/arisatriop/goilerplate.git
cd goilerplate
```

2. **Install dependencies**

```bash
go mod download
```

3. **Copy configuration file**

```bash
cp config/config.example.yaml config/config.yaml
```

4. **Configure your environment**

Edit `config/config.yaml` with your settings:

```yaml
db:
  driver: postgres
  host: localhost
  port: 5432
  name: your_database
  username: your_username
  password: your_password

redis:
  enabled: true
  host: localhost:6379
  password: ""

jwt:
  secret_key: your-super-secret-jwt-key
  access_secret: your-access-token-secret
  refresh_secret: your-refresh-token-secret
```

5. **Run database migrations**

```bash
go run cmd/migrate/main.go
```

6. **Start the server**

```bash
go run cmd/server/main.go
```

The server will start at `http://localhost:3000` by default.

---

## 📚 API Endpoints

### Authentication

| Method | Endpoint               | Description            | Auth Required    |
| ------ | ---------------------- | ---------------------- | ---------------- |
| POST   | `/api/auth/register`   | Register new user      | ❌               |
| POST   | `/api/auth/login`      | User login             | ❌               |
| POST   | `/api/auth/logout`     | Logout current session | ✅               |
| POST   | `/api/auth/logout-all` | Logout all devices     | ✅               |
| POST   | `/api/auth/refresh`    | Refresh access token   | 🔄 Refresh Token |

### Example Requests

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

**Logout:**

```bash
curl -X POST http://localhost:3000/api/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Refresh Token:**

```bash
curl -X POST http://localhost:3000/api/auth/refresh \
  -H "Authorization: Bearer YOUR_REFRESH_TOKEN"
```

---

## 🔧 Configuration

### Environment Variables

You can override config values using environment variables:

```bash
export APP_ENV=production
export SERVER_PORT=8080
export DB_HOST=your-db-host
export REDIS_HOST=your-redis-host
```

### Configuration File

The `config/config.yaml` supports:

```yaml
app:
  env: local # Environment: local, development, production
  name: Goilerplate
  version: 1.0.0

server:
  host: localhost
  port: 3000
  prefork: false # Enable prefork for production
  read_timeout: 5s
  write_timeout: 5s
  idle_timeout: 120s
  enable_cors: true

db:
  driver: postgres # postgres or mysql
  host: localhost
  port: 5432
  name: postgres
  username: postgres
  password: postgres
  min_open_connections: 10
  max_open_connections: 100

redis:
  enabled: true # Set false to disable Redis
  host: localhost:6379
  password: ""
  db: 0

jwt:
  access_token_expiry: 15m # Access token TTL
  refresh_token_expiry: 168h # Refresh token TTL (7 days)

log:
  level: debug # debug, info, warn, error
  source: false # Include source code location
```

---

## 🎯 Key Features Explained

### 1. **JWT Authentication Flow**

```
Login → Generate Access Token (15m) + Refresh Token (7d)
     → Store in Redis + Database
     → Return both tokens to client

Protected Request → Validate Access Token
                 → Check Redis blacklist
                 → Verify DB existence
                 → Allow/Deny access

Token Expired → Use Refresh Token
             → Generate new Access Token
             → Return new token pair

Logout → Blacklist current token
      → Remove from Redis + DB
      → Deactivate session

Logout All → Blacklist all user tokens
          → Remove all from Redis + DB
          → Deactivate all sessions
```

### 2. **Role-Based Access Control (RBAC)**

```
User → Has Roles → Has Permissions
    → Direct Permissions (overrides)
    → Menu-based Permissions (hierarchical)

Permission Check:
1. Check user-specific override (is_granted true/false)
2. Check role permissions
3. Check menu permissions (with tree traversal)
```

### 3. **Redis Caching Strategy**

- **Sessions:** Cached for fast validation
- **Tokens:** Cached to avoid DB hits on every request
- **Permissions:** Cached to speed up authorization checks
- **Blacklist:** Immediate token invalidation

**Critical:** When Redis is enabled, cache writes are **critical** (must succeed)

### 4. **Clean Architecture Layers**

```
HTTP Request
    ↓
Middleware (Auth, CORS, Logger)
    ↓
Handler (Parse, Validate, Extract Context)
    ↓
Usecase (Business Logic)
    ↓
Repository (Database Operations)
    ↓
Database
```

### 5. **Error Handling**

```go
// Global error handler
response.HandleError(ctx, err)

// Automatically handles:
- Client errors (400, 401, 403, 404, 409, etc.)
- Server errors (500)
- Validation errors (422)
- Custom error responses
```

---

## 🧪 Development

### Running Tests

```bash
go test ./...
```

### Running with Hot Reload

```bash
# Install air
go install github.com/air-verse/air@latest

# Run with air
air
```

### Building for Production

```bash
# Build binary
go build -o bin/server cmd/server/main.go

# Run binary
./bin/server
```

### Database Migrations

```bash
# Run migrations
go run cmd/migrate/main.go up

# Rollback migrations
go run cmd/migrate/main.go down

# Create new migration
go run cmd/migrate/main.go create migration_name
```

---

## 🔒 Security Best Practices

This boilerplate implements:

- ✅ **Password hashing** with bcrypt (cost 12)
- ✅ **JWT signature verification** with HS256
- ✅ **Token blacklisting** for immediate revocation
- ✅ **CORS protection** with configurable origins
- ✅ **Request validation** with strict rules
- ✅ **SQL injection prevention** via GORM parameterization
- ✅ **Session tracking** with device information
- ✅ **Secure password requirements** (min length, complexity)
- ✅ **Rate limiting** ready (add middleware as needed)

---

## 📝 Code Quality

### Principles Applied

- **DRY (Don't Repeat Yourself)** - Reusable helpers and utilities
- **SOLID** - Single responsibility, Open/closed, Liskov substitution
- **Clean Code** - Readable, maintainable, self-documenting
- **Separation of Concerns** - Clear boundaries between layers
- **Defensive Programming** - Input validation, error handling

### Best Practices

- ✅ Middleware validates tokens once (no duplicate validation)
- ✅ Context carries authenticated user info
- ✅ Handlers trust middleware (no re-validation)
- ✅ Global error handler for consistency
- ✅ Builder patterns for complex structs
- ✅ Type-safe constants and enums
- ✅ Comprehensive error messages
- ✅ Structured logging with context

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Commit Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `chore:` - Maintenance tasks

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

Built with modern Go best practices and inspired by:

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) by Robert C. Martin
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Fiber Framework](https://gofiber.io/) documentation
- Community best practices and patterns

---

## 📧 Support

If you have any questions or need help, please:

- Open an issue on GitHub
- Check existing documentation
- Review the code examples

---

**Made with ❤️ using Go**

_Star ⭐ this repo if you find it useful!_
