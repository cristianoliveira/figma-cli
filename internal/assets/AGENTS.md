# Agent Instructions for `internal/assets/`

## Purpose

Own asset discovery, export URLs, download manifests, file downloads, and default export paths.

## Rules

- Keep Figma document/API calls and asset filtering together here; commands only validate flags and render results.
- Preserve manifest request order and report every attempted asset, including failures.
- Keep filenames deterministic and collision-safe.
- Inject HTTP clients and URL fetchers in tests; never require a real Figma token.

## Testing

- Test successful downloads, per-asset failures, filtering, and filename collisions.
- Run `go test ./internal/assets ./cmd` after asset/export changes.
