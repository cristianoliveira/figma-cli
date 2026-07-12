# Automatic transparent-padding crop

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Decision

Do not implement as default or silent behavior.

Transparent space and low-alpha shadows may be intentional. Prefer explicit crop or validated export metadata.

If experimentation is desired, expose analysis only:

```bash
--suggest-transparent-crop
```

Return candidate crop and evidence without applying it. Require explicit follow-up crop.

## Pre-analysis

- PNG alpha bounds can be measured in pure Go, but transparency may represent intentional spacing or effects.
- Existing equal-dimension rejection is a valuable guardrail and must remain default.
- Any candidate crop needs coordinate provenance and must not be applied implicitly.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

Fixtures include transparent intentional margins, shadows, isolated translucent pixels, and fully transparent images.
