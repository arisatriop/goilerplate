Create a git commit with a detailed message for complex changes by following these steps:

1. Run `git status` and `git diff` (staged and unstaged) in parallel with `git log --oneline -10` to understand the changes and match the repo's commit message style.
2. Analyze all changes and draft a commit message with subject + body:
   - **Subject line**: `<type>(<scope>): <short description>` (e.g. `feat(auth): add refresh token rotation`), under 72 characters
   - Types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`
   - **Blank line** separating subject from body
   - **Body**: explain *why*, not *what*. Focus on motivation, context, or non-obvious decisions. Keep each line under 72 characters.
3. Staging strategy:
   - If there are **already staged changes**, commit only those — do not stage anything else.
   - If there are **no staged changes**, stage all modified/untracked files with `git add -A`.
4. Create the commit. Always append the co-author trailer after the body:
   ```
   Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
   ```
5. Run `git status` to confirm the commit succeeded.

Do not push. Do not ask for confirmation before running git commands — execute them directly.
