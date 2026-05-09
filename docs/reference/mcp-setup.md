# MCP Setup Guide

Model Context Protocol (MCP) setup to allow Claude Code to interact with PostgreSQL, GitHub, the web, and Jira.

---

## How It Works

MCP uses 3 files for setup:

1. **`.mcp.json`** — Defines MCP servers (committed to git)
2. **`.claude/settings.json`** — Enables specific MCP servers automatically (committed to git)
3. **`config/.env`** — Provides credentials (gitignored, created locally)

When you open the project in Claude Code, the enabled MCP servers start automatically.

---

## Available MCP Servers

| Server     | Purpose                                      | Enabled by default |
|------------|----------------------------------------------|--------------------|
| `postgres` | Query local PostgreSQL database              | Yes                |
| `github`   | Read/write GitHub issues, PRs, files         | Yes                |
| `fetch`    | Fetch web URLs (documentation, APIs)         | Yes                |
| `jira`     | Read/write Jira tickets and comments         | No — opt-in below  |

---

## Setup for New Team Members

### Step 1 — Create `config/.env`

Copy from the example and fill in your credentials:

```bash
cp config/.env.example config/.env
```

Minimum required variables:

```env
# PostgreSQL (required for postgres MCP)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=goilerplate
DB_USERNAME=postgres
DB_PASSWORD=your_password

# GitHub (required for github MCP)
GITHUB_PERSONAL_ACCESS_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx

# Jira (optional — only needed for /ship, /work-on, /mark-as-start, /mark-as-done, /next-transition)
JIRA_URL=https://your-org.atlassian.net
JIRA_EMAIL=your@email.com
JIRA_API_TOKEN=your_jira_api_token
```

**Never commit `config/.env` — it is gitignored.**

### Step 2 — Open the Project in Claude Code

```bash
claude
```

`postgres`, `github`, and `fetch` start automatically via `.claude/settings.json`.

### Step 3 — Verify

Ask Claude: "please query the database and show me the tables" — or run `/mcp` to check server status.

---

## Enabling the Jira MCP Server

The `jira` server is defined in `.mcp.json` but not auto-enabled. To activate it, add `"jira"` to `enabledMcpjsonServers` in `.claude/settings.json`:

```json
"enabledMcpjsonServers": ["postgres", "github", "fetch", "jira"]
```

Note: the Jira slash commands (`/ship`, `/work-on`, `/mark-as-start`, etc.) call the Jira REST API directly via `curl` using credentials from `config/.env` — they do not require the Jira MCP server. The MCP server is for ad-hoc interactive queries.

---

## Configuration Files

### `.mcp.json` (committed to git)

```json
{
  "mcpServers": {
    "postgres": {
      "command": "sh",
      "args": ["-c", "export $(grep -v '^#' config/.env | xargs) && npx -y @modelcontextprotocol/server-postgres \"postgresql://$DB_USERNAME:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME\""]
    },
    "github": {
      "command": "sh",
      "args": ["-c", "export $(grep -v '^#' config/.env | xargs) && GITHUB_PERSONAL_ACCESS_TOKEN=$GITHUB_PERSONAL_ACCESS_TOKEN npx -y @modelcontextprotocol/server-github"]
    },
    "fetch": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-fetch"]
    },
    "jira": {
      "command": "sh",
      "args": ["-c", "export $(grep -v '^#' config/.env | xargs) && JIRA_URL=$JIRA_URL JIRA_USERNAME=$JIRA_EMAIL JIRA_API_TOKEN=$JIRA_API_TOKEN npx -y mcp-atlassian"]
    }
  }
}
```

### `.claude/settings.json` (committed to git)

```json
{
  "enabledMcpjsonServers": ["postgres", "github", "fetch"]
}
```

---

## Reconnecting MCP

If a server disconnects, run:

```
/mcp
```

---

## Best Practices

- Keep `config/.env` in `.gitignore` — everyone sets up their own locally
- Use a GitHub PAT scoped to the minimum required permissions (repo read/write)
- Use a Jira API token (not your password) — generate one at `id.atlassian.com`
- Do not use production DB credentials locally

---

## Troubleshooting

### postgres MCP not connecting

1. Confirm `config/.env` has correct `DB_*` values
2. Verify PostgreSQL is running locally
3. Run `/mcp` to reconnect

### github MCP not working

1. Confirm `GITHUB_PERSONAL_ACCESS_TOKEN` is set in `config/.env`
2. Ensure the token has `repo` scope

### Database not found

```bash
createdb goilerplate
go run cmd/migrate/main.go
```

### Permission denied

```bash
psql -U postgres -l
```

---

## Related

- [Configuration Guide](../deployment/configuration.md) — Environment variables
- [Development Setup](../getting-started/development.md) — Local development workflow
- [Claude Commands](../../CLAUDE.md#claude-commands) — Slash commands that use Jira/GitHub
