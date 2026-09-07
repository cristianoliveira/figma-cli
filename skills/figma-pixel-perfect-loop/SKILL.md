---
name: figma-pixel-perfect-loop
description: >
  Iteratively refine a local UI against a Figma frame using measured visual diffs.
  Use for "make this match Figma" or "pixel perfect this page" when code changes are expected.
  Requires a Figma reference and local implementation; not for Figma editing or one-time image comparison.
---

# Figma Pixel-Perfect Loop

## Objective

Recreate Figma web UI as real, accessible DOM and improve measured similarity one bounded region at a time.

## Workflow

1. **Establish source of truth**
   - Verify `figma me`.
   - Export selected frame at native size with metadata:
     ```bash
     figma export --format png --output output/visual-diff/reference.png \
       --metadata output/visual-diff/reference.export.json --id <frame-id> <file-key>
     ```
   - Capture Figma facts instead of guessing:
     ```bash
     figma inspect --recursive --id <frame-id> <file-key> \
       --annotations-output output/visual-diff/reference.annotations.json \
       > .tmp/figma/frame.json
     figma css --recursive --id <frame-id> <file-key> > .tmp/figma/frame.css
     ```
   - Always pass export metadata when comparing a PNG produced by `figma export`:
     ```bash
     pixel-perfect \
       output/visual-diff/reference.png \
       output/visual-diff/actual.png \
       --reference-metadata output/visual-diff/reference.export.json \
       --annotations output/visual-diff/reference.annotations.json \
       --output output/visual-diff/mask.png \
       --overlay output/visual-diff/overlay.png
     ```
     Metadata applies the exported node's logical crop and removes effect overflow from comparison coordinates.

## Isolate stable implementation capture

How to achieve this? Create a new temporary page in the App you are working on.
Make sure to import all styles and components from the page you want to compare, and then export the page as a PNG.
Render the component/section you want to compare in the same way as the reference page.

Tips:
- Match native frame dimensions with `deviceScaleFactor: 1`.
- Wait for `document.fonts.ready`; verify expected fonts loaded.
- Disable animations, transitions, carets, and dynamic content.
- Keep browser, OS, zoom, background, and capture method fixed.
- Include equivalent effect/shadow padding.

## Measure one region
- Run `pixel-perfect` baseline with `--annotations output/visual-diff/reference.annotations.json`, then narrow to one direct child region.
- Use annotations when raw mismatch coordinates do not reveal which design element owns region. Example: instead of only seeing mismatch at `{x: 24, y: 80, width: 240, height: 48}`, enriched region may identify Figma node `13576:15248`, label `Selected sidebar row`, with 92% region overlap. This tells agent where to inspect Figma tree and which implementation component to search for; it does not prove whether problem is padding, translation, color, typography, or shape.
- Treat annotation matches as navigation hints. Start with match having strongest region overlap, inspect its Figma node facts and corresponding DOM/component, then combine that context with offset, edge, color, and bounds evidence before changing code. Parent and child annotations may both match same region; prefer most specific useful node rather than assuming first match is cause.
- For component work—especially when shared Figma URL targets specific component or frame rather than whole page—render real production component in isolated page, route, story, or preview. Compare there first, then verify it once in full page for integration regressions.
- Prefer exporting exact target node by node-scoped URL/`--id`. Always use `--reference-metadata` with a `figma export` PNG; use explicit crop flags only when metadata is unavailable or verified invalid. Do not export parent frame and manually subtract canvas coordinates when target node can be exported directly.
- A reference/actual dimension mismatch is a signal to verify that `figma export --metadata` and `pixel-perfect --reference-metadata` were used before calculating crops manually.
- When metadata is unavailable, use `pixel-perfect --reference-crop` / `--actual-crop` for comparison crops. Do not use ImageMagick crop chains: `+repage`/virtual-canvas offsets can silently change later crop geometry. Native CLI crops decode raster pixels directly and preserve `inputs.*.crop` provenance. Annotation coordinate space must match prepared reference dimensions; regenerate annotations for selected scope instead of editing coordinates by hand.
- If a `pixel-perfect` skill is available, load it for comparison flags and metric diagnosis. Otherwise, inspect `pixel-perfect --help` and use deterministic metrics, masks, and overlays directly.
- For exact color or boundary questions, use pixel-perfect `probe`/`scan`, never media description or visual-context prose. For spacing, run back-to-back `scan --row <y>` and `scan --column <x>` through same suspect coordinate; compare run boundaries and lengths between reference and actual. Multimodal descriptions may orient review but are not pixel, color, or geometry measurement tools.
- Keep screenshot, mask, overlay, report, and JSON evidence under `output/visual-diff/`.

## Change one cause
- Form one hypothesis from Figma facts, DOM bounds, overlay, and regional metrics.
- Make one bounded CSS, layout, typography, or asset change.
- Export real Figma SVG/image assets rather than approximating artwork.
- Re-capture and compare same region.
- Revert or revise when evidence worsens.

## **Refine outside-in**
- canvas/navigation → sections/grid → cards → typography/controls/icons/shadows.
- Re-diff every consumer after changing shared token.
- Stop speculative CSS changes when remaining error is verified rasterization noise.

## Loop Contract

After baseline, repeat until acceptance gate passes or budget is exhausted:

1. Select highest-impact actionable region: geometry first, then color/effects, then raster details. Prefer isolated development page for component work.
2. Record current regional metrics and DOM/Figma bounds.
3. Form one evidence-backed hypothesis and make smallest related change.
4. Verify behavior and key DOM facts; capture under same conditions.
5. Compare same region and affected shared-token consumers.
6. Keep change only when target improves without meaningful regression; otherwise revert it. A meaningful regression fails an accepted metric gate or breaks behavior or accessibility.
7. Record result and choose next region.

Default budget is 10 iterations per region and 30 iterations total. User-provided budget or gates take precedence. Do not loop indefinitely.

### Accessibility Gate

Before retaining each visual change:

- Prefer native semantic HTML over ARIA and custom controls.
- Use real `button`, `a`, `input`, `label`, `nav`, heading, list, and table elements where appropriate; never fake controls with images, SVG, canvas, or generic `div` elements.
- Keep interactive elements keyboard reachable and operable with visible focus.
- Preserve accessible names, roles, states, form labels, instructions, and errors.
- Hide decorative images and SVGs from assistive technology; give meaningful images appropriate alternative text.
- Preserve meaningful heading and landmark structure.
- Do not communicate meaning by color alone; retain reduced-motion behavior and usable reflow at 200% zoom.
- Run automated project accessibility checks when available.
- For interactive UI, always perform a keyboard smoke test: verify logical Tab order, visible focus, Enter/Space activation, Escape behavior where relevant, and accessible names.

Any visual improvement that breaks semantics, keyboard behavior, focus, or accessible naming is a regression and must be reverted.

### Acceptance Gates

Stop successfully when either:

- configured raw/perceptual metric gates pass for important regions—all visible user-facing regions inside selected Figma scope; or
- Figma and DOM geometry, content, typography, colors, and effects match, while remaining difference is evidenced as capture/font/browser rasterization noise.

Never define success as zero changed pixels unless user explicitly requires exact raster equality in fixed environment.
