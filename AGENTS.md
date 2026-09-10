# Purpose

Figma CLI turns Figma design data into agent-facing structured output, assets, CSS, and tokens.

# Architecture

[cmd](cmd/AGENTS.md) owns Cobra composition and process entrypoints. It wires capabilities from [internal](internal/AGENTS.md), but application policy belongs in the internal package that owns it. Figma transport and typed API boundaries stay in [internal/figma](internal/figma/AGENTS.md); generated models are isolated in [internal/figma/api](internal/figma/api/AGENTS.md). Pure document shaping belongs in [internal/extract](internal/extract/AGENTS.md). Output contracts are centralized in [internal/output](internal/output/AGENTS.md).

# Modules

- [Commands](cmd/AGENTS.md): Figma command composition.
- [Documentation](docs/AGENTS.md): user-facing command and workflow contracts.
- [Internal packages](internal/AGENTS.md): private runtime, API, extraction, and output capabilities.
- [Scripts](scripts/AGENTS.md): repository generation and smoke-support utilities.
- [Skills](skills/AGENTS.md): agent workflows for Figma exploration and implementation.

# Landmarks

- `cmd/figma/main.go:main`: Figma executable entrypoint; execution continues through `cmd/root.go:Execute`.
- `cmd/root.go:Execute`: composes command execution with structured errors and stable exit codes.
- `internal/figma/input.go:ParseInput`: accepts a file key or Figma URL at the API boundary.
- `internal/figma/document.go:FetchDocument`: retrieves and normalizes a Figma document tree.
- `internal/extract/inspect.go:InspectTree`: shapes a document tree into inspection output.
- `internal/output/printer.go:Printer.Structured`: emits the stable TOON/JSON result contract.

# Boundary flows

- Information flow: `internal/figma/client.go:Client.FetchJSON` -> `internal/figma/document.go:UnmarshalDocument` via `internal/figma/document.go:FetchDocument`; value: `api.GetFileResponse.Document`.
- Information flow: `internal/figma/document.go:FetchDocument` -> `internal/extract/inspect.go:InspectTree` via `cmd/root.go:Execute`; value: `document (any)`.

# Placement

Add a responsibility to the module that owns its boundary and stable output. Create a new top-level module only when the capability has distinct cohesion, collaborators, and dependency direction; do not split a small adapter or model folder merely because it has a name.
