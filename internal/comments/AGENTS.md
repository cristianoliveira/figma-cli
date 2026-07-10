# Agent Instructions for `internal/comments/`

## Purpose

Own Figma comment API mapping, retrieval, and document-node scoping.

## Rules

- Keep generated comment API types and their mapping in this package; pass `extract.CommentOutput` to pure extraction helpers.
- Build comment web URLs here, where file key, node ID, and comment ID meet.
- Scope comments from fetched document ancestry/descendency; never infer hierarchy from node ID syntax.
- Preserve comment order and retain thread replies when their root is in scope.

## Testing

- Mock Figma HTTP responses; cover anchored and unanchored comments plus scope edge cases.
- Run `go test ./internal/comments ./cmd` after changes.
