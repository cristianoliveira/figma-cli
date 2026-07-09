# Agent Instructions for `internal/figma/`

## Purpose

Own the Figma boundary: input parsing, node ID normalization, API URL construction, HTTP client behavior, typed generated responses, and document unmarshalling.

## Rules

- Accept user-friendly inputs at the boundary and normalize once.
- Keep URL builders and node ID helpers centralized; do not duplicate parsing in `cmd`.
- Inject configured HTTP clients for tests; do not add package-level mutable HTTP state.
- Wrap errors with operation context.
- Keep generated API type usage here when possible so extractors can stay simple.
- Do not hand-edit `api/api.gen.go`; update `openapi/` and generation scripts/config instead.

## Testing

- Use table-driven tests for URL parsing/building and node ID normalization.
- Use `httptest` or injected `http.Client` behavior for client tests.
- Do not require `FIGMA_ACCESS_TOKEN` or real network access.
