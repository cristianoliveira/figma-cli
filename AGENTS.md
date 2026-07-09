# Agent Instructions

## Start Here

- Read this file before changing code.
- For documentation search, read `docs/AGENTS.md` and use `qmd` when available.
- Check nested guidance before editing: `cmd/AGENTS.md`, `internal/AGENTS.md`, and package-specific `AGENTS.md` files.

## Project Shape

This is a Go Cobra CLI for querying Figma and printing agent-friendly JSON.

- Entry point: `cmd/figma/main.go`
- Cobra commands: `cmd/`
- Runtime/env wiring: `internal/cli/`, `internal/env/`
- Figma API boundary: `internal/figma/`
- Pure extraction/transforms: `internal/extract/`
- Generated API models: `internal/figma/api/`
- OpenAPI source/config: `openapi/`

## Architecture Rules

- Keep commands thin: parse flags, resolve input, call Figma/extract packages, print JSON.
- Put Figma URL parsing, node ID normalization, API URL building, HTTP, and generated API handling in `internal/figma`.
- Put document traversal and output shaping in `internal/extract`; no Cobra, env vars, stdout, or HTTP there.
- Keep dependency direction one-way: `cmd` may import `internal/*`; extract should not import CLI or command packages.
- Do not hand-edit `internal/figma/api/api.gen.go`; regenerate from `openapi/` instead.

## CLI Design Rules

- Prefer noun commands with flags over positional filter subcommands.
- Accept either a Figma file key or full Figma URL when possible.
- Normalize Figma node IDs at boundaries: user-facing `1-2`, API-facing `1:2`.
- Output structured JSON; include enough context for agents to identify nodes.
- Layer names may not be unique; return all matches with node IDs and extracted text.
- For named layer text extraction, preserve this shape:
  `figma texts --layer "Rectangle 3 Copy 15" <figma-file-url-or-id>`

## Testing and Verification

- Use TDD for behavior changes: add/adjust tests before implementation.
- Prefer package-level tests near changed logic.
- Run `go test ./...` before finalizing code changes.
- Run `go fmt ./...` for Go edits; use `make fmt` only when `goimports` is available.
- Mock HTTP with test servers or injected clients; do not require real Figma tokens in tests.

## Useful Commands

```bash
go test ./...
go fmt ./...
go build -o bin/figma ./cmd/figma
make test
```
