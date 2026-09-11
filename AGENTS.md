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

# Dependency direction

These rules are enforced by `internal/architecture` fitness tests (run via `go test ./...`). They state direction, not a rigid layer taxonomy.

- **generated-api-boundary**: `internal/figma/api` (generated OpenAPI types) may be imported only under `internal/figma/**`.
- **domain-infra**: application/domain packages under `internal/**` must not import Cobra (`github.com/spf13/cobra`), HTTP (`net/http`), filesystem (`os`, `io/fs`), environment (`internal/env`), TOON (`github.com/toon-format/toon-go`), or generated Figma API (`internal/figma/api`). Legitimate edge owners: `internal/cli` (Cobra/env/HTTP/filesystem), `internal/figma` (HTTP transport + generated API), `internal/output` (TOON rendering only), `internal/artifact` (filesystem artifact persistence), `internal/assetsedge` (asset HTTP/filesystem adapters), `internal/env` (environment), `internal/components` (legacy codebase filesystem discovery).
- **internal-cmd**: no package under `internal/**` may import `cmd`; commands depend on internals, never the reverse.

Exemptions: `_test.go` files and generated `internal/figma/api/**` source. `path/filepath` is pure path manipulation (allowed); it is not filesystem access.

# Operational errors

`internal/operr` is the transport-neutral operational-error contract: a data-focused `Category` plus `ClassifiedError{Category, Message, Recovery, Err}`. Categories are `authentication`, `authorization`, `rate_limit`, `dependency_unavailable`, `invalid_input`, `artifact_access`, and `operational` (generic fallback).

- Adapters translate concrete failures at their boundary and retain the cause: `internal/figma` (status/network/decode), `internal/env` (missing token), `internal/artifact` (filesystem), `internal/assetsedge` (asset download status).
- `ClassifiedError` retains the underlying cause via `Unwrap`, so `errors.Is`/`errors.As` still reach the concrete type.
- Safe `Message`/`Recovery` never contain tokens, response bodies, headers, private paths, or raw OS text. Unknown failures fall back to the generic `operational` category with a safe message.
- The CLI renderer (`internal/cli/error_output.go`) consumes only `internal/operr`; it does not import Figma, HTTP, environment, or filesystem error types.

# Placement

Add a responsibility to the module that owns its boundary and stable output. Create a new top-level module only when the capability has distinct cohesion, collaborators, and dependency direction; do not split a small adapter or model folder merely because it has a name.
