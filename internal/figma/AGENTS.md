# Purpose

`internal/figma` is the Figma boundary: parse user inputs, normalize node IDs, build API URLs, perform HTTP requests, decode typed responses, and expose document/history helpers to capability packages.

# Boundaries

Normalize friendly URL and node-ID forms once at this boundary. Keep generated API types behind this package where possible; map document responses before passing them to pure extraction. Commands must not duplicate URL or scope rules.

# Connections

- [Generated API](internal/figma/api/AGENTS.md): supplies generated response and model types; never hand-edit generated output.
- [Extraction](internal/extract/AGENTS.md): consumes normalized generic document trees from this boundary.
- [History diff](internal/diff/AGENTS.md): receives a Figma-backed `TextHistory` implementation for blame queries.
- [CLI wiring](internal/cli/AGENTS.md): constructs configured clients but does not own Figma protocol policy.
- [Commands](cmd/AGENTS.md): callers provide file and node scope and consume boundary results.

# Landmarks

- `internal/figma/input.go:ParseInput`: parses file, node, and comment scope.
- `internal/figma/node_id.go:NormalizeNodeID`: normalizes user-facing node IDs.
- `internal/figma/client.go:Client.FetchJSON`: shared authenticated HTTP boundary.
- `internal/figma/document.go:FetchDocument`: fetches and maps a document tree.
- `internal/figma/document.go:FetchNodeDocuments`: fetches exact requested subtrees.

# Boundary flows

- Information flow: `internal/figma/client.go:Client.FetchJSON` -> `internal/figma/document.go:UnmarshalDocument` via `internal/figma/document.go:FetchDocument`; value: `api.GetFileResponse.Document`.
- Information flow: `internal/figma/document.go:FetchDocument` -> `internal/extract/inspect.go:InspectTree` via `cmd/root.go:Execute`; value: `document (any)`.

# Placement

Put transport, URL, authentication, and Figma-specific mapping here. Put generic traversal in extraction and command-specific flag policy in commands.
