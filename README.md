# 🚀 Goilerplate

**Production-ready Go backend boilerplate** with comprehensive authentication, authorization, and enterprise-grade features built-in.

A clean, scalable Go REST API template featuring JWT authentication, role-based access control (RBAC), Redis caching, file storage (local/S3), PostgreSQL support, and clean architecture patterns.

---

## ✨ Features

### 🔐 Authentication & Authorization

- **JWT-based authentication** with access & refresh tokens
- **Token blacklisting** for immediate logout
- **Session management** with device tracking
- **Role-based access control (RBAC)** with hierarchical permissions
- **Menu-based permissions** with tree structure
- **User-specific permission overrides**
- **Multi-device login support**
- **OAuth2 integration ready** (Google, etc.)

### 🏗️ Architecture

- **Clean Architecture** (Handler → Usecase → Repository)
- **Separation of Concerns** with proper layering
- **Dependency Injection** using Wire
- **Middleware pattern** for cross-cutting concerns
- **Global error handling** with consistent responses

### 🛠️ Technical Features

- **Redis caching** for sessions, tokens, and permissions
- **PostgreSQL primary support** with MySQL compatibility
- **File storage** with local and AWS S3 support
- **Database migrations** with up/down support
- **Request validation** with go-playground/validator
- **Structured logging** with slog
- **Configuration management** with Viper
- **Password encryption** with bcrypt
- **Pagination helpers** with metadata
- **UUID generation** and utilities
- **Decimal handling** for financial data

### 📦 Code Quality

- **DRY principle** applied throughout
- **Reusable helpers** and utilities
- **Consistent error handling**
- **Type-safe context passing**
- **Comprehensive comments** and documentation

---

## 🛠️ Tech Stack

### Core

