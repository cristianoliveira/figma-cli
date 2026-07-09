# Agent Instructions for `cmd/`

## Purpose

`cmd/` contains Cobra command definitions and the CLI entry wiring.

## Rules

- Keep command files thin: flags/args, input parsing, client loading, package calls, JSON printing.
- Do not put document traversal or extraction logic here; move it to `internal/extract`.
- Do not put HTTP or Figma URL construction details here unless command-specific glue is unavoidable; prefer `internal/figma`.
- Use `cli.Die(err)` for command failures so error formatting stays consistent.
- Register commands in `init()` with `rootCmd.AddCommand(...)`.
- Keep command output JSON and stable for scripts/agents.

## Testing

- Prefer testing logic in `internal/*` packages instead of Cobra command closures.
- If adding command-specific behavior, extract it into a small testable function or package helper.
- Run `go test ./...` after command changes.
