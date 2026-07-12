# Pain points from full-page Figma→code session

**Date:** 2026-07-12
**Figma:** Drive--Cells- `grnVU2vAihHXwYgHryu2xE`
**Cards implemented:** 7 (Team Info, App Lock, SSO, SCIM, Sign, Cancel, Drive)

## Evaluation (2026-07-12)

This is a session retrospective, not an implementation backlog. Current triage:

- **Build:** scoped frame discovery and bounded pixel-perfect presentation.
- **Investigate first:** nested-instance PNG export.
- **Document instead of build:** perceptual metric calibration.
- **Already resolved:** export crop metadata, scoped `find`, computed inspect gaps, and explicit pixel-perfect cropping.
- **Move to owning tools/workflow:** browser defaults, Vite lifecycle, Playwright session cleanup, and accessibility-first implementation.

Live verification against the session file found:

```bash
# Scoped search succeeds; this returned matches without HTTP 400.
figma find --id 4:1082 --name Customise grnVU2vAihHXwYgHryu2xE

# Existing export metadata describes the 8px effect overflow exactly.
figma export --id 'I19247:678066' --format png \
  --output app-lock.png --metadata app-lock.json grnVU2vAihHXwYgHryu2xE

# Existing crop handoff removes the manual sips step.
pixel-perfect app-lock.png actual.png --reference-metadata app-lock.json
```

The verified App Lock metadata reports a 591×493 export, 575×477 logical crop, and 8px inset on each side.

## Actionable backlog

### P0 — investigate with evidence

1. **Nested-instance export:** compare the live Figma canvas, REST-exported PNG, and direct exports of the missing nested nodes. Determine whether the cause is Figma rendering, component permissions, node selection, or local export handling before designing a flag. The REST API renders PNGs server-side, so `--resolve-instances` or `--flatten` may not be implementable client-side.
2. **Scoped find regression:** the original failure no longer reproduces with `--id 4:1082`. Add a regression only if an exact scoped invocation still produces HTTP 400.

### P1 — product work

1. **Frame discovery:** design `figma frames [url] --id <page-or-section>`. It should return useful screen-level frames without dumping every nested layout frame. Define direct-versus-recursive semantics and test both before implementation.
2. **Bounded comparison presentation:** add an explicit summary mode without changing stable full JSON. It should retain dimensions, core ratios, mismatch bounds, top regions, and artifact paths while omitting unbounded row arrays.
3. **Perceptual guidance:** document raw versus perceptual evidence and provide example profiles/CI thresholds. Do not introduce a universal letter grade without validated calibration data.

### Workflow follow-ups

- Capture CSS user-agent resets and accessibility-first checks in the Figma-to-code workflow/skill.
- Report failed-session cleanup to `playwright-cli` ownership.
- Keep browser capture/process management outside `figma` and `pixel-perfect`; compose it in an orchestration script or skill.

## Figma CLI

### 1. No way to discover sibling frames
There's no `figma frames` or `figma pages` command. Finding the "next" frame required curling the REST API:
```bash
curl -s -H "X-Figma-Token: $TOKEN" "https://api.figma.com/v1/files/KEY/nodes?ids=4:1082&depth=4"
```
`inspect` on a CANVAS page returns the page itself, not its children. Had to hunt through section nodes to find sibling frames. A `figma siblings --id <node>` or `figma page-frames <page-id>` would cut this from minutes to seconds.

### 2. Export always adds ~16px padding — resolved
Every `figma export --format png --width 575` produces 591×493 for a 575×477 card. The extra pixels are shadow/effect overflows. `figma export --metadata` now emits `contentInset` and `logicalCrop`; `pixel-perfect --reference-metadata` consumes it, eliminating manual `sips` cropping.

### 3. Component instances don't render nested instances in PNG
Exporting `I19247:678066` (App Lock card) gave a card with empty button/dropdown regions. The nested `Button / Text Button` and `Dropdown / Dropdown` instances rendered as solid white. Pixel-perfect comparisons against these exports show 99% "changed" in those regions — false positives. Need a `--resolve-instances` or `--flatten` flag.

