# Explicit image cropping in pixel-perfect

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

Effect padding makes equal logical elements produce unequal PNG dimensions.

## Proposed outcome

Support independent input crops:

```bash
pixel-perfect reference.png actual.png \
  --reference-crop 8,8,320,1675 \
  --actual-crop 0,0,320,1675 \
  --output mask.png
```

Output records original dimensions and applied crops. Comparison coordinates refer to cropped images; crop origins remain available to map results back.

## Pre-analysis

- New pure crop/decode helpers under `internal/imagediff` or a narrowly named image-input package.
- `internal/pixelperfectcmd/command.go`: parse flags and construct inputs.
- Do not hide cropping inside existing comparison without making coordinate mapping explicit.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

- Out-of-bounds, zero, and negative crops fail clearly.
- Cropped dimensions must match before comparison.
- Input files are never overwritten.
- Region, ignore-region, overlay, mask, and suggested offset use documented cropped coordinates.
- Existing no-crop JSON remains compatible.

## Guardrails

Do not silently resize or align inputs.

---

