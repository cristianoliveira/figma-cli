# Agent Instructions

## Start Here

- Read the nearest nested `AGENTS.md` before editing. Guidance exists for `cmd/`, `docs/`, `internal/`, and the `assets`, `comments`, `diff`, `extract`, `figma`, and generated API packages.
- Treat source, command tests, and live Cobra help as command-contract truth when prose documentation differs.

## Project Shape

This is a Go Cobra CLI for querying Figma and producing agent-friendly JSON, assets, CSS, and design tokens.

- Main Figma CLI entry point: `cmd/figma/main.go`
- Standalone image comparison entry point: `cmd/pixel-perfect/main.go`
- Figma Cobra commands: `cmd/`
- Runtime/env wiring: `internal/cli/`, `internal/env/`
- Output envelopes and file/stdout handling: `internal/output/`
- Figma API boundary: `internal/figma/`
- Pure extraction/transforms: `internal/extract/`
- Capability workflows: `internal/assets/`, `internal/comments/`, `internal/components/`, `internal/diff/`
- Image analysis and comparison: `internal/imagecontext/`, `internal/imagediff/`, `internal/pixelperfectcmd/`, `internal/pixelperfectreport/`
- Generated API models: `internal/figma/api/`
- OpenAPI source/config: `openapi/`

## Feature Map

- Workspace/API discovery: `me`, `projects`, `files`, `meta`, `versions`, `comments`.
- Design exploration: `find`, `inspect`, `layout`, `components`, `texts`, `colors`.
- Generation and export: `assets`, `export`, `css`, `tokens`.
- Change analysis: `diff text` compares copy across file versions.
- Screenshot analysis: `pixel-perfect` compares images independently of Figma; keep its generic image logic out of Figma packages.

## Architecture Rules

- Keep commands thin: parse flags, resolve input, call Figma/extract packages, print JSON.
- Put Figma URL parsing, node ID normalization, API URL building, HTTP, and generated API handling in `internal/figma`.
- Put document traversal and output shaping in `internal/extract`; no Cobra, env vars, stdout, or HTTP there.
- Keep dependency direction one-way: `cmd` composes internal packages; `internal/*` never imports `cmd`; `extract` never imports CLI, Cobra, environment, or HTTP packages.
- Keep generated Figma API types at the Figma or owning capability boundary; map them before pure extraction logic.
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
- Run smoke tests with `-count=1`; they invoke `go build` in subprocesses, which Go test caching does not reliably invalidate when command sources change.
- Enter the reproducible toolchain with direnv (`.envrc` uses `flake.nix`); it provides the Go, `golangci-lint`, and `goimports` versions expected here.
- Run `goimports -w <changed-go-files>`, `golangci-lint run ./...`, and `go test ./...` before finalizing.
- Mock HTTP with test servers or injected clients; do not require real Figma tokens in tests.
