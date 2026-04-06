# MCP Setup Guide for Local Development

This guide explains how to set up Model Context Protocol (MCP) on your local machine to enable Claude Code to query the **goilerplate** database directly.

## Prerequisites

- [Claude Code](https://claude.ai/code) installed (`npm install -g @anthropic-ai/claude-code`)
- Node.js >= 18 (for `npx`)
- Access to the local PostgreSQL database (running on the port configured in `config/.env`)

## Steps

### 1. Copy the environment file

```bash
cp config/.env.example config/.env
```

Then edit `config/.env` and fill in the actual credentials:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=postgres
DB_USERNAME=postgres
DB_PASSWORD=your_actual_password

GITHUB_PERSONAL_ACCESS_TOKEN=ghp_your_token_here
```

> Generate a GitHub token at https://github.com/settings/tokens (needs `repo` scope).

### 2. No changes needed to `.mcp.json`

The `.mcp.json` is already configured to load credentials automatically from `config/.env`:

```json
{
  "mcpServers": {
    "postgres": {
      "command": "sh",
      "args": ["-c", "export $(grep -v '^#' config/.env | xargs) && npx -y @modelcontextprotocol/server-postgres \"postgresql://$DB_USERNAME:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME\""]
    }
  }
}
```

Just make sure your `config/.env` is filled in correctly — no need to touch `.mcp.json`.

### 3. Copy the Claude settings file

```bash
cp .claude/settings.json .claude/settings.local.json
```

The `settings.local.json` file enables the postgres MCP server for your local Claude Code session. It is gitignored so your local settings won't affect others.

### 4. Start Claude Code

```bash
claude
```

To verify the MCP connection is working, run inside Claude Code:

```
/mcp
```

You should see `postgres` listed as a connected server.

### 5. Test the connection

Ask Claude Code:

```
run SELECT version()
```

You should get a response showing the PostgreSQL version.

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `postgres` not listed in `/mcp` | Check `.claude/settings.local.json` has `"postgres"` in `enabledMcpjsonServers` |
| Connection refused | Ensure your local DB is running on the configured port |
| Authentication failed | Re-check your credentials in `config/.env` |
| `npx` not found | Install Node.js >= 18 |

## File Reference

| File | Purpose |
|------|---------|
| `.mcp.json` | MCP server config (committed, loads credentials from `.env`) |
| `.claude/settings.json` | Shared Claude settings (committed) |
| `.claude/settings.local.json` | Your local Claude settings (gitignored) |
| `config/.env.example` | Environment variable template |
| `config/.env` | Your local environment variables (gitignored) |
