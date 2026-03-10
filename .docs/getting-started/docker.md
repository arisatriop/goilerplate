# Docker Setup Guide

Guide to using Docker and Docker Compose for development and production.

---

## 🐳 Development with Docker (Recommended)

The easiest way to start development with hot reload.

### Prerequisites

- **Docker** — [Install Docker](https://docs.docker.com/get-docker/)
- **Docker Compose** — Usually included with Docker Desktop

### Quick Start

```bash
# Start environment
make up
# or
docker compose up -d
```

Server will be accessible at `http://localhost:3000` and automatically reload when files change.

### What Happens

Docker Compose will start:

1. **PostgreSQL** — Database on port 5432
2. **Redis** — Cache on port 6379
3. **Go App** — Server on port 3000 (hot reload enabled via `air`)

Source code is mounted as a volume, so changes are immediately reflected in the container.

---

## 📋 Development Docker Compose Setup

File: `docker-compose.yml`

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.local
    ports:
      - "3000:3000"
    volumes:
      - .:/app
      - /app/node_modules
    environment:
      - APP_ENV=local
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=goilerplate
      - DB_USERNAME=postgres
      - DB_PASSWORD=postgres
      - REDIS_HOST=redis:6379
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=goilerplate
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

---

## 🛠️ Common Commands

### Start Development Environment

```bash
# Start all services
make up
docker compose up -d

# View logs
docker compose logs -f app
docker compose logs -f postgres
docker compose logs -f redis

# Stop all services
docker compose down

# Stop and remove volumes (⚠️ deletes data)
docker compose down -v

# Rebuild services
docker compose up -d --build
```

### Access Services

**PostgreSQL:**
```bash
# From host
psql -h localhost -U postgres -d goilerplate

# From app container
docker compose exec app psql -h postgres -U postgres -d goilerplate
```

**Redis:**
```bash
docker compose exec redis redis-cli
```

**App Shell:**
```bash
docker compose exec app bash
```

---

## 🏗️ Production Build

### Build Production Image

```bash
# Using Makefile
make docker-build

# Using Docker directly
docker build -t goilerplate:latest .
```

Uses `Dockerfile` (optimized production build, no hot reload).

### Run Production Container

```bash
# Using Makefile
make docker-run

# Using Docker directly
docker run -p 3000:3000 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-password \
  -e REDIS_HOST=your-redis-host \
  goilerplate:latest
```

### Production-Grade Setup

For production, use:

1. **Separate database service** (managed DB like GCP Cloud SQL)
2. **Separate cache service** (managed Redis like Google MemoryStore)
3. **Container registry** (Google Artifact Registry, Docker Hub, etc)
4. **Kubernetes** or orchestration service

See [Kubernetes Deployment](../deployment/kubernetes.md) for details.

---

## 📦 Dockerfiles

### Dockerfile (Production)

```dockerfile
# Build stage
FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY config/config.example.yaml config/
EXPOSE 3000
CMD ["./server"]
```

**Features:**
- Multi-stage build (small & optimized)
- Alpine base image (minimal size)
- Non-root user (security best practice)

### Dockerfile.local (Development with Hot Reload)

```dockerfile
FROM golang:1.24-alpine
WORKDIR /app

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 3000

CMD ["air"]
```

**Features:**
- Volume mounting for source code
- Air for auto-reload
- Development tools included

---

## 🔐 Security Best Practices

✅ **DO:**

- Use official base images (golang, alpine, postgres)
- Multi-stage builds for production (reduce image size)
- Non-root user in container
- Environment variables for secrets
- Scan images for vulnerabilities

❌ **DON'T:**

- Hardcode secrets in Dockerfile
- Use `:latest` tag in production
- Run as root user
- Mount production volumes unnecessarily
- Ship development tools in production image

---

## 🚀 Docker Commands Cheat Sheet

```bash
# Build
docker build -t goilerplate:1.0 .

# Run
docker run -p 3000:3000 goilerplate:1.0

# List images
docker images

# Remove image
docker rmi goilerplate:1.0

# Compose commands
docker compose up -d
docker compose down
docker compose logs -f
docker compose exec app bash
docker compose ps
```

---

## 🐛 Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -i :3000

# Kill process
kill -9 <PID>

# Or use different port
docker run -p 3001:3000 goilerplate
```

### Volume Permission Issues (Mac/Windows)

```bash
# Rebuild with fresh volumes
docker compose down -v
docker compose up -d --build
```

### Database Migration Failed

```bash
# Run migrations manually
docker compose exec app go run cmd/migrate/main.go

# Check logs
docker compose logs postgres
```

---

## 🔗 Related

- [Installation Guide](./installation.md) - Quick start without Docker
- [Development Guide](./development.md) - Development workflow
- [Kubernetes Deployment](../deployment/kubernetes.md) - Production deployment
