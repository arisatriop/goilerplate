#!/usr/bin/env bash
# List or execute Jira workflow transitions for a ticket.
#   jira-transition.sh TICKET_ID                 -> list available transitions
#   jira-transition.sh TICKET_ID first           -> execute the first available transition
#   jira-transition.sh TICKET_ID "In Progress"   -> execute by transition name (case-insensitive)
#   jira-transition.sh TICKET_ID 2               -> execute by list position
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$(git rev-parse --show-toplevel)"
# shellcheck source=jira-lib.sh
source "$SCRIPT_DIR/jira-lib.sh"
jira_load_env

TICKET_ID="${1:?Usage: jira-transition.sh TICKET_ID [first|<name>|<number>]}"
TRANSITION="${2:-}"

TRANSITIONS="$(jira_curl GET "/issue/$TICKET_ID/transitions")"

if [ -z "$TRANSITION" ]; then
  echo "$TRANSITIONS" | python3 -c '
import json, sys
ts = json.load(sys.stdin).get("transitions", [])
if not ts:
    print("No transitions available for this ticket.")
    sys.exit(0)
for i, t in enumerate(ts, 1):
    print("%d. %s -> %s (id %s)" % (i, t["name"], t["to"]["name"], t["id"]))
'
  exit 0
fi

SEL="$(echo "$TRANSITIONS" | python3 -c '
import json, sys
arg = sys.argv[1].strip().lower()
ts = json.load(sys.stdin).get("transitions", [])
if not ts:
    sys.exit(2)
chosen = None
if arg in ("first", ""):
    chosen = ts[0]
elif arg.isdigit():
    idx = int(arg)
    if 1 <= idx <= len(ts):
        chosen = ts[idx - 1]
    else:
        chosen = next((t for t in ts if t["id"] == arg), None)
else:
    chosen = next((t for t in ts if t["name"].lower() == arg), None)
if not chosen:
    sys.exit(3)
print("%s|%s|%s" % (chosen["id"], chosen["name"], chosen["to"]["name"]))
' "$TRANSITION")" || {
  code=$?
  if [ "$code" -eq 2 ]; then
    echo "ERROR: no transitions available for $TICKET_ID." >&2
  else
    echo "ERROR: transition '$TRANSITION' not found for $TICKET_ID." >&2
  fi
  exit 1
}

TID="${SEL%%|*}"
REST="${SEL#*|}"
TNAME="${REST%%|*}"
TO="${REST##*|}"
echo "Transitioning $TICKET_ID via '$TNAME' -> $TO ..."

RESPONSE_FILE="$(mktemp -t jira_resp.XXXXXX)"
trap 'rm -f "$RESPONSE_FILE"' EXIT

HTTP="$(jira_curl POST "/issue/$TICKET_ID/transitions" \
  -H "Content-Type: application/json" \
  -d "{\"transition\":{\"id\":\"$TID\"}}" \
  -o "$RESPONSE_FILE" -w "%{http_code}")"

case "$HTTP" in
  204) echo "Ticket $TICKET_ID moved to $TO." ;;
  401) echo "ERROR: invalid Jira credentials (check JIRA_EMAIL / JIRA_API_TOKEN)." >&2; exit 1 ;;
  404) echo "ERROR: ticket $TICKET_ID not found or transition unavailable." >&2; exit 1 ;;
  *)   echo "ERROR: Jira returned HTTP $HTTP" >&2; cat "$RESPONSE_FILE" >&2; echo >&2; exit 1 ;;
esac
