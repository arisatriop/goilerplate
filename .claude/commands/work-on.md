---
description: Fetch a Jira ticket and implement the work it describes end-to-end
argument-hint: [TICKET_ID]
---

# Implement Ticket — Read and implement a Jira ticket's instructions

Fetch a Jira ticket's description and implement the work it describes end-to-end.

Usage:
- `/work-on TICKET_ID` — implement that ticket directly
- `/work-on` — list tickets assigned to you, then pick one

Jira credentials (`JIRA_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`) are read from
`config/.env` by the helper scripts in `.claude/scripts/`. If a script reports
missing credentials, stop and tell the user which vars to add to `config/.env`.

---

## Step 1 — Pick a ticket (skip if TICKET_ID already given)

```bash
.claude/scripts/jira-tickets.sh
```
Ask: "Which ticket do you want to implement? (enter number or ticket ID)". Wait
for the answer, then set TICKET_ID.

---

## Step 2 — Fetch ticket details

```bash
.claude/scripts/jira-ticket.sh <TICKET_ID>
```
This prints `SUMMARY`, `TYPE`, `STATUS`, `GITHUB REPO`, `BASE BRANCH`, and the
plain-text `DESCRIPTION`. Store the printed `GITHUB REPO` and `BASE BRANCH`
values as `GITHUB_REPO` and `BASE_BRANCH` for later steps. Show the output and
proceed immediately.

---

## Step 3 — Create a feature branch

Derive the branch name from the ticket ID and issue type:
- Default: `feat/<ticket-id-lowercase>` (e.g. `feat/is-331`)
- Bug fix (TYPE = Bug): `fix/<ticket-id-lowercase>`

Check out from `$BASE_BRANCH` (from Step 2):
```bash
git checkout $BASE_BRANCH && git pull origin $BASE_BRANCH
git checkout -b <branch-name>
```

---

## Step 4 — Analyze and implement

Read the parsed description and determine what needs to be done. Use this decision tree:

### New domain / new entity / new CRUD API
Follow the `/add-domain` pattern exactly:
- Read the `bar` reference files listed in `add-domain.md` before writing anything.
- Create all domain, infrastructure, delivery, and wire files.
- Generate the migration with `make migrate-create name=create_<names>_table`.
- Populate the SQL up/down migration based on the fields described in the ticket.

### New standalone API endpoint on an existing domain
- Read the existing handler, usecase, and repository for that domain first.
- Add the new method to the usecase interface, implement it, add the repository method, and register the route.

### Bug fix or adjustment
- Locate the affected files across `domain/`, `application/`, `infrastructure/repository/`, and `delivery/http/handler/`.
- Identify the root cause, not just the symptom. Make the minimal change necessary.

### Data migration / data ingestion
- Create a SQL migration file with `make migrate-create name=<description>`.
- Populate it with the exact SQL described or derived from the ticket.

### General rules (always apply)
- Follow Clean Architecture boundaries: `domain` ← `application` ← `delivery`; never import inward layers outward.
- Never use `float64` for financial values — use `shopspring/decimal`.
- All DB access through the repository interface — no GORM calls in handlers or use cases.
- Use `go-playground/validator` for request validation in DTOs.
- Follow the response envelope format via `pkg` response helpers.
- Match HTTP status codes to API conventions (`201` create, `200` read/update, `400` bad input, `404` not found).

---

## Step 5 — Verify

```bash
go build ./...
```
Fix any compile errors before reporting. Do not run migrations — leave that to the user.

---

## Step 6 — Report

Summarize:
- What was implemented (files created/modified with paths)
- Branch name
- Any assumptions made where the ticket description was ambiguous
- Next steps for the user (e.g. run migrations, adjust field names, add tests)
