# Purpose

`internal/` contains the private capabilities behind the Figma CLI. It keeps composition, external adapters, pure Figma transforms, and output contracts separate so each boundary can be tested without a live service.

# Boundaries

- [CLI wiring](internal/cli/AGENTS.md) owns environment-backed dependency construction and error classification.
- [Figma transport](internal/figma/AGENTS.md) owns user input normalization, URLs, HTTP, and typed responses.
- [Document extraction](internal/extract/AGENTS.md) owns pure traversal and result shaping.
- [Asset workflows](internal/assets/AGENTS.md), [comments](internal/comments/AGENTS.md), and [history diff](internal/diff/AGENTS.md) own their capability decisions.
- [Output contracts](internal/output/AGENTS.md) own structured rendering and filesystem artifacts.
- Annotations and component parity helpers remain bounded cross-cutting packages owned by this guide unless their boundaries grow.

# Connections

- [Commands](cmd/AGENTS.md): the composition layer calls internal capabilities; internal packages never import commands.
- [Generated API](internal/figma/api/AGENTS.md): Figma adapters consume generated models and map them before pure extraction.
- [Output](internal/output/AGENTS.md): capabilities provide stable values and artifact paths to the shared renderer.

# Landmarks

- `internal/cli/runtime.go:LoadClient`: composition root for configured Figma transport.
- `internal/figma/input.go:ParseInput`: normalized Figma input boundary.
- `internal/extract/inspect.go:InspectTree`: pure document-to-output boundary.

# Boundary flows

- Information flow: `internal/cli/runtime.go:LoadClient` -> `internal/figma/client.go:NewClient` via `cmd/root.go:Execute`; value: `FIGMA_ACCESS_TOKEN`.
- Information flow: `internal/figma/document.go:FetchDocument` -> `internal/extract/inspect.go:InspectTree` via `cmd/root.go:Execute`; value: `document (any)`.

# Placement

Keep infrastructure at the edge and domain decisions in the owning capability. Add a package when a responsibility has a distinct input/output boundary and would otherwise create a dependency cycle or force unrelated callers to share policy.
