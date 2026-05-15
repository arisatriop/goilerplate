---
description: Review an open GitHub PR — auto-routes to inline or the go-reviewer subagent, then posts
argument-hint: [PR_NUMBER] [inline|agent]
---

# Review an open pull request on GitHub

Review an open pull request using the gh CLI, then post the review.

`$ARGUMENTS` may contain a PR number and/or a mode override (`inline` / `agent`),
in any order. Mode override forces the review path; if omitted it is decided
automatically from the PR size.

## Preflight — GitHub auth

Ensure the gh CLI is authenticated (non-interactive — reads the token from `config/.env`):
```bash
.claude/scripts/gh-login.sh
```
If it reports a missing or invalid token, stop and tell the user to set a valid `GITHUB_PERSONAL_ACCESS_TOKEN` (scopes `repo` + `read:org`) in `config/.env`.

## 1. Select the PR

If no PR number was given, list open PRs:
```bash
gh pr list --state open --json number,title,author,url \
  --jq '.[] | "[\(.number)] \(.title) by \(.author.login) — \(.url)"'
```
- No open PRs → tell the user and stop.
- Exactly one → use it.
- More than one → show them and ask which to review.

## 2. Measure the PR

```bash
gh pr view <number> --json title,body,headRefOid,additions,deletions,changedFiles
```
Note `changedFiles` and `additions + deletions`.

## 3. Pick the mode

- `$ARGUMENTS` contains `inline` → inline review (step 4a).
- `$ARGUMENTS` contains `agent` → delegated review (step 4b).
- otherwise → **auto**: if `changedFiles` > **10** OR `additions + deletions` >
  **400**, use delegated (4b); otherwise inline (4a).

State which mode was chosen and why.

## 4a. Inline review

Read the review criteria from `.claude/agents/go-reviewer.md`. Fetch the diff
with `gh pr diff <number>`, read the changed files for context, and review it
yourself against the criteria.

## 4b. Delegated review

Use the Agent tool with `subagent_type: go-reviewer`. Instruct it to review pull
request `#<number>` (it will fetch the diff and read files in its own context).
Use the findings report it returns.

## 5. Post the review

Build the review from the findings (whichever path produced them). Get the head
SHA and repo:
```bash
COMMIT=$(gh pr view <number> --json headRefOid --jq '.headRefOid')
REPO=$(gh repo view --json nameWithOwner --jq '.nameWithOwner')
```
Write the payload to a file, then post in a single call:
```bash
python3 -c "
import json, sys
json.dump(json.loads(sys.argv[1]), open('/tmp/gh_review_payload.json', 'w'))
" "$REVIEW_JSON"

gh api repos/$REPO/pulls/<number>/reviews --method POST --input /tmp/gh_review_payload.json
```
Payload structure:
```json
{
  "commit_id": "<COMMIT>",
  "body": "## Summary\n\n**Critical**: ...\n**Warning**: ...\n**Suggestion**: ...",
  "event": "REQUEST_CHANGES",
  "comments": [
    {"path": "path/to/file.go", "line": 42, "body": "inline comment"}
  ]
}
```
- `event`: `APPROVE`, `REQUEST_CHANGES`, or `COMMENT`.
- Omit `comments` if there are no inline remarks.
- Be constructive — explain *why* and suggest a fix; distinguish blocking issues from suggestions.

## 6. Finish

```bash
rm -f /tmp/gh_review_payload.json
```
Output the PR URL and a brief recap of the review.
