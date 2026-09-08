# Purpose

`internal/components` discovers frontend component names and compares them with published Figma components, returning a stable parity result.

# Boundaries

The package owns filesystem discovery and name matching. Figma retrieval remains in [Figma transport](internal/figma/AGENTS.md), and command parsing remains in [commands](cmd/AGENTS.md).

# Connections

- [Figma transport](internal/figma/AGENTS.md): provides published component records.
- [Commands](cmd/AGENTS.md): supplies the codebase path and renders the comparison.

# Landmarks

- `internal/components/compare.go:DiscoverCodeComponents`: discovers immediate code components.
- `internal/components/compare.go:Compare`: returns matched, missing, and extra components.

# Placement

Keep component parity naming and filesystem policy here. Put document traversal in extraction and API calls in the Figma boundary.
