# Pre-commit Hooks

This project uses [lefthook](https://github.com/evilmartians/lefthook) to manage Git hooks that automatically check code quality before commits.

## Installation

Run the following command to install lefthook and set up Git hooks:

```bash
make install-hooks
```

This will:
1. Install lefthook globally via `go install` (if not already installed)
2. Run `lefthook install` to set up the Git hooks

If you use Nix, lefthook is already included in the dev shell (`nix develop`).

## What Hooks Run

On every `git commit`, the following checks run automatically:

### 1. Go Formatting (`gofmt`)
- Runs `go fmt` on all staged Go files
- Automatically stages fixed files

### 2. Import Fixing (`goimports`)
- Runs `goimports -w` on all staged Go files
- Organizes imports and removes unused ones
- Automatically stages fixed files

### 3. Linting (`golangci-lint`)
- Runs `golangci-lint run` using the project's `.golangci.yml` configuration
- Fails if any linting errors are found
- See [golangci-lint](https://golangci-lint.run/) for details

### 4. Quick Tests (`go test`)
- Runs `go test ./... -short` to ensure tests pass
- Fails if any test fails

### 5. Large File Check
- Checks that no staged file exceeds 1MB
- Prevents accidentally committing large binary files

### 6. Whitespace Check
- Checks for trailing whitespace in staged changes
- Fails if whitespace issues are found

### 7. TODO/Warning Check
- Warns about `TODO`, `FIXME`, or `XXX` comments in staged changes
- Does not fail the commit (warning only)
- Suggests filing an issue with `bd create`

### 8. Beads Sync Hook (Original Hook)
- The existing `bd` (beads) pre‑commit hook still runs after the above checks
- It ensures beads issue tracking data is synchronized before commits

## Bypassing Hooks

In rare cases, you may need to bypass the hooks (e.g., for an emergency fix). Use:

```bash
git commit --no-verify
```

⚠️ **Warning**: Only bypass hooks when absolutely necessary. Bypassing means code quality checks are skipped, which could introduce bugs or style violations.

## Troubleshooting

### Hooks are not running
- Verify that `lefthook install` succeeded (check `.git/hooks/pre-commit`)
- Ensure lefthook is in your PATH (`which lefthook`)

### Linting errors
- Run `make lint` to see all lint issues
- Many issues can be auto‑fixed with `golangci-lint run --fix`

### Import errors after `goimports`
- Sometimes `goimports` may incorrectly remove needed imports
- Review the staged changes before committing (`git diff --cached`)

### Conflicts with existing hooks
- Lefthook automatically backs up any existing hooks (e.g., beads) as `.old`
- Both sets of hooks run (lefthook first, then the original)

## Manual Hook Execution

You can run the hooks manually without committing:

```bash
lefthook run pre-commit
```

This is useful to check your staged changes before committing.

## Configuration

The hook configuration is in [`lefthook.yml`](../lefthook.yml) at the project root. You can adjust commands, skip patterns, or add new hooks there.

## Adding New Hooks

To add a new pre‑commit check:

1. Edit `lefthook.yml`
2. Add a new command under the `pre-commit.commands` section
3. Run `lefthook install` to apply changes

See the [lefthook documentation](https://github.com/evilmartians/lefthook/blob/master/docs/configuration.md) for advanced options.

## Notes for Contributors

- The hooks are meant to be **non‑blocking** during the initial rollout (warnings only)
- After a team adaptation period, hooks may be made mandatory
- Always run `make install-hooks` after pulling new changes to ensure your hooks are up‑to‑date