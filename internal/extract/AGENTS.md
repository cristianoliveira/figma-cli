# Purpose

`internal/extract` converts generic Figma document trees into stable, JSON-ready inspection, layout, text, component, CSS, token, asset, and diff values.

# Boundaries

This package is pure transformation logic. It may consume mapped document values and annotation types, but it does not own Cobra, environment, HTTP, stdout, or Figma URL decisions. Preserve document order and represent ambiguity rather than guessing.

# Connections

- [Figma boundary](internal/figma/AGENTS.md): maps generated API responses into document trees before extraction; extraction does not call the API.
- [Document model](internal/document/AGENTS.md): owns the stable tree shape and shared document-tree value coercion helpers (StringValue, NumberValue, OptionalNumber, NumberSlice, MapValue). Extract delegates to these via thin re-exports while the canonical home lives in the document package.
- [Internal cross-cutting types](internal/AGENTS.md): annotations and component helpers share bounded data contracts with extraction.
- [Diff capability](internal/diff/AGENTS.md): consumes extracted text values for history analysis.
- [Output](internal/output/AGENTS.md): commands pass extracted values to stable rendering; extraction does not render them.

# Landmarks

- `internal/extract/inspect.go:InspectTree`: bounded node inspection.
- `internal/extract/text.go:ExtractTextNodes`: ordered text extraction.
- `internal/extract/css.go:ExtractCSSRules`: CSS-oriented document transformation.
- `internal/extract/tokens.go:ExtractTokensFromDocument`: design-token transformation.
- `internal/extract/changes.go:DiffDocuments`: structural document comparison.

# Boundary flows

- Information flow: `internal/figma/document.go:UnmarshalDocument` -> `internal/extract/inspect.go:InspectTree` via `cmd/root.go:Execute`; value: `map[string]any`.
- Information flow: `internal/extract/text.go:ExtractTextNodes` -> `internal/diff/text_blame.go:FindTextChange` via `cmd/root.go:Execute`; value: `[]extract.TextNode`.

# Placement

Add a function here when it deterministically derives a user-facing value from a document tree. Put network retrieval, command validation, and artifact writing in their owning modules.
