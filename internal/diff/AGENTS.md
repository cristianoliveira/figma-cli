# Agent Instructions for `internal/diff/`

## Purpose

Own pure design-diff use cases, beginning with text-change blame across version history.

## Rules

- Depend on narrow injected history interfaces, not Figma clients or generated API types.
- Keep version-search and diff algorithms deterministic and side-effect free.
- Put Figma-backed history adapters in `internal/figma`; compose them from `cmd`.

## Testing

- Use in-memory histories for introducing-version, bounded-history, and error-path tests.
- Run `go test ./internal/diff ./cmd` after changes.
