# 📚 Goilerplate Documentation

Complete documentation for Goilerplate - Production-ready Go backend boilerplate with authentication, authorization, and best practices.

← [Back to Project Overview](../README.md)

---

## 🚀 Getting Started

Guides to start development and setup the project.

- **[Installation & Setup](./getting-started/installation.md)** - How to install and initial configuration
- **[Docker Setup](./getting-started/docker.md)** - Development with Docker & Docker Compose
- **[Development Guide](./getting-started/development.md)** - Local development workflow

---

## 📖 Guides & How-To

In-depth guides for features and how the system works.

- **[CRUD Operations](./guides/crud-operations.md)** - Step-by-step guide to create new CRUD
- **[gRPC Guide](./guides/grpc.md)** - Proto setup, code generation, adding services, testing with grpcurl
- **[Observability Guide](./guides/observability.md)** - OpenTelemetry tracing setup and configuration
- **[Router & API Routes](./api/router.md)** - Route structure (Public, Partner, Internal)
- **[Architecture](./guides/architecture.md)** - Clean architecture & layer explanation

---

## 🚢 Deployment

Guides for deploying to different environments.

- **[Configuration Guide](./deployment/configuration.md)** - Setup environment variables & configuration
- **[Kubernetes Deployment](./deployment/kubernetes.md)** - Deploy to GKE with ConfigMap & Secret
- **[CI/CD Pipeline](./deployment/ci-cd.md)** - GitHub Actions workflow setup

---

## 🤖 Claude Code Integration

Slash commands for AI-assisted development workflows.

- **[MCP Setup](./reference/mcp-setup.md)** — PostgreSQL, GitHub, fetch, and Jira MCP server configuration

Key commands (defined in `.claude/commands/`):

| Command | Description |
|---|---|
| `/ship [TICKET_ID]` | Full workflow: Jira → implement → commit → PR → done |
| `/work-on [TICKET_ID]` | Fetch a Jira ticket and implement it |
| `/commit` / `/commit-body` | Create conventional commits |
| `/push` / `/pr` | Push branch and open a GitHub PR |
| `/mark-as-start` / `/mark-as-done` | Post Jira comments |
| `/next-transition [TICKET_ID]` | Move Jira ticket to next status |
| `/add-domain <name>` | Scaffold a new Clean Architecture domain |
| `/code-review` / `/pr-review` | Review code or a GitHub PR |

See [CLAUDE.md](../CLAUDE.md#claude-commands) for the full command reference.

---

## 🔧 Reference

Technical documentation and special setup.

- **[MCP Setup](./reference/mcp-setup.md)** — PostgreSQL, GitHub, fetch, and Jira MCP server configuration

---

## 📡 API Documentation

- Endpoint authentication at **[Router & API Routes](./api/router.md)**
- Request/response examples at **[README root](../README.md#-api-endpoints)**

---

## ✨ Quick Links

| Topic | Link |
|-------|------|
| Main README | [README.md](../README.md) |
| gRPC Guide | [guides/grpc.md](./guides/grpc.md) |
| Observability | [guides/observability.md](./guides/observability.md) |
| Project Structure | [README.md#-project-structure](../README.md#-project-structure) |
| Tech Stack | [README.md#-tech-stack](../README.md#-tech-stack) |
| Contributing | [README.md#-contributing](../README.md#-contributing) |

---

For more information, see [README.md](../README.md) in the root project.
