# Top actionable issues

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

Agents need prioritization, but deterministic metrics do not identify implementation actions.

## Proposed outcome

Split ranking from wording:

- deterministic engine ranks regions by explicit evidence (area, perceptual error, edge error);
- optional provider gives concise appearance context;
- command joins both without inventing calibrated confidence.

Suggested output:

```json
{
  "rankedIssues":[{
    "region":"r1",
    "rank":1,
    "rankEvidence":{"changedPixels":1600,"perceptualChangedRatio":0.9},
    "visualSummary":"Reference contains grey square; actual area is blank.",
    "advisory":true
  }]
}
```

## Pre-analysis

- Existing region metrics already provide deterministic ranking inputs in `internal/imagediff`.
- Optional descriptions live in `internal/imagecontext`; ranking must not move there.
- Joining deterministic rank and advisory wording belongs in pixel-perfect orchestration/output.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

Ranking is identical with and without provider access. Provider failure/disclaimer does not alter order or gates. Avoid `type: missing-element` unless visual provider clearly marks it advisory.

---

