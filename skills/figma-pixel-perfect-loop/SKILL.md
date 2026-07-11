---
name: figma-pixel-perfect-loop
description: >
  Iteratively match a local UI to a Figma frame using measured visual diffs.
  Use when user says "make this match Figma", "pixel perfect this page", "compare our UI to Figma", "improve visual similarity", or "keep refining the Figma implementation".
  Works with a Figma URL, local implementation, pixel-perfect CLI, figma CLI, and Playwright; ImageMagick is only an optional cross-check.
  Do NOT use for generic frontend implementation without a Figma reference, Figma editing, or a one-time visual review.
---

# Figma Pixel-Perfect Loop

## Objective

Turn Figma from inspiration into measurable source of truth. Improve one bounded region at a time; retain evidence for every iteration.

## Learned Constraints

- Measure before implementing. Recursive Figma inspect/CSS and native frame export are the source of truth; visual descriptions only orient the first inspection.
- Do not start with whole-page RMSE. It is a baseline, not a verdict: layout displacement can worsen it while a component gets closer.
- Preserve causality: one bounded region, one hypothesis, one CSS/asset change, one regional diff.
- Treat Figma-provided geometry, typography, colours, radius, shadows, and SVGs as data to reuse—not details to approximate.
- Calibrate the shared coordinate system before content: frame size, page offsets, sidebar, grids, and cards.
- Extract CSS variables only for Figma-proven shared tokens—colours, font family, radii, shadows, and repeated spacing. Keep region-specific geometry explicit so a local correction cannot silently move unrelated UI.

## Workflow

1. **Establish the baseline**
   - Confirm `figma me` works.
   - Export selected frame at native size:
     ```bash
     figma export --format png --output output/visual-diff/figma-reference.png --id <frame-id> <file-key>
     ```
   - Run a whole-frame baseline. `pixel-perfect` rejects unequal dimensions before producing misleading metrics:
     ```bash
     pixel-perfect \
       output/visual-diff/figma-reference.png \
       output/visual-diff/implementation.png \
       --output output/visual-diff/baseline-mask.png
     ```
   - Use returned disconnected `regions` to locate high-impact mismatch clusters. The whole-frame RMSE remains orientation, not verdict.
   - Capture specifications, never guess when Figma can answer:
     ```bash
     figma css --recursive --id <frame-id> <file-key> > .tmp/figma/frame.css
     figma inspect --recursive --id <frame-id> <file-key> > .tmp/figma/frame.json
     ```

2. **Match the page coordinate system first**
   - Read frame and child bounds from `frame.json`.
   - Implement viewport size, canvas, sidebar/content offsets, section positions, grid widths, card dimensions, and gaps before inner content.
   - Screenshot the implementation at the exported frame size with Playwright. Verify DOM bounding boxes against Figma bounds.

3. **Compare one region, not the whole page**
   - Whole-frame RMSE is only a baseline; it over-penalizes moved content.
   - Compare the same Figma and implementation region using a direct child frame's bounds relative to the selected parent:
     ```bash
     pixel-perfect figma-reference.png implementation.png \
       --region <x>,<y>,<width>,<height> \
       --threshold 8 \
       --output output/visual-diff/region-mask.png
     ```
   - Region masks and overlays are crop-sized; `comparedRegion`, mismatch `bounds`, and disconnected region bounds remain absolute full-image coordinates.
   - Add `--overlay <path>` when direction matters: red means stronger/present in Figma, green means stronger/present in implementation.
   - Use `--region-gap` to group nearby glyph clusters, `--min-region-pixels` to omit tiny clusters, and repeat `--ignore-region` for known dynamic or irrelevant areas.
   - Use `--suggest-offset <radius>` to report likely translation. Never apply it silently; original metrics remain authoritative.
   - Read global and per-region metrics diagnostically: `edgeRmse` emphasizes geometry, `luminanceRmse` brightness, `rgbRmse` color, and `alphaRmse` transparency/effects.
   - Use region hints carefully: `solid-fill` plus `dominantColorPairs` points to fill mismatch; `geometry` points to bounds/spacing; `sparse-raster` is commonly text or antialiasing; `mixed` needs visual inspection.
   - Use `--mask <png>` for irregular comparison areas. Visible non-black pixels are included; black or transparent pixels are ignored.
   - Keep screenshots, masks, overlays, and JSON metrics under `output/visual-diff/`.

