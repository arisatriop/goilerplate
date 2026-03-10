# MCP (Model Context Protocol) Setup

This project uses MCP to allow Claude to interact directly with the PostgreSQL database.

## How It Works

- `.mcp.json` — defines the MCP servers (committed to git)
- `.claude/settings.json` — enables all project MCP servers automatically (committed to git)
- `config/.env` — provides the actual DB credentials (gitignored, created locally)

When you open this project in Claude Code, the postgres MCP server will start automatically and Claude can query the database.

## Setup for New Team Members

1. Make sure `config/.env` exists with your local DB credentials:

   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=goilerplate
   DB_USERNAME=postgres
   DB_PASSWORD=your_password
   ```

2. Open the project in Claude Code — MCP will be enabled automatically.

3. To verify the connection, ask Claude: `please try to query the database`

## Reconnecting MCP

If the MCP server disconnects, run `/mcp` inside Claude Code to reconnect.

## Available MCP Servers

| Server | Description |
|--------|-------------|
| `postgres` | Read-only access to the local PostgreSQL database |
