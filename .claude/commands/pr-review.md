---
description: Review an open GitHub pull request and post the review via the gh CLI
argument-hint: [PR_NUMBER]
---

# Review an open pull request on GitHub

Review an open pull request on GitHub using the gh CLI.

1. List open PRs (only the fields needed):
   ```bash
   gh pr list --state open --json number,title,author,url \
     --jq '.[] | "[\(.number)] \(.title) by \(.author.login) — \(.url)"'
   ```
2. If **no open PRs** are found, inform the user and stop.
3. If **exactly one PR** is open, proceed to review it automatically.
4. If **more than one PR** is open, display them and ask the user which one to review before proceeding.
5. Fetch only the fields needed:
   ```bash
   gh pr view <number> --json number,title,body,state,headRefName
   gh pr diff <number>
   ```
6. Review the PR thoroughly:
   - Understand the intent from the PR title and description
   - Read through all changed files and the diff
   - Check against the criteria below
7. Post inline comments and the summary in a single API call. Get the head commit SHA and repo name:
   ```bash
   COMMIT=$(gh pr view <number> --json headRefOid --jq '.headRefOid')
   REPO=$(gh repo view --json nameWithOwner --jq '.nameWithOwner')
   ```
   Write the review payload to a file, then post:
   ```bash
   python3 -c "
   import json, sys
   payload = json.loads(sys.argv[1])
   json.dump(payload, open('/tmp/gh_review_payload.json', 'w'))
   " "$REVIEW_JSON"

   gh api repos/$REPO/pulls/<number>/reviews \
     --method POST --input /tmp/gh_review_payload.json
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
   - `event`: `"APPROVE"`, `"REQUEST_CHANGES"`, or `"COMMENT"`
   - Omit `comments` if there are no inline remarks
   - Be constructive — explain *why* something is an issue and suggest a fix
   - Distinguish **blocking issues** (must fix) from **suggestions** (nice to have)
8. Clean up: `rm -f /tmp/gh_review_payload.json`
9. Output the PR URL and a brief recap of the review when done.

---

## Review Criteria

### Correctness
- Does the logic work as intended? Any off-by-one errors, nil pointer risks, or unhandled errors?
- Are all returned errors checked? No `_` discards on error returns.

### Clean Architecture Boundaries
- `domain/` has zero external dependencies — no GORM, no Fiber, no infrastructure imports.
- `application/` depends only on `domain/` interfaces — never on `infrastructure/` directly.
- `infrastructure/` implements `domain/` interfaces — never imported by `application/` or `delivery/`.
- `delivery/` depends on `domain/` Usecase interfaces only.
- `wire/` is the only place that wires all concrete implementations.
- No GORM calls in handlers or use cases — all DB access via repository interfaces.

### Financial Safety
- All monetary/financial values must use `github.com/shopspring/decimal` — never `float64`.
- User input for financial fields must be parsed with `decimal.NewFromString()`.

### Security
- No hardcoded credentials, tokens, or secrets — secrets come from `config/.env` via Viper.
- No SQL injection risk — raw queries must use parameterised placeholders.
- All input validated at the handler layer using `go-playground/validator` struct tags before passing to use cases.
- Internal routes protected by `middleware.Auth`; partner routes by `middleware.APIKey`.
- Auth data accessed via `ctx.Locals("user")` — not raw header parsing.

### Conventions
- Files: `snake_case.go`. Exported types/functions: `PascalCase`. Unexported: `camelCase`.
- Domain entities in `domain/<name>/entity.go` — no GORM tags.
- GORM models in `infrastructure/model/` — separate from domain entities.
- Request DTOs in `delivery/http/dto/request/`, response DTOs in `delivery/http/dto/response/` — DB models never exposed directly.
- New domains follow the pattern: `domain/<name>/`, `application/<name>/`, `infrastructure/repository/<name>.go`, `delivery/http/handler/<name>.go`, wired in `internal/wire/`.
- Migration files generated via `make migrate-create`, placed in `internal/migrations/`.
- Commit messages follow conventional commits: `<type>(<scope>): <description>`.

### Error Handling
- Errors always wrapped with context: `fmt.Errorf("doing X: %w", err)`.
- Errors returned up to the handler layer — not logged and swallowed mid-stack.
- Domain-defined errors (`domain/<pkg>/error.go`) used for expected error cases.

### Fiber & API Conventions
- Responses use the standard envelope via the `pkg` response helper (`success`, `message`, `data`).
- Paginated responses include `meta` with `page`, `limit`, `total`.
- Correct HTTP status codes: 200 (GET/PUT/PATCH), 201 (POST created), 400 (validation), 401 (no/invalid token), 403 (forbidden), 404 (not found), 500 (server error).
- Request body parsed with `ctx.BodyParser(&req)`.
- Routes registered in the correct file: `router/public.go`, `router/internal.go`, or `router/partner.go`.

### Tests
- New business logic in `application/` must have at least a happy-path + one error-path test.
- Financial calculations must include zero, negative, large-number, and decimal-precision cases.
- No DB mocking — integration tests use a real PostgreSQL test DB.
