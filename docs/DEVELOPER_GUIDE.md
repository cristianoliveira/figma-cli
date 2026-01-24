# Developer Guide

This guide covers development practices for the Figma CLI project.

## Getting Started

1. **Clone the repository**
2. **Set up environment**:
   ```bash
   nix develop   # or use direnv
   ```
3. **Install dependencies**:
   ```bash
   go mod download
   ```
4. **Install pre-commit hooks** (recommended):
   ```bash
   make install-hooks
   ```

## Code Quality

### Pre‑commit Hooks

We use [lefthook](https://github.com/evilmartians/lefthook) to enforce code quality checks before commits. The hooks automatically:

- Format Go code with `gofmt` and `goimports`
- Run linting with `golangci-lint`
- Execute quick tests
- Check for large files, whitespace issues, and TODO comments

**Full documentation**: [PRE_COMMIT_HOOKS.md](./PRE_COMMIT_HOOKS.md)

### Running Hooks Manually

You can run the hooks without committing to check your changes:

```bash
lefthook run pre-commit
```

### Bypassing Hooks

In emergencies, use `git commit --no-verify`. Avoid this unless absolutely necessary.

### Linting

Run the full lint suite:

```bash
make lint
```

Many lint issues can be auto‑fixed:

```bash
golangci-lint run --fix
```

### Testing

Run all tests:

```bash
make test
```

Run tests with coverage:

```bash
make test-cover
```

## Project Structure

- `cmd/` – CLI command definitions (Cobra)
- `internal/` – Private application code
- `pkg/` – Public reusable packages
- `scripts/` – Build and utility scripts
- `docs/` – Project documentation

## Adding New Commands

Use Cobra CLI:

```bash
cobra-cli add <command-name>
```

See [Cobra CLI documentation](https://github.com/spf13/cobra-cli) for details.

## Logging

We use [zap](https://github.com/uber-go/zap) for structured logging. See [logging.md](./logging.md) for configuration and usage.

## Issue Tracking

We use [beads](https://github.com/Dicklesworthstone/beads_viewer) for issue tracking.

- View issues: `bv` (TUI) or `bd list`
- Claim work: `bd update <id> --status in_progress`
- Close work: `bd close <id>`
- Sync changes: `bd sync`

See [AGENTS.md](./AGENTS.md) for full agent workflow.

## Agent Development

When working with agents, follow the [Best Practices for Ordering Agents](./best-practices-agent-ordering.md) and [Cobra CLI Output Methods](./cobra-cli-output-methods.md).

## Commit Guidelines

- Write clear, concise commit messages
- Include relevant issue IDs when applicable
- Ensure all hooks pass before pushing
- Use `git pull --rebase` before pushing to avoid merge commits

## Pull Requests

1. Create a feature branch
2. Ensure all tests and lint checks pass
3. Update documentation if needed
4. Create a PR with a clear description of changes and why they're needed

## Need Help?

- Check existing documentation in `docs/`
- Search previous agent reports with `qmd search`
- File an issue with `bd create`