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
     figma inspect --recursive --id <frame-id> <file-key> > .tmp/figma/frame.json
     figma css --recursive --id <frame-id> <file-key> > .tmp/figma/frame.css
     ```

2. **Create stable implementation capture**
   - Match native frame dimensions with `deviceScaleFactor: 1`.
   - Wait for `document.fonts.ready`; verify expected fonts loaded.
   - Disable animations, transitions, carets, and dynamic content.
   - Keep browser, OS, zoom, background, and capture method fixed.
   - Include equivalent effect/shadow padding.

3. **Calibrate coordinate system first**
   - Match viewport, canvas, navigation, major sections, grid, card bounds, and gaps.
   - Compare DOM bounding boxes with Figma `relativeBounds`.
   - Use `spacingFromPrevious` to verify auto-layout gaps.

4. **Measure one region**
   - Run `pixel-perfect` baseline, then narrow to one direct child region.
   - For component work—especially when shared Figma URL targets specific component or frame rather than whole page—render real production component in isolated page, route, story, or preview. Compare there first, then verify it once in full page for integration regressions.
   - Prefer exporting exact target node by node-scoped URL/`--id`; use `--reference-metadata` when effect padding still requires logical cropping. Do not export parent frame and manually subtract canvas coordinates when target node can be exported directly.
   - Use `pixel-perfect --reference-crop` / `--actual-crop` for all comparison crops. Do not use ImageMagick crop chains: `+repage`/virtual-canvas offsets can silently change later crop geometry. Native CLI crops decode raster pixels directly and preserve `inputs.*.crop` provenance.
   - If a `pixel-perfect` skill is available, load it for comparison flags and metric diagnosis. Otherwise, inspect `pixel-perfect --help` and use deterministic metrics, masks, and overlays directly.
   - For exact color or boundary questions, use pixel-perfect `probe`/`scan`; visual-context prose is not pixel, color, or geometry measurement evidence.
   - Keep screenshot, mask, overlay, report, and JSON evidence under `output/visual-diff/`.

5. **Change one cause**
   - Form one hypothesis from Figma facts, DOM bounds, overlay, and regional metrics.
   - Make one bounded CSS, layout, typography, or asset change.
   - Export real Figma SVG/image assets rather than approximating artwork.
   - Re-capture and compare same region.
   - Revert or revise when evidence worsens.

6. **Refine outside-in**
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

### Blocked and Regression Policy

- Missing Figma permission/data: report missing evidence; do not guess.
- Missing font or asset: obtain correct source or report blocker; do not tune around fallback.
- Unstable capture: stop edits and stabilize environment first.
- Shared change regresses another accepted region: revert or replace with local correction.
- Two consecutive non-improving iterations on same hypothesis: abandon it and inspect new evidence.
- Budget exhausted: stop with best verified state, metrics, blockers, and next hypothesis.

## Non-negotiable Guardrails

- Figma exports are evidence and asset sources, never component implementations.
- Never render reference screenshot, frame export, crop, base64 capture, canvas copy, or screenshot-wrapped SVG as UI.
- Build selectable, semantic, accessible, interactive DOM components.
- Images are allowed only for genuine artwork such as photos, illustrations, logos, and icons—never flattened panels, forms, text, controls, navigation, sections, or screens.
- Never resize comparison images or silently apply suggested alignment.
- Never eyeball crop coordinates or use ImageMagick as the primary crop pipeline. Use Figma metadata or explicit `pixel-perfect` crop flags; use ImageMagick only as an independent diagnostic cross-check.
- Whole-frame RMSE is baseline, not proof of regional progress.
- Raster classification and visual context are advisory. Verify pixel dimensions and exact colors with Figma/DOM facts and pixel-perfect probe/scan.
- Preserve responsive behavior after calibrating reference viewport.

## Routing Checks

Should trigger:
- “Make this local page match this Figma frame.”
- “Keep refining CSS until visual diff improves.”
- “Compare our implementation to Figma and fix largest mismatch.”
- “Pixel perfect this component from Figma.”
- “Continue visual convergence loop.”

Should not trigger:
- “Inspect this Figma and list its components.” → use `figma-cli`.
- “Compare these two PNG files.” → use `pixel-perfect`.
- “Create a React page without a Figma reference.”
- “Edit this Figma design.”

## Completion

- No reference pixels are rendered by implementation.
- Semantic HTML, keyboard operation, visible focus, and accessible names pass the accessibility gate.
- Reference and implementation dimensions and capture conditions match.
- Important regions have recorded metrics, masks, and overlays.
- Changed CSS values trace to Figma facts or exported assets.
- DOM content and key bounds are verified after changes.
- Report commands, before/after regional evidence, artifacts, and remaining differences; never merely claim “matches.”
