# Purpose

`internal/cli` owns shared runtime wiring for Cobra commands: configured client and printer construction, usage/error classification, executable discovery, and process exit-code mapping.

# Boundaries

This is composition glue, not a capability layer. It may read environment configuration and bind Cobra streams, but Figma workflows and result transformations belong in their owning packages.

# Connections

- [Environment](internal/AGENTS.md): supplies token configuration to the client composition path.
- [Figma boundary](internal/figma/AGENTS.md): receives the configured token and constructs the API client.
- [Output](internal/output/AGENTS.md): binds command output streams to the selected format.
- [Commands](cmd/AGENTS.md): callers use error and runtime helpers while retaining command ownership.

# Landmarks

- `internal/cli/runtime.go:LoadClient`: builds the configured Figma client.
- `internal/cli/runtime.go:NewPrinter`: binds TOON/JSON output to a Cobra command.
- `internal/cli/runtime.go:ExitCode`: maps errors to process contract codes.
- `internal/cli/error_output.go:RenderError`: renders structured process errors.

# Boundary flows

- Information flow: `internal/env/env.go:GetFigmaToken` -> `internal/figma/client.go:NewClient` via `internal/cli/runtime.go:LoadClient`; value: `FIGMA_ACCESS_TOKEN`.
- Information flow: `internal/cli/runtime.go:NewPrinter` -> `internal/output/printer.go:New` via `internal/cli/runtime.go:NewPrinter`; value: `output.Format`.

# Placement

Place shared process wiring here only when multiple commands need the same contract. Keep domain validation and network operations outside this package.
