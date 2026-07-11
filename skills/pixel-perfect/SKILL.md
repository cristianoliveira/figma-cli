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

1. Establish baseline and review directional overlay:
   ```bash
   pixel-perfect reference.png implementation.png \
     --output diff-mask.png \
     --overlay diff-overlay.png
   ```
2. Reject unequal dimensions. Same dimensions are necessary but do not prove shared coordinates.
3. Read global `rmse`, `changedRatio`, `edgeRmse`, `rgbRmse`, `luminanceRmse`, and `alphaRmse`.
4. Rank returned regions. Inspect each region's RMSE, changed ratio, dominant color pairs, and classification.
5. Narrow the next comparison:
   ```bash
   pixel-perfect reference.png implementation.png \
     --region <x>,<y>,<width>,<height> \
     --threshold 8 \
     --region-gap 8 \
     --min-region-pixels 12 \
     --output region-mask.png \
     --overlay region-overlay.png
   ```
6. Re-run after one bounded change; retain JSON and PNG artifacts.

## Diagnosis

- `solid-fill` plus dominant color pair: inspect fill/background color.
- `geometry` or high `edgeRmse`: inspect bounds, spacing, border, icon, or displacement.
- `sparse-raster`: likely text or antialiasing; verify font and DOM line boxes.
- High `alphaRmse`: inspect transparency, shadows, and effect padding.
- `--suggest-offset <radius>` reports likely translation but never applies it.
- Red overlay means stronger/present in reference; green means stronger/present in implementation.

## Exclusions and CI

```bash
pixel-perfect reference.png implementation.png \
  --ignore-region <x>,<y>,<width>,<height> \
  --mask comparison-mask.png \
  --max-rmse 0.03 \
  --max-changed-ratio 0.02 \
  --output diff-mask.png
```

Repeat `--ignore-region` for known dynamic areas. In comparison masks, visible non-black pixels are included; black or transparent pixels are ignored.

## Guardrails

- Never resize inputs before comparison.
- Never silently apply suggested translation.
- Keep viewport, device scale, browser, fonts, background, capture method, and shadow padding stable.
- Use regional metrics for component progress; whole-image RMSE can be dominated by unrelated background or effects.
- Do not claim exact CSS diagnosis from raster heuristics alone.

Full CLI reference: [`../../cmd/pixel-perfect/README.md`](../../cmd/pixel-perfect/README.md).
