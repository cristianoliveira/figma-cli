# Relative bounds for scoped inspection

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

`inspect --recursive --id <scope>` returns absolute canvas coordinates. Consumers manually subtract scope origin.

## Proposed outcome

Add `relativeBounds` to descendants when inspection has an explicit scope:

```json
{
  "bounds": {"x":435.464,"y":623,"width":9.071,"height":16},
  "relativeBounds": {
    "x":299.464,"y":61,"width":9.071,"height":16,
    "relativeTo":"13576:15248"
  }
}
```

The scoped root should have relative `x=0,y=0`. Keep absolute `bounds` unchanged for compatibility.

## Pre-analysis

- `internal/extract/inspect.go`: output model and pure coordinate transform.
- `cmd/inspect.go`: pass selected scope identity/origin into extraction.
- Nearby tests: `internal/extract/inspect_test.go`, `cmd/inspect_test.go`.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

- Fractional coordinates are preserved without integer rounding.
- Nested descendants remain relative to requested scope, not immediate parent.
- Unscoped inspection omits `relativeBounds`.
- URL and explicit `--id` scopes behave identically.

## Implementation freedom

Implementer may calculate during traversal or in a post-transform. Prefer one source of truth and avoid mutating generated Figma models.

---

