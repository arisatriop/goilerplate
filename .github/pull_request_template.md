## Summary

<!-- 1-3 bullet points: what changed and why -->
-

## Related ticket

<!-- Jira ticket ID / link, or "n/a" -->

## Type of change

- [ ] `feat` — new feature
- [ ] `fix` — bug fix
- [ ] `chore` / `refactor` / `docs`
- [ ] `perf` / `test`

## Test plan

- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] <!-- specific scenarios relevant to this change -->

## Checklist

- [ ] Respects Clean Architecture boundaries (`domain` ← `application` ← `delivery`)
- [ ] No `float64` for financial values — `shopspring/decimal` used
- [ ] DB access only through the repository interface — no GORM in handlers/use cases
- [ ] New migration has matching `.up.sql` and `.down.sql` (if applicable)
- [ ] No secrets committed — configuration comes from `config/.env`
