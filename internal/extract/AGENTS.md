# Agent Instructions for `internal/extract/`

## Purpose

Convert parsed Figma document trees into command output structs and JSON-ready values.

## Rules

- Keep this package free of CLI and infrastructure concerns.
- Inputs are usually generic `map[string]any` / `[]any` trees produced from generated Figma API models.
- Prefer small traversal helpers with early returns over deeply nested walkers.
- Preserve all matches when names are ambiguous; include node IDs where users need disambiguation.
- Keep output structs close to extractor behavior and use JSON tags intentionally.
- Reuse shared value helpers from `values.go` instead of duplicating map conversion logic.

## Testing

- Test happy and unhappy/empty paths.
- Use minimal inline Figma-like maps unless a fixture is clearly better.
- Tests should assert stable output shape, not incidental traversal implementation.