### 4. `find --name` fails on large files — scoped path resolved
The unscoped command can still request too much data, but `figma find --id <page-or-frame> --name "Customise"` scopes search to a node subtree. Live verification with `--id 4:1082` succeeded. Keep this as a regression report only if a scoped invocation still fails.

### 5. inspect doesn't show computed gaps — resolved
`figma inspect --recursive` now emits `spacingFromPrevious` for computed auto-layout sibling gaps. A separate `--gaps` mode is unnecessary unless a compact presentation is shown to improve the workflow materially.

## Pixel-perfect CLI

### 6. changedRows output is bloated
The full output includes 200+ line arrays of row numbers. For a 575×477 image that's the whole image. The `--summary` or first-N-regions output is what agents need. The raw row list is only useful for programmatic consumption.

### 7. No auto-crop to match dimensions — resolved explicitly
`pixel-perfect` accepts `--reference-crop`, `--actual-crop`, and `--reference-metadata`. Automatic largest-common-rectangle detection should not be added: it could silently compare unrelated coordinate spaces and conflicts with the tool's explicit, fail-fast behavior.

### 8. Perceptual threshold is a black box — documentation gap
`--perceptual-threshold 0.1` needs calibration examples and recommended starting profiles. A universal `--grade` would add subjective authority and hide tradeoffs between raw, edge, and perceptual evidence. Prefer explicit CI thresholds and documented profiles until grades can be validated against human judgments.

## CSS / DOM

### 9. Inline elements silently ignore vertical margins
`<label>` and `<span>` are `display: inline` by default. `margin-bottom: 8px` computed as 8px but visually applied as 0px. Spent 3 iterations adjusting numbers before realizing the computed style was lying. The fix was `display: block` — but the browser gave no warning. A linter or devtools hint would help.

### 10. Browser button defaults override custom styles
`.card-chevron` had `width:16px; height:16px` but rendered as 16×16 with `background: rgb(239,239,239)` and `border: 2px outset`. The CSS was correct but the browser user-agent stylesheet won. Needed explicit `background: none; border: none; padding: 0`. Same trap every time a `<div>` becomes a `<button>`.

### 11. Uniform flex gap can't match Figma's irregular spacing
Figma cards have gaps like 22, 8, 16, 7, 6, 33, 8 between consecutive elements. A CSS `gap: 16px` on a flex column applies uniformly. Switching to per-element `margin-top` values was the fix, but it meant 5+ CSS rules per card instead of one. An `inspect --gaps` output would make this translation mechanical instead of guesswork.

### 12. `<h2>` for a11y adds default margin
Changing card titles from `<span>` to `<h2>` for accessibility silently added browser default margins. Had to explicitly set `margin: 0` via the `.card-title` class. Easy to miss when measuring positions because the DOM doesn't show user-agent styles.

## Dev workflow

### 13. Vite dev server dies unpredictably
`npx vite --port 5174 &` in background survives ~5 minutes, then connection refused. Every 3rd playwright screenshot fails because the server died. Had to restart it ~10 times during the session. A `vite --watch` in a tmux pane or a proper process manager would help, but the quickest fix is a `figma compare` command that handles capture internally.

### 14. Playwright sessions accumulate from failed commands
Each failed `playwright-cli` command leaves the browser open with that session name. Had to close and reopen with new names (pi-v9, pi-v10, pi-v11...) because sessions couldn't be reused after errors. A `playwright-cli reopen` or auto-close-on-error would fix this.

### 15. Accessibility was retrofitted
Started with pixel-perfect, then user asked "is this a11y?" Had to rewrite: `<span>`→`<h2>`, `<div>`→`<input type="checkbox">`, `<div>`→`<select>`, `<div>`→`<button>`, add `aria-label`, `role`, skip-link. Each change shifted pixel positions and required re-measurement. A checklist upfront (headings, landmarks, form controls, focus, contrast) would make this a one-pass operation.

## What worked well

- `figma css --recursive` gave exact padding, gap, font values instantly
- `figma export --format svg --id` pulled real icons — no approximations
- `pixel-perfect` region analysis with dominant color pairs (`#D8D8D8 → #E8E8E8`) pinpointed exact color fixes
- playwright `eval` for DOM measurement gave ground truth positions
- The three-tool loop (inspect → measure DOM → pixel-perfect verify) converged reliably
