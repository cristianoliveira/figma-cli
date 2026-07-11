# Agent Instructions for `internal/`

## Purpose

Private application code for the CLI. Keep boundaries explicit and easy to test.

## Package Responsibilities

- `cli/`: runtime glue only: configured client/printer construction and exit-code errors.
- `env/`: environment configuration such as `FIGMA_ACCESS_TOKEN`.
- `figma/`: Figma domain/API boundary: input parsing, node IDs, URL builders, HTTP client, typed API responses, export/tokens API helpers.
- `extract/`: pure transforms from Figma document trees to CLI output models.
- `assets/`: asset discovery, export, and file-download workflows.
- `comments/`: comment API mapping, retrieval, and node-scoping workflows.
- `diff/`: Figma design-history diff use cases; adapters supply its history interfaces.
- `imagediff/`: generic PNG comparison, masks, overlays, metrics, alignment hints, regions, and classification; no Figma dependencies.
- `pixelperfectcmd/`: standalone Cobra command orchestration for the generic image comparison engine.

## Dependency Rules

- `internal/*` must not import `cmd`.
- `extract` must remain mostly pure: no env vars, Cobra, stdout/stderr, or network calls.
- `figma` and capability adapters may depend on generated `internal/figma/api` types; map them before passing data to `extract`.
- `cli` wires env and Figma clients only; place business workflows in their capability package.

## Testing

- Add table-driven tests for parsing, traversal, formatting, and error paths.
- Use deterministic fixtures/maps; avoid live Figma API calls.
- Run `go test ./internal/...` for focused validation, then `go test ./...` before finalizing.
