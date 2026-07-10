# Agent Instructions for `cmd/`

## Purpose

`cmd/` contains Cobra command definitions and the CLI entry wiring.

## Rules

- Keep command files thin: flags/args, input parsing, client loading, package calls, JSON printing.
- Do not put document traversal or extraction logic here; move it to `internal/extract`.
- Do not put HTTP or Figma URL construction details here unless command-specific glue is unavoidable; prefer `internal/figma`.
- Return errors from Cobra `RunE`; root execution owns consistent stderr formatting and exit behavior.
- Register commands in `init()` with `rootCmd.AddCommand(...)`.
- Keep command output JSON and stable for scripts/agents.
- Parse file URLs once with `figma.ParseInput`; resolve URL node scope and optional `--id` through shared `figma` helpers.
- Use `figma.FetchNodeDocuments` when traversal must be restricted to requested subtrees; do not fetch whole file and manually guess selected node.
- State whether command accepts one or many node IDs. Reject unsupported multiple IDs instead of using first silently.
- Render through `cli.NewPrinter(cmd)` so global `--json` behavior remains consistent.

## Testing

- Prefer testing logic in `internal/*` packages instead of Cobra command closures.
- If adding command-specific behavior, extract it into a small testable function or package helper.
- Run `go test ./...` after command changes.
