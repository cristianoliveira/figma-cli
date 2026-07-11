---
name: figma-pixel-perfect-loop
description: >
  Iteratively match a local UI to a Figma frame using measured visual diffs.
  Use when user says "make this match Figma", "pixel perfect this page", "compare our UI to Figma", "improve visual similarity", or "keep refining the Figma implementation".
  Works with a Figma URL, local implementation, figma CLI, Playwright, Nix, and ImageMagick.
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
   - Crop the same Figma and implementation region using a direct child frame's bounds relative to the selected parent. Compare with a temporary Nix shell:
     ```bash
     nix shell nixpkgs#imagemagick -c sh -c '
       magick figma-reference.png -crop <width>x<height>+<x>+<y> +repage reference.png
       magick implementation.png -crop <width>x<height>+<x>+<y> +repage implementation.png
       magick compare -metric RMSE reference.png implementation.png diff.png 2>rmse.txt; test $? -le 1
     '
     ```
   - Keep the reference, implementation crop, diff image, and metric under `output/visual-diff/`.

4. **Refine outside-in, one region per loop**
   1. canvas and sidebar
   2. major sections and card grid
   3. individual cards
   4. typography, controls, icons, and shadows
   - For the selected region, extract its Figma bounds and styles. Translate exact fills, border radius, shadows, typography, padding, and child positions into CSS.
   - Export Figma SVG children for non-trivial icons instead of redrawing approximations.
   - Re-screenshot, re-crop, and compare. Record the new RMSE.

5. **Use metrics correctly**
   - A lower crop RMSE is evidence of improvement.
   - An increased score is feedback, not failure: inspect `diff.png`, check coordinate alignment and crop boundaries, then correct the largest discrepancy.
   - Do not claim pixel-perfect based only on DOM content or a whole-page metric.

## Guardrails

- Use the Figma frame's native dimensions for both images.
- Calculate every crop from Figma bounds; never eyeball crop coordinates.
- Change the smallest region that can improve the result. Do not mix unrelated refactors into a visual iteration.
- Before changing a shared CSS variable, identify all its consumers and re-diff each affected region. A token change is a multi-region iteration, not a local correction.
- Preserve responsive behavior, but calibrate the reference desktop viewport first.
- Verify meaningful DOM facts after each edit: page title, target card existence, and key element bounds.
- Write a short report with commands, metrics, artifacts, what improved, and remaining mismatch.

## Completion Checklist

- Reference and implementation screenshots have identical dimensions.
- Every important region has a matching crop and `diff.png`.
- CSS values trace back to Figma inspect/CSS output or exported assets.
- Region metrics are recorded and improving or explicitly explained.
- Final response states evidence and remaining differences; never merely says “matches.”
