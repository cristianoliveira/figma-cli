# Agent Instructions

## Start Here

- Read this file before changing code.
- For documentation search, read `docs/AGENTS.md` and use `qmd` when available.
- Check nested guidance before editing: `cmd/AGENTS.md`, `internal/AGENTS.md`, and package-specific `AGENTS.md` files.

## Project Shape

This is a Go Cobra CLI for querying Figma and producing agent-friendly JSON, assets, CSS, and design tokens.

- Entry point: `cmd/figma/main.go`
- Cobra commands: `cmd/`
- Runtime/env wiring: `internal/cli/`, `internal/env/`
- Output envelopes and file/stdout handling: `internal/output/`
- Figma API boundary: `internal/figma/`
- Pure extraction/transforms: `internal/extract/`
- Generated API models: `internal/figma/api/`
- OpenAPI source/config: `openapi/`

## Feature Map

- Workspace/API discovery: `me`, `projects`, `files`, `fetch-meta`, `versions`, `comments`.
- Design exploration: `find`, `inspect`, `layout`, `components`, `texts`, `colors`.
- Generation and export: `assets`, `export`, `css`, `tokens`.
- Change analysis: `diff text` compares copy across file versions.
- Treat live Cobra help and command tests as command-contract truth when README examples differ.

## Architecture Rules

- Keep commands thin: parse flags, resolve input, call Figma/extract packages, print JSON.
- Put Figma URL parsing, node ID normalization, API URL building, HTTP, and generated API handling in `internal/figma`.
- Put document traversal and output shaping in `internal/extract`; no Cobra, env vars, stdout, or HTTP there.
- Keep dependency direction one-way: `cmd` may import `internal/*`; extract must not import CLI or command packages. Imports of generated Figma API types are allowed for typed endpoint responses.
- Do not hand-edit `internal/figma/api/api.gen.go`; regenerate from `openapi/` instead.

## CLI Design Rules

- Prefer noun commands with flags over positional filter subcommands.
- Accept either a Figma file key or full Figma URL when possible.
- For node-scoped commands, infer `node-id` from URL and keep optional `--id` as explicit override for bare-key exploration. Never silently discard extra node IDs; reject unsupported multi-node input.
- Normalize Figma node IDs at boundaries: user-facing `1-2`, API-facing `1:2`.
- Treat numeric Figma URL fragments as comment IDs. Exact comment lookup takes precedence over broader node filtering.
- Keep machine-readable output stable. Use structured JSON for query results and deterministic text/files for CSS, tokens, and downloaded exports; preserve global `--json` envelope behavior.
- Layer names may not be unique; return all matches with node IDs and extracted text.
- For named layer text extraction, preserve this shape:
  `figma texts --layer "Rectangle 3 Copy 15" <figma-file-url-or-id>`

## Testing and Verification

- Use TDD for behavior changes: add/adjust tests before implementation.
- Prefer package-level tests near changed logic.
- Run `go test ./...` before finalizing code changes.
- Run `goimports -w <changed-go-files>` when available; otherwise use `gofmt`. CI checks both formatting and imports.
- Mock HTTP with test servers or injected clients; do not require real Figma tokens in tests.

## Useful Commands

```bash
go test ./...
go vet ./...
golangci-lint run ./...
go build -o bin/figma ./cmd/figma
make test
```
