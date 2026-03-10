# 🚀 Goilerplate

**Production-ready Go backend boilerplate** with authentication, authorization, and best practices built-in.

A clean, scalable REST API template featuring JWT authentication, role-based access control (RBAC), Redis caching, and clean architecture patterns.

> 📚 **[→ View Full Documentation](./docs/README.md)** — Installation, deployment, architecture, and development guides

---

## ✨ Key Features

- **🔐 Authentication & Authorization** — JWT tokens, RBAC, multi-device sessions
- **🏗️ Clean Architecture** — Layered structure with clear separation of concerns
- **⚡ Performance** — Redis caching, connection pooling, optimized queries
- **📦 Database Agnostic** — PostgreSQL or MySQL support
- **☁️ Cloud Ready** — S3, Google Drive, Kubernetes deployment
- **🧪 Quality First** — Type-safe code, comprehensive validation, error handling

---

## 🛠️ Tech Stack

- **Go 1.24** — Modern Go with generics
- **Fiber v2** — Fast HTTP framework
- **GORM** — Type-safe ORM
- **PostgreSQL/MySQL** — Relational database
- **Redis** — Caching & sessions
- **JWT** — Stateless authentication

---

## 🚀 Quick Start (5 minutes)

### Prerequisites
- Go 1.24+
- PostgreSQL or MySQL
- Redis (optional)

### Setup

```bash
# 1. Clone & install
git clone https://github.com/arisatriop/goilerplate.git
cd goilerplate
go mod download

# 2. Configure
cp config/config.example.yaml config/config.yaml
# Edit config/config.yaml with your database credentials

# 3. Create database
createdb goilerplate

# 4. Run migrations
go run cmd/migrate/main.go

# 5. Start server
go run cmd/server/main.go
```

Server runs at `http://localhost:3000`

**→ [Full setup guide](./docs/README.md)**

---

## 📖 Documentation

All detailed documentation is in the [`/docs`](./docs/README.md) folder:

- **[Getting Started](./docs/getting-started/)** — Installation, Docker, development setup
- **[Guides](./docs/guides/)** — CRUD operations, clean architecture
- **[Deployment](./docs/deployment/)** — Kubernetes, CI/CD, configuration
- **[API Reference](./docs/api/)** — Routes, authentication, permissions

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit with [conventional commits](https://www.conventionalcommits.org/)
4. Push and open a Pull Request

**Commit convention:**
- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation
- `refactor:` — Code refactoring
- `test:` — Tests
- `chore:` — Maintenance

---

## 📄 License

This project is licensed under the MIT License.

---

## 📧 Support

- **Questions?** Open an [issue on GitHub](https://github.com/arisatriop/goilerplate/issues)
- **Full docs:** See [`/docs`](./docs/README.md)
- **Examples:** Check example requests in the documentation

---

Made with ❤️ using Go — Star ⭐ this repo if you find it useful!
