# Agent Instructions for `internal/`

## Purpose

Private application code for the CLI. Keep boundaries explicit and easy to test.

## Package Responsibilities

- `cli/`: runtime glue, shared command helpers, downloads, error/client construction helpers.
- `env/`: environment configuration such as `FIGMA_ACCESS_TOKEN`.
- `figma/`: Figma domain/API boundary: input parsing, node IDs, URL builders, HTTP client, typed API responses, export/tokens API helpers.
- `extract/`: transforms Figma document trees into CLI output models.

## Dependency Rules

- `internal/*` must not import `cmd`.
- `extract` must remain mostly pure: no env vars, Cobra, stdout/stderr, or network calls.
- `figma` may depend on generated `internal/figma/api` types.
- `cli` may wire env and figma clients, but avoid hiding business logic there.

## Testing

- Add table-driven tests for parsing, traversal, formatting, and error paths.
- Use deterministic fixtures/maps; avoid live Figma API calls.
- Run `go test ./internal/...` for focused validation, then `go test ./...` before finalizing.
