# Shape mismatch warnings

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

A raster score may improve while topology becomes visually wrong, such as attached wave versus detached circle.

## Deterministic evidence to explore

- connected-component count;
- whether changed/reference foreground touches region boundary or known container edge;
- contour compactness and occupancy;
- edge distribution;
- reference/actual component bounds.

## Proposed outcome

Keep raw topology evidence deterministic. Optional visual provider may label it:

```json
{
  "visualContext": {
    "warnings":[{
      "type":"shape-mismatch",
      "summary":"Reference appears attached to edge; actual appears detached."
    }]
  }
}
```

Do not let advisory warning affect exit status.

## Pre-analysis

- Pure evidence under `internal/imagediff` only after a fixture demonstrates value.
- Prompt/output additions under `internal/imagecontext`.
- Provider parsing must reject unknown region IDs as today.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

Start with committed wave-versus-circle fixtures plus controls: translated wave, color-only change, antialias-only change. Require low false-positive rate before exposing a deterministic classification.

## Implementation freedom

The implementer may decide topology evidence is too brittle after fixture experiments. It is acceptable to ship only advisory descriptions rather than a misleading classifier.

---

