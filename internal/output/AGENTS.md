# Purpose

`internal/output` owns stable result envelopes, TOON/JSON rendering, and text/file compatibility behavior. It is rendering-only: it owns no filesystem, directory creation, or file-permission policy.

# Boundaries

Callers provide domain values and select the output mode. This package must not know Figma command semantics, image-analysis policy, environment configuration, or durable storage. Persistence is owned by [artifact](internal/artifact/AGENTS.md); asset URL delivery by [assetsedge](internal/assetsedge/AGENTS.md).

# Connections

- [CLI wiring](internal/cli/AGENTS.md): selects the format and binds the printer to a command stream.
- [Commands](cmd/AGENTS.md): consume the stable renderer for user-facing results.
- [Internal capabilities](internal/AGENTS.md): provide query/detail values and artifact paths.
- [Artifact persistence](internal/artifact/AGENTS.md): owns durable byte writes to paths.

# Landmarks

- `internal/output/contracts.go:NewQuery`: creates a non-null collection envelope.
- `internal/output/contracts.go:NewLimitedQuery`: adds bounded-result metadata.
- `internal/output/printer.go:Printer.Structured`: emits TOON by default or compatibility JSON.
- `internal/output/printer.go:Printer.File`: renders a path/metadata file-result envelope (does not persist).

# Boundary flows

- Information flow: `internal/extract/inspect.go:InspectTree` -> `internal/output/printer.go:Printer.Structured` via `cmd/root.go:Execute`; value: `extract.InspectOutput`.

# Placement

Add a renderer or envelope here only when it is shared output policy. Keep domain-specific result construction in the producer package, and durable storage in `internal/artifact`.
