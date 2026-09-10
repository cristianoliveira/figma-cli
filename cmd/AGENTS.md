# Purpose

`cmd/` owns Cobra command definitions and the Figma executable composition root. Commands translate flags and arguments into calls to private capability packages, then render results.

# Boundaries

Keep command handlers thin: validate command-owned input, resolve Figma scope, load dependencies, invoke an internal capability, and select output. Document traversal, API construction, and output shaping belong in internal packages.

# Connections

- [Internal runtime](internal/AGENTS.md): commands consume private capabilities and must not become their implementation layer.
- [CLI wiring](internal/cli/AGENTS.md): commands use client/printer construction and error classification.
- [Figma boundary](internal/figma/AGENTS.md): commands pass parsed file and node scope to transport helpers.
- [Extraction](internal/extract/AGENTS.md): document-oriented commands consume pure extraction results.
- [Output](internal/output/AGENTS.md): command results cross the output contract through the shared printer.
- [Assets](internal/assets/AGENTS.md): asset/export commands delegate download workflows.
- [Comments and history](internal/comments/AGENTS.md): comment commands delegate comment retrieval and scope handling.
- [Diff](internal/diff/AGENTS.md): history commands delegate text-blame use cases.

# Landmarks

- `cmd/root.go:Execute`: process-level execution, error rendering, and exit-code handoff.
- `cmd/figma/main.go:main`: Figma executable composition root.

# Boundary flows

- Information flow: `cmd/figma/main.go:main` -> `internal/cli/runtime.go:LoadClient` via `cmd/root.go:Execute`; value: `FIGMA_ACCESS_TOKEN`.
- Information flow: `internal/figma/document.go:FetchDocument` -> `internal/extract/inspect.go:InspectTree` via `cmd/root.go:Execute`; value: `document (any)`.

# Placement

Put a new user-facing workflow here only when it is command composition. Put reusable policy and transformations in the owning internal package; keep new executable wiring at the nearest `main` package.
