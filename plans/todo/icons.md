# icons (batch asset export)

## Problem

Designers create icon sets in Figma. Developers then manually:

1. Find each icon layer (out of 50–200 icons on a page)
2. Select it, set export format to SVG, click Export
3. Download the file, rename it to match a naming convention
4. Copy it into the project's `assets/icons/` directory
5. Repeat for dark mode variants, different sizes

This takes 30–90 minutes per icon set update. It's the most tediously manual step
in any design-to-code pipeline. Nobody likes it. Everyone has a half-broken script.

## Success criteria

- A single command exports all layers matching a name pattern as SVG/PNG
- Output files follow a naming convention: `icon-name.svg` (or configurable)
- Handles variants: same icon, different color/size → separate files
- Works with component instances (generic "Icon" component with different props)
- Zero manual renaming or organizing

## API mapping

| Step | API endpoint |
|---|---|
| Export a node as SVG/PNG | `/images/{file_key}?ids=...&format=svg&svg_include_id=false&svg_simplify_stroke=true` |
| Find all icon nodes | `/files/{file_key}/nodes?ids=...` (then walk children) |
| Export at specific scales | `/images/{file_key}?scale=2` |

## CLI shape

```
figma icons <file-url> [flags]

Flags:
  --name       Layer name pattern to match (default: matches all top-level nodes)
  --format     Export format: svg, png (default: svg)
  --output     Output directory (default: ./icons)
  --sizes      1x, 1.5x, 2x, 3x, 4x (for PNG only, default: 2x)
  --prefix     Filename prefix (default: "")
  --suffix     Filename suffix (default: "")
  --include-variants  Also export component variants within matches
```

## Output example

```
figma icons --name "icon/" --format svg --output ./src/icons
```

Produces:
```
src/icons/
  arrow-left.svg
  arrow-right.svg
  calendar.svg
  check.svg
  chevron-down.svg
  close.svg
  ...
```

## Naming rules

Figma layer name → filename:
```
"icon/arrow-left"           → arrow-left.svg
"icon / arrow-left"         → arrow-left.svg
"icon/24/arrow-left"        → arrow-left.svg      (size prefix stripped)
"icon/arrow-left / default" → arrow-left.svg       (variant stripped)
"icon/arrow-left / dark"    → arrow-left-dark.svg   (variant preserved)
```

## Edge cases

- Layer is a component instance, not a FRAME → extract the main component first
- Layer has no visible content (empty frame) → skip, warn
- Duplicate names after normalization → append counter (arrow-left.svg, arrow-left-2.svg)
- Mixed formats in same run → not supported; must pick one format per invocation
- SVG with embedded IDs (`<clipPath id="abc">`) → strip IDs to avoid collisions (`svg_include_id=false`)
- Very large icon sets (>200) → batch into groups of 10 API calls to avoid rate limiting
