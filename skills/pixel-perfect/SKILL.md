---
name: pixel-perfect
description: >
  Use for requests like "compare these screenshots" or "generate a visual diff".
---

# pixel-perfect

## Objective

Measure and localize screenshot differences without silently resizing or aligning images. Preserve raw metrics as source of truth; treat classifications as diagnostic hints.

## Workflow

1. Establish baseline metrics first. The CLI writes a default diff mask beside actual image as `<actual>.diff.png` and reports path in structured `mask` output (TOON by default, JSON with `--json`):
   ```bash
   pixel-perfect reference.png actual.png --threshold 8
   ```
   For named review artifacts, override the mask path and add overlay/report:
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
6. Re-run after one bounded change; retain structured evidence and PNG artifacts.
7. Probe exact colors instead of using external image tools:
   ```bash
   pixel-perfect probe reference.png implementation.png --at <x>,<y> --at <x2>,<y2>
   pixel-perfect probe reference.png implementation.png --from <x1>,<y1> --to <x2>,<y2> --step 4 --radius 1
   ```
   Use repeatable `--at` for known sparse points. Use inclusive `--from`/`--to` for diagonal or arbitrary straight boundaries; `--step` reduces samples and `--radius` catches thin/antialiased neighbors. Probe returns token-efficient CSV by default; use `--format json` when scripting needs structured `points[]`, RGBA, per-channel delta, and `inputPoint`. Output defaults to 25 points; when `truncated: true`, run emitted `hint` only if all points are needed, or set `--limit` explicitly. Reuse `--reference-crop`, `--actual-crop`, or `--reference-metadata` when diff used cropped inputs; probe coordinates are comparison/cropped coordinates.
8. Use scan for edge transitions instead of N probe calls:
   ```bash
   pixel-perfect scan reference.png implementation.png --row <row>
   pixel-perfect scan reference.png implementation.png --column <column>
   ```
   For spacing and bounds diagnosis, run a back-to-back cross through same comparison coordinate:
   ```bash
   # Horizontal line: widths, left/right edges, and horizontal gaps.
   pixel-perfect scan reference.png implementation.png --row <y>
   # Vertical line: heights, top/bottom edges, and vertical gaps.
   pixel-perfect scan reference.png implementation.png --column <x>
   ```
   Choose `<x>,<y>` so both lines cross suspected region while avoiding text or antialiased artwork when possible. Each command scans both reference and actual; a cross requires two commands, not four. Compare run start/end positions and lengths to distinguish shifted edges from wrong spacing. It returns compact CSV color runs by default; use `--format json` when scripting needs RGBA and `inputLine`. Output defaults to 25 runs per image; follow emitted `hint` only when truncation hides needed transitions. Reuse identical crop/metadata flags for both scans when original diff used cropped inputs; scan indexes are comparison/cropped coordinates.

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

## Guardrails

- Never resize inputs before comparison.
- Prefer CLI crops/metadata over external crop tools so `inputs.*.crop`, `inputBounds`, and reports preserve coordinate provenance.
- Never silently apply suggested translation.
- Keep viewport, device scale, browser, fonts, background, capture method, and shadow padding stable.
- Use regional metrics for component progress; whole-image RMSE can be dominated by unrelated background or effects.
- Do not claim exact CSS diagnosis from raster heuristics alone.
