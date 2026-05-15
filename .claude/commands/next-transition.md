---
description: Move a Jira ticket to its next workflow status
argument-hint: [TICKET_ID] [TRANSITION]
allowed-tools: Bash
---

# Jira Transition — Move Ticket to Next Status

Move a Jira ticket to a new status by executing an available transition.

Usage:
- `/next-transition TICKET_ID TRANSITION` — move directly to that transition (name or list number)
- `/next-transition TICKET_ID` — auto-select the first available transition
- `/next-transition` — list tickets assigned to you, then auto-transition the chosen one

Jira credentials (`JIRA_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`) are read from
`config/.env` by the helper scripts in `.claude/scripts/`.

## 1. Resolve the ticket

If a TICKET_ID was given, use it. Otherwise list assigned tickets and ask which one:
```bash
.claude/scripts/jira-tickets.sh
```
Ask: "Which ticket do you want to transition? (enter number or ticket ID)".

## 2. (Optional) Inspect available transitions

If you want to show the user the options first, or no TRANSITION was given and
you want to confirm the target:
```bash
.claude/scripts/jira-transition.sh <TICKET_ID>
```
This lists each transition as `N. <name> -> <target status> (id <id>)`.

## 3. Execute the transition

```bash
# Specific transition (by name or list number):
.claude/scripts/jira-transition.sh <TICKET_ID> "<TRANSITION>"

# Or auto-select the first available transition:
.claude/scripts/jira-transition.sh <TICKET_ID> first
```

The script prints `Ticket <TICKET_ID> moved to <new status>.` on success or a
clear error otherwise — relay its output.
