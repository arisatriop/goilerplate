#!/usr/bin/env bash
# Fetch a single Jira ticket and print its summary, type, status,
# GitHub repo / base branch custom fields, and description as plain text.
# Usage: jira-ticket.sh TICKET_ID
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$(git rev-parse --show-toplevel)"
# shellcheck source=jira-lib.sh
source "$SCRIPT_DIR/jira-lib.sh"
jira_load_env

TICKET_ID="${1:?Usage: jira-ticket.sh TICKET_ID}"

jira_curl GET "/issue/$TICKET_ID?fields=*navigable&expand=names" \
| python3 -c '
import json, sys

def adf_to_text(node):
    if not node:
        return ""
    t = node.get("type", "")
    if t == "text":
        return node.get("text", "")
    if t in ("doc", "paragraph", "blockquote", "listItem"):
        inner = "".join(adf_to_text(c) for c in node.get("content", []))
        return inner + ("\n" if t in ("paragraph", "listItem") else "")
    if t == "bulletList":
        return "".join("- " + adf_to_text(c) for c in node.get("content", []))
    if t == "orderedList":
        return "".join("%d. %s" % (i + 1, adf_to_text(c))
                       for i, c in enumerate(node.get("content", [])))
    if t == "heading":
        level = node.get("attrs", {}).get("level", 2)
        text = "".join(adf_to_text(c) for c in node.get("content", []))
        return "#" * level + " " + text + "\n"
    if t == "codeBlock":
        code = "".join(adf_to_text(c) for c in node.get("content", []))
        return "```\n" + code + "\n```\n"
    if t == "hardBreak":
        return "\n"
    if t == "rule":
        return "---\n"
    return "".join(adf_to_text(c) for c in node.get("content", []))

data = json.load(sys.stdin)
if "fields" not in data:
    err = data.get("errorMessages") or data.get("errors") or data
    print("ERROR: could not load ticket:", err, file=sys.stderr)
    sys.exit(1)

fields = data.get("fields", {})
names = data.get("names", {})  # field_id -> display_name

github_repo = ""
base_branch = "main"
for fid, fname in names.items():
    val = fields.get(fid)
    if not isinstance(val, str) or not val:
        continue
    if fname.lower() == "github repo":
        github_repo = val
    elif fname.lower() == "base branch":
        base_branch = val

print("SUMMARY    :", fields.get("summary", ""))
print("TYPE       :", (fields.get("issuetype") or {}).get("name", ""))
print("STATUS     :", (fields.get("status") or {}).get("name", ""))
print("GITHUB REPO:", github_repo or "(not set)")
print("BASE BRANCH:", base_branch)
print()
print("DESCRIPTION:")
print(adf_to_text(fields.get("description") or {}))
'
