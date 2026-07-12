---
name: pixel-perfect
description: >
  Compare equal-sized PNG screenshots with deterministic metrics, masks, overlays, regions, alignment hints, and CI gates using the pixel-perfect CLI.
  Use when user asks to compare screenshots, diagnose visual differences, validate pixel similarity, create a visual diff, or gate screenshot changes.
  Triggers: "compare these screenshots", "run pixel-perfect", "generate a diff mask", "measure visual similarity", "why do these PNGs differ".
  Works with local PNG files and the pixel-perfect CLI. Do NOT use for capturing browser screenshots, editing Figma files, or semantic CSS diagnosis without image inputs.
---

# pixel-perfect

## Objective

Measure and localize screenshot differences without silently resizing or aligning images. Preserve raw metrics as source of truth; treat classifications as diagnostic hints.

## Workflow

1. Establish baseline and review directional overlay/report:
   ```bash
   pixel-perfect reference.png implementation.png \
     --output diff-mask.png \
     --overlay diff-overlay.png \
     --report diff-report.html
   ```
2. Reject unequal dimensions unless you can make the logical regions equal with explicit crops. Prefer `--reference-metadata <figma-export.json>` when comparing a Figma export; otherwise use `--reference-crop` / `--actual-crop`.
3. Preserve three evidence layers:
   - raw: `changedPixels`, `changedRatio`, `rmse`, `rgbRmse`, `luminanceRmse`, `alphaRmse`, and `edgeRmse`
   - perceptual: `perceptualRmse`, `perceptualChangedPixels`, and `perceptualChangedRatio`
   - rendering: `antialiasedPixels`
4. Use `bounds` and `changedRows` to localize work, then rank regions. Read the same raw, perceptual, and antialias evidence per region before using classification hints.
5. Narrow the next comparison:
   ```bash
   pixel-perfect reference.png implementation.png \
     --reference-metadata reference.export.json \
     --actual-crop <x>,<y>,<width>,<height> \
     --region <x>,<y>,<width>,<height> \
     --threshold 8 \
     --perceptual-threshold 0.1 \
     --region-gap 8 \
     --min-region-pixels 12 \
     --output region-mask.png \
     --overlay region-overlay.png \
     --report region-report.html
   ```
6. Re-run after one bounded change; retain JSON and PNG artifacts.

## Diagnosis

| Evidence | Next action |
|---|---|
| High raw, low perceptual, high antialias share | Verify fonts, browser, device scale, and capture stability; avoid speculative CSS changes. |
| High `edgeRmse` | Inspect bounds, spacing, border, icon size, or displacement. |
| High perceptual error with low edge error | Inspect fill, text color, opacity, shadow, or gradient. |
| High `alphaRmse` | Inspect transparency, shadows, and effect padding. |
| `solid-fill` plus dominant color pair | Inspect fill/background color. |
| `mixed`, especially after `--region-gap` | Inspect pixels, Figma/DOM facts, and grouped subregions; do not force one diagnosis. |

`--suggest-offset <radius>` reports likely translation but never applies it. Red overlay means stronger/present in reference; green means stronger/present in implementation. Classifications are heuristic; raw evidence is authoritative. With crops, JSON `bounds` are cropped comparison coordinates; `inputBounds` maps regions back to original input screenshots. Use `--visual-context-prompt` only to focus advisory review text, not to create pass/fail evidence.

## Exclusions and CI

```bash
pixel-perfect reference.png implementation.png \
  --ignore-region <x>,<y>,<width>,<height> \
  --mask comparison-mask.png \
  --max-rmse 0.03 \
  --max-changed-ratio 0.02 \
  --max-perceptual-changed-ratio 0.01 \
  --output diff-mask.png
```

Repeat `--ignore-region` for known dynamic areas. In comparison masks, visible non-black pixels are included; black or transparent pixels are ignored.

## Guardrails

- Never resize inputs before comparison.
- Prefer CLI crops/metadata over external crop tools so `inputs.*.crop`, `inputBounds`, and reports preserve coordinate provenance.
- Never silently apply suggested translation.
- Keep viewport, device scale, browser, fonts, background, capture method, and shadow padding stable.
- Use regional metrics for component progress; whole-image RMSE can be dominated by unrelated background or effects.
- Do not claim exact CSS diagnosis from raster heuristics alone.

Full CLI reference: [`../../cmd/pixel-perfect/README.md`](../../cmd/pixel-perfect/README.md).
