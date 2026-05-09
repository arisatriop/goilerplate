# Push — Push Current Branch to GitHub

Push the current branch to the remote origin.

1. Get the current branch name:
   ```bash
   git branch --show-current
   ```

2. **Safety check** — if the branch is `main` or `master`, stop immediately and tell the user:
   > "Refusing to push: you are on `<branch>`. Switch to a feature branch first."
   Do NOT proceed.

3. Push to origin:
   ```bash
   git push -u origin <branch-name>
   ```

4. Confirm success or report any error output to the user.
