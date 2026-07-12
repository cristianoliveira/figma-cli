# Export metadata sidecar

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

Figma node logical bounds differ from SVG/PNG export bounds due to scale and effects.

## Proposed outcome

Add explicit metadata output rather than overloading normal file output:

```bash
figma export ... --metadata rectangle.export.json
```

Suggested versioned shape:

```json
{
  "version":1,
  "nodeId":"I13576:15248;0:86",
  "format":"svg",
  "scale":1,
  "nodeBounds":{"x":136,"y":562,"width":320,"height":166},
  "exportBounds":{"width":336,"height":182},
  "padding":{"left":8,"top":6,"right":8,"bottom":10},
  "paddingEvidence":["DROP_SHADOW radius=8 offsetY=2"],
  "output":"rectangle-copy-13.svg"
}
```

Separate measured facts from inferred attribution. If per-edge padding cannot be proved, emit total dimension delta and omit speculative edges/reasons.

## Pre-analysis

- `cmd/export.go`: flags and output contract.
- `internal/assets` / `internal/figma/export.go`: export workflow and node metadata retrieval.
- SVG dimensions/viewBox may require parsing exported bytes; PNG dimensions use standard library decoding.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

- PNG and SVG fixtures with/without effects.
- Scale factors are represented correctly.
- Metadata paths are deterministic and collision-safe.
- Export success with metadata-write failure returns an explicit partial-artifact error.

## Required investigation

Confirm what Figma export API returns versus what must be derived from exported files. Do not promise exact effect attribution before proving it with fixtures.

---

