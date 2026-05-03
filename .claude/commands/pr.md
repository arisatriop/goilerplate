# Create a GitHub pull request

Create a pull request for the current branch using the GitHub MCP server.

1. Run these in parallel to understand the current state:
   - `git branch --show-current` to get the current branch name
   - `git log --oneline -10` to review recent commits
2. Base branch is always `main`.
3. Run these in parallel:
   - `git log main...HEAD --oneline` to see all commits on this branch
   - `git diff main...HEAD --stat` to see files changed
4. If there are uncommitted changes, warn the user and ask whether to commit them first before proceeding.
5. Infer the GitHub repo owner and name from `git remote get-url origin`.
6. Draft the PR title and body:
   - **Title**: follow conventional commit format — `<type>(<scope>): <short description>` (e.g. `feat(auth): add refresh token rotation`). Keep it under 70 characters, imperative mood.
   - Types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`
   - **Body**:
     ```
     ## Summary
     <1-3 bullet points describing what changed and why>

     ## Test plan
     - [ ] `make test` passes
     - [ ] `make lint` passes
     - [ ] <specific scenarios relevant to the change>

     🤖 Generated with [Claude Code](https://claude.com/claude-code)
     ```
7. Push the branch if it has no remote tracking branch yet: `git push -u origin <branch>`.
8. Use the GitHub MCP tool to create the PR targeting `main`.
9. Output the PR URL when done.
