---
description: Review the current git diff — auto-routes to inline or the go-reviewer subagent
argument-hint: [inline|agent]
---

# Code Review

Review the current uncommitted changes (`git diff HEAD`) for correctness,
security, and project conventions.

`$ARGUMENTS` — optional mode override:
- `inline` — force an in-conversation review (no subagent)
- `agent` — force delegation to the `go-reviewer` subagent
- empty — decide automatically from the diff size

## 1. Measure the diff

```bash
git diff HEAD --stat
```
Note the number of files changed and total lines changed.

## 2. Pick the mode

- `$ARGUMENTS` = `inline` → inline review (step 3a).
- `$ARGUMENTS` = `agent` → delegated review (step 3b).
- empty → **auto**: if more than **10 files** OR more than **400 lines** changed,
  use delegated (3b); otherwise inline (3a).

State which mode was chosen and why (e.g. "Auto: 3 files / 60 lines → inline").

## 3a. Inline review

Read the review criteria from `.claude/agents/go-reviewer.md`, then review
`git diff HEAD` against them yourself, in this conversation.

## 3b. Delegated review

Use the Agent tool with `subagent_type: go-reviewer`. Instruct it to review the
current working-tree diff (`git diff HEAD`). The subagent does the file reading
and analysis in its own context and returns a findings report — relay that
report to the user.

## Output

Either way, present findings grouped by severity: **Critical**, **Warning**,
**Suggestion**. This command only reviews — it does not commit or modify code.
