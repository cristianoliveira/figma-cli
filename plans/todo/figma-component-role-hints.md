# Component-role hints

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

Layer lists do not identify likely structural relevance.

## Proposed outcome

Add opt-in `--handoff`/role hints with evidence:

```json
{
  "roleHint":"container",
  "roleEvidence":["large area relative to scope","behind 12 visible descendants"]
}
```

Use domain-neutral roles such as container, text, image, vector/icon candidate, and decoration candidate. Avoid claiming “button” or “avatar” unless component metadata supplies that identity.

## Pre-analysis

Pure heuristics belong in a dedicated extraction/handoff package, not command closures. Existing Figma type and hierarchy should carry more weight than layer names.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

Fixture corpus includes backgrounds, decorative vectors, selected-row fills, icons, text, and misleading names. Hints must be omittable and must never change base inspection output.

## Guardrails

Hints are advisory. Do not emit numeric confidence unless calibrated against labeled data.

---