- **[Go 1.24.3](https://golang.org/)** - Programming language
- **[Fiber v2](https://gofiber.io/)** - Fast HTTP web framework
- **[GORM](https://gorm.io/)** - ORM for database operations

### Authentication & Security

- **[JWT](https://github.com/golang-jwt/jwt)** - JSON Web Tokens
- **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** - Password hashing
- **[Validator](https://github.com/go-playground/validator)** - Request validation

### Database & Storage

- **[PostgreSQL](https://www.postgresql.org/)** - Primary database with full UUID support
- **[MySQL](https://www.mysql.com/)** - Alternative database support
- **[Redis](https://redis.io/)** - Caching and session storage
- **[go-redis](https://github.com/redis/go-redis)** - Redis client
- **[AWS S3](https://aws.amazon.com/s3/)** - Cloud file storage
- **Local Storage** - File system storage option

### Configuration & Tools

- **[Viper](https://github.com/spf13/viper)** - Configuration management
- **[Wire](https://github.com/google/wire)** - Dependency injection
- **[UUID](https://github.com/google/uuid)** - Unique ID generation
- **[Decimal](https://github.com/shopspring/decimal)** - Precise decimal arithmetic
- **[OAuth2](https://golang.org/x/oauth2)** - OAuth2 client implementation

---

## 📁 Project Structure

```
goilerplate/
├── cmd/
│   ├── migrate/          # Database migration commands
│   ├── server/           # HTTP server entry point
│   └── worker/           # Background worker entry point
├── config/               # Configuration files
│   ├── config.example.yaml
│   └── config.yaml
├── internal/
│   ├── application/      # Application services layer for multi-domain orchestration
│   ├── bootstrap/        # Application bootstrap
│   │   ├── database/     # Database setup
│   │   └── *.go          # App, Fiber, Redis, Viper setup
│   ├── constants/        # Application constants
│   ├── delivery/         # Delivery layer (HTTP handlers)
│   ├── domain/           # Business logic layer
│   ├── infrastructure/   # Infrastructure layer
│   │   ├── context/      # Context utilities
│   │   ├── model/        # Database models (PostgreSQL-optimized)
│   │   ├── repository/   # Repository implementations
│   │   └── transaction/  # Transaction management
│   ├── migrations/       # Database migrations (PostgreSQL)
│   │   ├── *_create_users_table.{up,down}.sql
│   │   ├── *_create_roles_table.{up,down}.sql
│   │   ├── *_create_permissions_table.{up,down}.sql
│   │   └── ... (RBAC tables)
│   └── wire/             # Wire dependency injection
├── pkg/                  # Shared packages
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 🚀 Getting Started

### Prerequisites

- **Go 1.24.3+** installed
- **PostgreSQL 13+** database (primary support)
- **Redis 6+** server (recommended for production)
- **Make** (optional, for Makefile commands)
- **AWS Account** (optional, for S3 file storage)

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
app:
  env: local
  name: Goilerplate
  version: 1.0.0

server:
  host: localhost
  port: 3000

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
  access_token_expiry: 15m
  refresh_token_expiry: 168h

filesystem:
  default_driver: local # or "s3"
  drivers:
    local:
      root_path: ./storage
    s3:
      region: us-east-1
      bucket: your-bucket-name
      access_key_id: your-access-key
      secret_access_key: your-secret-key
```

5. **Run database migrations**

```bash
# Run all migrations
go run cmd/migrate/main.go -action=up

# Or rollback
go run cmd/migrate/main.go -action=down
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

### User Management

| Method | Endpoint             | Description      | Auth Required |
| ------ | -------------------- | ---------------- | ------------- |
| GET    | `/api/users/profile` | Get current user | ✅            |
| PUT    | `/api/users/profile` | Update profile   | ✅            |
| POST   | `/api/users/avatar`  | Upload avatar    | ✅            |

### Role & Permission Management

| Method | Endpoint                    | Description           | Auth Required |
| ------ | --------------------------- | --------------------- | ------------- |
| GET    | `/api/roles`                | List all roles        | ✅            |
| POST   | `/api/roles`                | Create new role       | ✅            |
| PUT    | `/api/roles/:id`            | Update role           | ✅            |
| DELETE | `/api/roles/:id`            | Delete role           | ✅            |
| GET    | `/api/permissions`          | List all permissions  | ✅            |
| POST   | `/api/users/:id/roles`      | Assign role to user   | ✅            |
| DELETE | `/api/users/:id/roles/:rid` | Remove role from user | ✅            |

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
export SERVER_PORT=80
export DB_HOST=your-db-host
export DB_PORT=5432
export REDIS_HOST=localhost:6379
export JWT_SECRET_KEY=your-super-secret-jwt-key
export FILESYSTEM_DRIVER=s3
export LOG_LEVEL=info
```

### Complete Configuration Reference

The `config/config.yaml` supports all these options:

```yaml
app:
  env: local # Environment: local, development, production
  name: Goilerplate # Application name
  version: 1.0.0 # Application version
  description: "Your app description"

server:
  host: localhost # Server host
  port: 80 # Server port
  prefork: false # Enable prefork for production
  read_timeout: 5s # Request read timeout
  write_timeout: 5s # Response write timeout
  idle_timeout: 120s # Keep-alive timeout
  enable_cors: true # Enable CORS middleware

db:
  driver: postgres # Database driver: postgres or mysql
  host: localhost # Database host
  port: 5432 # Database port (5432 for postgres, 3306 for mysql)
  name: goilerplate # Database name
  sslmode: disable # SSL mode: disable, require, verify-ca, verify-full
  username: postgres # Database username
  password: postgres # Database password
  min_open_connections: 10 # Minimum open connections
  max_open_connections: 100 # Maximum open connections
  connection_max_lifetime: 300 # Connection max lifetime (seconds)
  connection_max_idle_time: 60 # Connection max idle time (seconds)
  health_check_period: 30 # Health check period (seconds)

redis:
  enabled: true # Enable/disable Redis
  host: localhost:6379 # Redis host:port
  password: "" # Redis password (empty if none)
  db: 0 # Redis database number
  dial_timeout: 5s # Connection timeout
  read_timeout: 5s # Read timeout
  write_timeout: 5s # Write timeout
  pool_size: 10 # Connection pool size
  pool_timeout: 6s # Pool timeout

jwt:
  secret_key: your-super-secret-jwt-key # Main JWT secret
  access_secret: your-access-token-secret # Access token secret
  refresh_secret: your-refresh-token-secret # Refresh token secret
  access_token_expiry: 15m # Access token TTL (15 minutes)
  refresh_token_expiry: 168h # Refresh token TTL (1 week)
  issuer: goilerplate # JWT issuer name

filesystem:
  driver: local # Storage driver: local, s3, r2, drive
  max_file_size: 3145728 # Max file size in bytes (3MB)

  # Local filesystem storage
  local:
    base_path: ./storage # Local storage directory
    base_url: http://localhost:80/storage # Base URL for file access

  # AWS S3 configuration
  s3:
    access_key_id: your-access-key # AWS access key
    secret_access_key: your-secret-key # AWS secret key
    region: us-east-1 # AWS region
    bucket: your-bucket-name # S3 bucket name
    endpoint: "" # Custom endpoint (for MinIO, etc.)

  # Google Drive configuration
  drive:
    credentials_file: ./credentials/gdrive-sa-credentials.json
    client_id: your-client-id
    client_secret: your-client-secret
    refresh_token: your-refresh-token
    folder_id: your-folder-id # Drive folder ID

log:
  level: debug # Log level: debug, info, warn, error
  source: false # Include source code location in logs

crypto:
  encryption_key: your-32-byte-encryption-key-here # 32-byte encryption key
```

### Storage Driver Options

#### Local Storage

- **Use case:** Development, small deployments
- **Pros:** Simple setup, no external dependencies
- **Cons:** Not scalable, single server limitation

#### AWS S3

- **Use case:** Production, scalable cloud storage
- **Pros:** Highly scalable, reliable, CDN integration
- **Cons:** Requires AWS account, costs money

#### Google Drive

- **Use case:** Integration with Google Workspace
- **Pros:** Large free storage, familiar interface
- **Cons:** API rate limits, not ideal for high-traffic apps

### Database Configuration

#### PostgreSQL (Recommended)

```yaml
db:
  driver: postgres
  host: localhost
  port: 5432
  name: goilerplate
  sslmode: disable # Use 'require' in production
```

#### MySQL (Alternative)

```yaml
db:
  driver: mysql
  host: localhost
  port: 3306
  name: goilerplate
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

### 6. **Clean Architecture Layers**

```
Filesystem Interface
↓
Driver Factory (Local/S3)
↓
Storage Operations (Put, Get, Delete, List)
↓
Token-based Access (Temporary URLs)

Supported Drivers:
- Local: File system storage with configurable root path
- S3: AWS S3 with regions, custom endpoints, credentials

Features:
- Unified interface for all storage backends
- Temporary download URLs with expiration
- Automatic file type detection
- Directory management
- Secure file access with tokens
```

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

### 7. **Error Handling**

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
# Run migrations up
go run cmd/migrate/main.go -action=up

# Rollback migrations
go run cmd/migrate/main.go -action=down

# Check migration status
go run cmd/migrate/main.go -action=status
```

### Makefile Commands

```bash
# Development
make run          # Run server
make run-migrate  # Run migrations

# Build
make build        # Build server binary
make build-migrate # Build migrate binary

# Testing
make test         # Run tests
make test-verbose # Run tests with verbose output

# Cleanup
make clean        # Clean build artifacts
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
- ✅ **File upload validation** (type, size restrictions)
- ✅ **UUID-based IDs** (no sequential enumeration)
- ✅ **PostgreSQL security** (prepared statements, transactions)
- ✅ **Rate limiting ready** (add middleware as needed)

---

## 📝 Code Quality

### Principles Applied

- **DRY (Don't Repeat Yourself)** - Reusable helpers and utilities
- **SOLID** - Single responsibility, Open/closed, Liskov substitution
- **Clean Code** - Readable, maintainable, self-documenting
- **Separation of Concerns** - Clear boundaries between layers
- **Defensive Programming** - Input validation, error handling

### Best Practices

- ✅ PostgreSQL-optimized models with proper UUID types
- ✅ Middleware validates tokens once (no duplicate validation)
- ✅ Context carries authenticated user info
- ✅ Handlers trust middleware (no re-validation)
- ✅ Global error handler for consistency
- ✅ Builder patterns for complex structs
- ✅ Type-safe constants and enums
- ✅ Comprehensive error messages
- ✅ Structured logging with context
- ✅ File storage abstraction for scalability
- ✅ Migration system for database versioning
- ✅ Decimal arithmetic for financial precision

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
