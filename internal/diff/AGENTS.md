# Purpose

`internal/diff` contains pure design-history use cases, including locating the version that introduced a text change.

# Boundaries

The package depends on a `TextHistory` contract and extracted text values. It does not fetch Figma versions, parse Cobra flags, or render output.

# Connections

- [Extraction](internal/extract/AGENTS.md): supplies text nodes and text-diff values.
- [Figma transport](internal/figma/AGENTS.md): provides history adapters to callers.
- [Commands](cmd/AGENTS.md): maps command input to the use case and renders its result.

# Landmarks

- `internal/diff/text_blame.go:FindTextChange`: finds the earliest history version containing target text.

# Boundary flows

- Information flow: `internal/figma/blame.go:FetchAllVersions` -> `internal/diff/text_blame.go:FindTextChange` via `cmd/root.go:Execute`; value: `diff.TextHistory`.

# Placement

Keep history reasoning and comparison policy here. Put Figma version retrieval in the adapter and generic text traversal in extraction.
