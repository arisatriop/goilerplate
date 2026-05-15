#!/usr/bin/env bash
# Post a comment to a Jira ticket.
# The comment body is read from stdin as an ADF "doc" JSON object, e.g.:
#   {"version":1,"type":"doc","content":[ ... ]}
# Usage: jira-comment.sh TICKET_ID < adf-body.json
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$(git rev-parse --show-toplevel)"
# shellcheck source=jira-lib.sh
source "$SCRIPT_DIR/jira-lib.sh"
jira_load_env

TICKET_ID="${1:?Usage: jira-comment.sh TICKET_ID < adf-body.json}"

ADF_DOC="$(cat)"
if [ -z "${ADF_DOC//[[:space:]]/}" ]; then
  echo "ERROR: no ADF body received on stdin." >&2
  exit 1
fi

PAYLOAD="$(python3 -c '
import json, sys
doc = json.loads(sys.argv[1])
print(json.dumps({"body": doc}))
' "$ADF_DOC")"

RESPONSE_FILE="$(mktemp -t jira_resp.XXXXXX)"
trap 'rm -f "$RESPONSE_FILE"' EXIT

HTTP="$(jira_curl POST "/issue/$TICKET_ID/comment" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD" \
  -o "$RESPONSE_FILE" -w "%{http_code}")"

case "$HTTP" in
  201) echo "Comment posted to $TICKET_ID." ;;
  401) echo "ERROR: invalid Jira credentials (check JIRA_EMAIL / JIRA_API_TOKEN)." >&2; exit 1 ;;
  404) echo "ERROR: ticket $TICKET_ID not found." >&2; exit 1 ;;
  *)   echo "ERROR: Jira returned HTTP $HTTP" >&2; cat "$RESPONSE_FILE" >&2; echo >&2; exit 1 ;;
esac