4. **Refine outside-in, one region per loop**
   1. canvas and sidebar
   2. major sections and card grid
   3. individual cards
   4. typography, controls, icons, and shadows
   - For the selected region, extract its Figma bounds and styles. Translate exact fills, border radius, shadows, typography, padding, and child positions into CSS.
   - Export Figma SVG children for non-trivial icons instead of redrawing approximations.
   - Re-screenshot, re-crop, and compare. Record the new RMSE.
   - For text, also compare DOM line boxes: width, height, top offset, font family, size, line-height, weight, and link baseline. Fix text geometry before using a raster score to tune glyph rendering.

5. **Use metrics correctly**
   - A lower crop RMSE and changed ratio are evidence of improvement.
   - Use `--max-rmse` and `--max-changed-ratio` when the loop needs deterministic pass/fail validation.
   - An increased score is feedback, not failure: inspect the directional overlay, regional classification and dominant color pairs, coordinate alignment, and crop boundaries; then correct the largest discrepancy.
   - Do not claim pixel-perfect based only on DOM content or a whole-page metric.

## Stable Playwright Capture

Before every screenshot:

1. Set viewport to Figma export dimensions and `deviceScaleFactor: 1`.
2. Wait for `document.fonts.ready`; verify expected font families loaded rather than accepting fallback fonts.
3. Disable CSS animations, transitions, carets, and blinking cursors.
4. Use the same page background and transparency treatment as Figma export.
5. Capture a page region with explicit effect padding when shadows extend beyond element bounds; locator screenshots commonly clip them.
6. Keep browser engine, OS, font files, zoom, and screenshot method fixed across iterations.
7. Record viewport, scale factor, browser version, and capture bounds with artifacts.

Example readiness step:

```js
await page.setViewportSize({ width: frameWidth, height: frameHeight });
await page.evaluate(async () => { await document.fonts.ready; });
await page.addStyleTag({ content: `
  *, *::before, *::after {
    animation: none !important;
    transition: none !important;
    caret-color: transparent !important;
  }
` });
```

## Guardrails

- Use the Figma frame's native dimensions for both images. Do not resize either image before comparison: a rescaled screenshot changes antialiasing and invalidates RMSE.
- Capture the implementation in a wrapper whose bounds include the same effect/shadow extent as the Figma export; locator screenshots commonly clip shadows.
- Calculate every crop from Figma bounds; never eyeball crop coordinates.
- Change the smallest region that can improve the result. Do not mix unrelated refactors into a visual iteration.
- Before changing a shared CSS variable, identify all its consumers and re-diff each affected region. A token change is a multi-region iteration, not a local correction.
- Preserve responsive behavior, but calibrate the reference desktop viewport first.
- Verify meaningful DOM facts after each edit: page title, target card existence, and key element bounds.
- Write a short report with commands, metrics, artifacts, what improved, and remaining mismatch.

## Completion Checklist

- Reference and implementation screenshots have identical dimensions without image resampling, including equivalent shadow/effect padding.
- Every important region has a recorded `--region` comparison and mask PNG.
- Capture metadata fixes viewport, device scale, browser, font readiness, background, and effect padding.
- Each text region records Figma and DOM bounds; multiline copy, links, and control labels have matching line-box height and baseline before final raster comparison.
- CSS values trace back to Figma inspect/CSS output or exported assets.
- Region metrics are recorded and improving or explicitly explained.
- Final response states evidence and remaining differences; never merely says “matches.”
