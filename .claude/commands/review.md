# Code Review

Review the current git diff (`git diff HEAD`) for the following:

1. **Correctness** — Does the logic work as intended? Any off-by-one errors, nil pointer risks, or unhandled errors?
2. **Architecture** — Does the code respect Clean Architecture boundaries? No GORM/DB calls in use cases or handlers? No domain importing infrastructure?
3. **Financial safety** — Are all monetary/financial values using `shopspring/decimal`, not `float64`?
4. **Security** — Any hardcoded secrets, SQL injection risks, unvalidated input, or missing auth middleware on routes?
5. **Conventions** — Does it follow the project's domain module structure and naming conventions?
6. **Missing tests** — Is there business logic that should be covered by tests?
7. **Error handling** — Are errors wrapped with context? Are they returned properly, not swallowed?
8. **Fiber specifics** — Are responses using the standard envelope? Correct HTTP status codes? Proper use of `ctx.Locals` for auth data?

Provide a concise summary with findings grouped by severity: **Critical**, **Warning**, **Suggestion**.
