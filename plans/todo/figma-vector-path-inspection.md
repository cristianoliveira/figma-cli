# Vector path truth

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

VECTOR inspection gives bounds/effects but not contour truth; agents approximate shapes incorrectly.

## Proposed outcome options

Prefer explicit opt-in because recursive path payloads may be large:

```bash
figma inspect --include-vector-paths --id <vector> <file>
```

or an export-oriented path:

```bash
figma export --format svg --json --id <vector> <file>
```

Return parsed SVG facts plus original asset path. Preserve original `d`; do not normalize away meaningful commands.

## Pre-analysis

- `cmd/inspect.go` or `cmd/export.go` orchestration.
- `internal/figma` for API access.
- A pure SVG metadata parser outside `extract` if paths only exist in exported SVG.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

- Multiple paths, fill/stroke, fill/clip rules, transforms, and viewBox.
- Instance IDs containing semicolons/colons.
- Payload limits for recursive inspection.
- Export/API errors remain distinguishable.

## Required investigation

Determine whether exact paths exist in Figma API document data. If export is required, make extra network request explicit in help/output. Do not imply `inspect` is purely local extraction when it triggers exports.

---

