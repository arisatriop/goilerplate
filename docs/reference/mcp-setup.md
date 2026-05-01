# MCP Setup Guide

Model Context Protocol (MCP) setup to allow Claude Code to interact directly with PostgreSQL database.

---

## 🔄 How It Works

MCP uses 3 files for setup:

1. **`.mcp.json`** — Defines MCP servers (committed to git)
2. **`.claude/settings.json`** — Enables project MCP servers automatically (committed to git)
3. **`config/.env`** — Provides DB credentials (gitignored, created locally)

When you open the project in Claude Code, PostgreSQL MCP server starts automatically and Claude can query the database.

---

## 🚀 Setup for New Team Members

### Step 1: Create `.env` File

Create file `config/.env` with your local DB credentials:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goilerplate
DB_USERNAME=postgres
DB_PASSWORD=your_password
```

**⚠️ Important:** This file is git-ignored, do not commit credentials!

### Step 2: Open Project in Claude Code

```bash
claude code
```

MCP will be enabled automatically via `.claude/settings.json`.

### Step 3: Verify Connection

Ask Claude:
> "please try to query the database"

Or use the `/mcp` command to verify connection.

---

## 🔧 Configuration Files

### `.mcp.json` (Committed to Git)

```json
{
  "mcpServers": {
    "postgres": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-postgres"],
      "env": {
        "DATABASE_URL": "postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
      }
    }
  }
}
```

### `.claude/settings.json` (Committed to Git)

```json
{
  "mcp": {
    "mcpServers": ["postgres"]
  }
}
```

---

## 🔌 Reconnecting MCP

If MCP server disconnects:

```
/mcp
```

This command reconnects all MCP servers.

---

## 📊 Available MCP Servers

| Server     | Description                          | Type      |
|-----------|--------------------------------------|-----------|
| `postgres` | Read-only access to local PostgreSQL | Database  |

---

## 🔐 Best Practices

✅ **DO:**
- Keep `config/.env` in `.gitignore`
- Use strong passwords locally
- Verify DB connection before starting development
- Use read-only MCP for development

❌ **DON'T:**
- Commit credentials to git
- Use production DB credentials locally
- Share `.env` file with team (everyone setup their own)
- Use MCP without proper credentials setup

---

## 🐛 Troubleshooting

### MCP Not Connecting

1. Check `.env` file exists and has correct credentials
2. Verify PostgreSQL is running locally
3. Try reconnect with `/mcp`
4. Check logs: `claude logs`

### Database Not Found

1. Verify `DB_NAME` is correct in `.env`
2. Create database if it doesn't exist:
   ```bash
   createdb goilerplate
   ```
3. Run migrations:
   ```bash
   go run cmd/migrate/main.go
   ```

### Permission Denied

1. Verify `DB_USERNAME` and `DB_PASSWORD` are correct
2. Check PostgreSQL user permissions:
   ```bash
   psql -U postgres -l
   ```

---

## 🔗 Related

- [Configuration Guide](../deployment/configuration.md) - Environment setup
- [Development Setup](../getting-started/development.md) - Local development workflow
