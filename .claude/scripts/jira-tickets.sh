#!/usr/bin/env bash
# List Jira tickets assigned to the current user as a Markdown table.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$(git rev-parse --show-toplevel)"
# shellcheck source=jira-lib.sh
source "$SCRIPT_DIR/jira-lib.sh"
jira_load_env

jira_curl GET "/search/jql" -G \
  --data-urlencode "jql=assignee=currentUser() AND statusCategory != Done ORDER BY updated DESC" \
  --data-urlencode "fields=summary,status,priority" \
  --data-urlencode "maxResults=20" \
| python3 -c '
import json, sys
data = json.load(sys.stdin)
issues = data.get("issues", [])
if not issues:
    print("No open tickets assigned to you.")
    sys.exit(0)
print("| # | Ticket ID | Status | Priority | Summary |")
print("|---|-----------|--------|----------|---------|")
for i, it in enumerate(issues, 1):
    f = it.get("fields", {})
    key = it.get("key", "")
    st = (f.get("status") or {}).get("name", "")
    pr = (f.get("priority") or {}).get("name", "")
    sm = (f.get("summary") or "").replace("|", "/")
    print("| %d | %s | %s | %s | %s |" % (i, key, st, pr, sm))
'
