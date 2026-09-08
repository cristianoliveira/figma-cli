# Purpose

`internal/comments` retrieves Figma comments, maps API replies into CLI values, and applies node-scoped comment workflows.

# Boundaries

It owns comment-specific API mapping and scope decisions. Generic document traversal belongs in [extraction](internal/extract/AGENTS.md); URL and HTTP details belong in [Figma transport](internal/figma/AGENTS.md).

# Connections

- [Figma transport](internal/figma/AGENTS.md): provides authenticated comment responses and file scope.
- [Extraction](internal/extract/AGENTS.md): provides comment output types and document-derived node relationships.
- [Commands](cmd/AGENTS.md): validates flags and renders comment results.

# Landmarks

- `internal/comments/comments.go:Fetch`: retrieves and maps comments for a file.
- `internal/comments/comments.go:Scope`: applies node, recursive, and ancestor scope.

# Boundary flows

- Information flow: `internal/figma/client.go:Client.FetchJSON` -> `internal/comments/comments.go:Fetch` via `cmd/root.go:Execute`; value: `api.GetCommentsResponse`.

# Placement

Put comment lifecycle and scope policy here. Keep generic node relationships in extraction and transport mechanics in the Figma boundary.
