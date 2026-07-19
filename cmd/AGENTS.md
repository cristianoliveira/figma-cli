# Agent Instructions for `cmd/`

## Purpose

`cmd/` contains Cobra command definitions and the CLI entry wiring.

## Rules

- Keep command files thin: flags/args, input parsing, client loading, package calls, JSON printing.
- Do not put document traversal or extraction logic here; move it to `internal/extract`.
- Do not put HTTP or Figma URL construction details here unless command-specific glue is unavoidable; prefer `internal/figma`.
- Return errors from Cobra `RunE`; root execution owns structured stdout error rendering and exit behavior. Stderr is for progress/debug diagnostics and deliberate result-bearing gate diagnoses.
- Register commands in `init()` with `rootCmd.AddCommand(...)`.
- Keep structured domain fields stable. TOON is default; global `--json` preserves compatibility JSON.
- Parse file URLs once with `figma.ParseInput`; resolve URL node scope and optional `--id` through shared `figma` helpers.
- Use `figma.FetchNodeDocuments` when traversal must be restricted to requested subtrees; do not fetch whole file and manually guess selected node.
- State whether command accepts one or many node IDs. Reject unsupported multiple IDs instead of using first silently.
- Render structured values through `cli.NewPrinter(cmd).Structured(...)`; global `--json` selects compatibility output. Keep intentional text/file artifacts on `Text`/`File`.

## Testing

- Prefer testing logic in `internal/*` packages instead of Cobra command closures.
- If adding command-specific behavior, extract it into a small testable function or package helper.
- Run `go test ./...` after command changes.
