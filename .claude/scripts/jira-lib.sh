# Shared Jira helpers — source this from other scripts in .claude/scripts/.
# Not meant to be executed directly.

JIRA_ENV_FILE="${JIRA_ENV_FILE:-config/.env}"

# Load and validate Jira credentials from config/.env into the environment.
jira_load_env() {
  if [ ! -f "$JIRA_ENV_FILE" ]; then
    echo "ERROR: $JIRA_ENV_FILE not found. Run this from the repo root." >&2
    exit 1
  fi
  JIRA_URL=$(grep -E '^JIRA_URL=' "$JIRA_ENV_FILE" | cut -d= -f2- || true)
  JIRA_EMAIL=$(grep -E '^JIRA_EMAIL=' "$JIRA_ENV_FILE" | cut -d= -f2- || true)
  JIRA_API_TOKEN=$(grep -E '^JIRA_API_TOKEN=' "$JIRA_ENV_FILE" | cut -d= -f2- || true)

  local missing=()
  [ -z "${JIRA_URL:-}" ]       && missing+=(JIRA_URL)
  [ -z "${JIRA_EMAIL:-}" ]     && missing+=(JIRA_EMAIL)
  [ -z "${JIRA_API_TOKEN:-}" ] && missing+=(JIRA_API_TOKEN)
  if [ "${#missing[@]}" -gt 0 ]; then
    echo "ERROR: missing or empty in $JIRA_ENV_FILE: ${missing[*]}" >&2
    exit 1
  fi
  JIRA_URL="${JIRA_URL%/}"
}

# jira_curl METHOD PATH [extra curl args...]
# Authenticated request against the Jira REST API v3 (PATH starts with "/").
jira_curl() {
  local method="$1" path="$2"
  shift 2
  curl -s -u "$JIRA_EMAIL:$JIRA_API_TOKEN" -X "$method" \
    "$JIRA_URL/rest/api/3$path" "$@"
}
