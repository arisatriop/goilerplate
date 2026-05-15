#!/usr/bin/env bash
# Authenticate the gh CLI non-interactively using the token in config/.env.
# Safe to run repeatedly and in non-interactive shells — exits early if gh is
# already authenticated. Run this before any command that uses gh.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

if gh auth status >/dev/null 2>&1; then
  exit 0
fi

ENV_FILE="config/.env"
if [ ! -f "$ENV_FILE" ]; then
  echo "ERROR: $ENV_FILE not found. Run this from the repo root." >&2
  exit 1
fi

TOKEN=$(grep -E '^GITHUB_PERSONAL_ACCESS_TOKEN=' "$ENV_FILE" | cut -d= -f2- || true)
if [ -z "${TOKEN:-}" ]; then
  echo "ERROR: GITHUB_PERSONAL_ACCESS_TOKEN missing or empty in $ENV_FILE." >&2
  echo "Add a GitHub personal access token to $ENV_FILE to authenticate gh." >&2
  exit 1
fi

printf '%s' "$TOKEN" | gh auth login --with-token
echo "gh CLI authenticated using the token from $ENV_FILE."
