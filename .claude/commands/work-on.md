# Implement Ticket — Read and implement a Jira ticket's instructions

Fetch a Jira ticket's description and implement the work it describes end-to-end.

Usage:
- `/implement-ticket TICKET_ID` — implement that ticket directly
- `/implement-ticket` — fetch tickets assigned to you, then pick one

---

## Step 1 — Read credentials

```bash
grep -E "^JIRA_URL=" config/.env | cut -d= -f2-
grep -E "^JIRA_EMAIL=" config/.env | cut -d= -f2-
grep -E "^JIRA_API_TOKEN=" config/.env | cut -d= -f2-
```

If any of the three are missing or empty, stop and tell the user which vars to add to `config/.env`.

---

## Step 2 — Pick a ticket (if no TICKET_ID given)

Fetch tickets assigned to the current user:
```bash
curl -s -G "$JIRA_URL/rest/api/3/search/jql" \
  -u "$JIRA_EMAIL:$JIRA_API_TOKEN" \
  --data-urlencode "jql=assignee=currentUser() AND statusCategory != Done ORDER BY updated DESC" \
  --data-urlencode "fields=summary,status,priority" \
  --data-urlencode "maxResults=20"
```

Display as a numbered list:
```
#   TICKET-ID  [Status]       Priority — Summary
1.  PROJ-42    [To Do]        Medium   — Add user profile endpoint
2.  PROJ-38    [In Progress]  High     — Fix auth token expiry
```

Ask: "Which ticket do you want to implement? (enter number or ticket ID)". Wait for answer.

---

## Step 3 — Fetch and display ticket details

```bash
curl -s "$JIRA_URL/rest/api/3/issue/$TICKET_ID?fields=summary,description,issuetype,priority,status" \
  -u "$JIRA_EMAIL:$JIRA_API_TOKEN" \
  -o /tmp/jira_ticket.json
```

Extract readable text from the ADF `description` field using Python:
```python
import json

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
        return "".join("• " + adf_to_text(c) for c in node.get("content", []))
    if t == "orderedList":
        return "".join(f"{i+1}. {adf_to_text(c)}" for i, c in enumerate(node.get("content", [])))
    if t == "heading":
        level = node.get("attrs", {}).get("level", 2)
        text = "".join(adf_to_text(c) for c in node.get("content", []))
        return "#" * level + " " + text + "\n"
    if t == "codeBlock":
        code = "".join(adf_to_text(c) for c in node.get("content", []))
        return f"```\n{code}\n```\n"
    if t == "hardBreak":
        return "\n"
    if t == "rule":
        return "---\n"
    return "".join(adf_to_text(c) for c in node.get("content", []))

with open("/tmp/jira_ticket.json") as f:
    data = json.load(f)

fields = data.get("fields", {})
print("SUMMARY:", fields.get("summary", ""))
print("TYPE   :", fields.get("issuetype", {}).get("name", ""))
print("STATUS :", fields.get("status", {}).get("name", ""))
print()
print("DESCRIPTION:")
print(adf_to_text(fields.get("description") or {}))
```

Show the output to the user. **Stop and confirm**: "This is what I'll implement. Proceed? (yes/no)". Wait for confirmation before continuing.

---

## Step 4 — Create a feature branch

Derive the branch name from the ticket ID:
- Default: `feat/<ticket-id-lowercase>` (e.g. `feat/is-331`)
- Bug fix tickets (issuetype = Bug): `fix/<ticket-id-lowercase>`

```bash
git checkout -b feat/<ticket-id-lowercase>
```

---

## Step 5 — Analyze and implement

Read the parsed description carefully and determine what needs to be done. Use the following decision tree:

### New domain / new entity / new CRUD API
Follow the `/add-domain` pattern exactly:
- Read the `bar` reference files as defined in `add-domain.md` before writing anything.
- Create all domain, infrastructure, delivery, and wire files.
- Generate the migration with `make migrate-create name=create_<names>_table`.
- Populate SQL up/down migration based on the fields described in the ticket.

### New standalone API endpoint on an existing domain
- Read the existing handler, usecase, and repository for that domain first.
- Add the new method to the usecase interface, implement it in the usecase struct, add the repository method, and register the route.

### Bug fix or adjustment
- Locate the affected files across `domain/`, `application/`, `infrastructure/repository/`, and `delivery/http/handler/`.
- Identify the root cause, not just the symptom.
- Make the minimal change necessary.

### Data migration / data ingestion
- Create a SQL migration file using `make migrate-create name=<description>`.
- Populate the migration with the exact SQL described or derived from the ticket.

### General rules (always apply)
- Follow Clean Architecture boundaries: `domain` ← `application` ← `delivery`; never import inward layers outward.
- Never use `float64` for financial values — use `shopspring/decimal`.
- All DB access through the repository interface — no GORM calls in handlers or use cases.
- Use `go-playground/validator` for request validation in DTOs.
- Follow the response envelope format via `pkg` response helpers.
- Match HTTP status codes to API conventions (`201` for create, `200` for read/update, `400` for bad input, `404` for not found, etc.).

---

## Step 6 — Verify

```bash
go build ./...
```

Fix any compile errors before reporting. Do not run migrations — leave that to the user.

---

## Step 7 — Report

Summarize:
- What was implemented (files created/modified with paths)
- Branch name
- Any assumptions made where the ticket description was ambiguous
- Next steps for the user (e.g. run migrations, adjust field names, add tests)

---

## Cleanup

```bash
rm -f /tmp/jira_ticket.json
```
