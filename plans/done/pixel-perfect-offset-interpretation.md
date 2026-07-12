# Suggested-offset evidence and interpretation

**Status:** Completed 12-07-2026. Output now includes baseline RMSE, improvement ratio, and deterministic `candidate-translation` / `inconclusive` interpretation with overlap guardrails and upstream negative-control fixtures.

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

`x/y/rmse` invites blind CSS shifts without showing whether offset materially helps.

## Proposed outcome

Extend deterministic evidence:

```json
{
  "suggestedOffset": {
    "x":-5,"y":-1,
    "rmse":0.15643,
    "baselineRmse":0.19192,
    "improvementRatio":0.1849,
    "interpretation":"candidate-translation"
  }
}
```

Interpretation must derive from documented deterministic thresholds. Use `inconclusive` when evidence is weak. Never say “apply this CSS transform.”

## Pre-analysis

- `internal/imagediff/offset.go` and tests.
- Keep interpretation pure and separately testable.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

Fixtures cover exact translation, partial translation, unrelated content change, flat/transparent image, and ties. Existing deterministic tie-breaking must remain stable.

## Implementation freedom

Calibrate thresholds against fixture corpus before choosing constants. Emit raw improvement even if no interpretation is assigned.

---

