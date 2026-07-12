# Self-contained HTML report

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

Reviewers manually open reference, actual, mask, overlay, and JSON.

## Proposed outcome

```bash
pixel-perfect reference.png actual.png \
  --output comparison.mask.png \
  --overlay comparison.overlay.png \
  --report comparison.html
```

Report contains:

1. run configuration and provenance;
2. reference and actual images/crops;
3. directional overlay and mask;
4. global metrics and gates;
5. ranked deterministic regions with crops;
6. suggested offset;
7. optional advisory visual context, visibly labeled advisory.

## Pre-analysis

- New `internal/pixelperfectreport` package using `html/template`.
- Pass completed result and artifact bytes into renderer.
- Keep HTML concerns out of `internal/imagediff`.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

- Golden or structural tests check escaping, required sections, and embedded images.
- Report works offline as one file.
- User-provided text/model output is HTML escaped.
- Missing optional overlay/context produces a useful report.
- Report cannot overwrite input or mask paths.

## Implementation freedom

Embedded data URLs maximize portability; linked files reduce report size. Start self-contained unless real fixture size proves unacceptable.

---

