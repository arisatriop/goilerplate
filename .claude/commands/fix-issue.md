# Fix Issue

Topic/Issue: $ARGUMENTS

Follow these steps:

1. **Understand the problem** — Read the issue description carefully. Identify what is broken or missing.
2. **Find relevant code** — Locate the affected layer(s): `domain/`, `application/`, `infrastructure/repository/`, `delivery/http/handler/`. Check all relevant files.
3. **Identify root cause** — Trace the bug to its origin. Don't fix symptoms, fix the cause.
4. **Propose the fix** — Describe the change before making it. If the fix touches financial logic, confirm `shopspring/decimal` is used.
5. **Implement the fix** — Make the minimal change necessary. Follow the Clean Architecture layer structure.
6. **Write or update tests** — Add a test that would have caught this bug.
7. **Verify** — Run `go build ./...` and `go test ./...` to confirm nothing is broken.

Branch naming: `fix/<topic>` (e.g. `fix/bar-pagination`)
Commit format: `fix(<scope>): <short description>`
