# Run tests for changed packages

Run Go tests targeted at packages that changed in the current branch.

1. Run `git diff main...HEAD --name-only` to get the list of changed files.
2. Extract unique package paths from the changed files (strip filename, keep directory).
3. If no changed files are found, run `make test` for the full suite.
4. Otherwise, build a targeted test command:
   - Convert each directory to a Go package path relative to the module root (e.g. `internal/application/bar/...`)
   - Run: `go test -v <pkg1> <pkg2> ...`
5. If any tests fail, show the failing test names and error output.
6. Report: packages tested, pass/fail count, and any failures.
