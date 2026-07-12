# pixel-perfect: reference-metadata float→int + probe/scan papercuts

**Date:** 2026-07-12
**Context:** pixel-perfect-loop on Figma Header component (780×72, node `9078:238320`)

## 1. `--reference-metadata` FLOAT→INT BUG (REPRODUCED)

`figma export --metadata` emits `float`. `pixel-perfect --reference-metadata` expects `int`. Reproduced cleanly:

```bash
figma export --format png --metadata meta.json --id 9078:238320 <file-key>
# logicalCrop: { "x": 26.9871, "y": 16, "width": 780, "height": 72 }

pixel-perfect ref.png actual.png --reference-metadata meta.json
# error: json: cannot unmarshal number 26.9871 into Go struct field Bounds.logicalCrop.x of type int
```

The metadata types:
| Field | Type |
|-------|------|
| `logicalCrop.x` | float |
| `logicalCrop.y` | int |
| `logicalCrop.width` | int |
| `logicalCrop.height` | int |
| `exportPadding.left` | float |
| `exportPadding.right` | float |
| `exportPadding.top` | int |
| `exportPadding.bottom` | int |

Either `figma export` should round before serializing, or `pixel-perfect` should accept floats and round internally.

## 2. `logicalCrop` coordinates exceed image bounds

For a 780×72 node exported as 780×74:

```json
"logicalCrop": { "x": 26.9871, "y": 16, "width": 780, "height": 72 }
```

`x: 27` on a 780px-wide image doesn't fit. The `exportPadding` values (`left: 27`, `right: -27`) also cancel each other out. Is this coordinate relative to the Figma canvas rather than the exported image? The relationship between `nodeBounds`, `exportBounds`, `logicalCrop`, and `exportPadding` isn't documented anywhere.

**What would help:** A single `contentInset: {top, left, bottom, right}` in pixel units of the exported image. No math, no guessing.

## 3. `probe --at` vs `scan --y` flag inconsistency

Every subcommand invents its own coordinate syntax:

| Command | Coordinate flag |
|---------|----------------|
| `pixel-perfect` (compare) | `--region x,y,w,h` |
| `probe` | `--at x,y` (repeatable) |
| `scan` | `--x N` / `--y N` |

First attempts: `scan --row 50 → ❌`, `scan --at 50 → ❌`. Found `--y 50` on third try.

**Fix:** Unify to one pattern. `--at x,y` everywhere is clearest.

## 4. `scan --format json` produced empty output

```bash
pixel-perfect scan ref.png actual.png --y 50 --format json
# → JSONDecodeError: Expecting value (empty stdin)
```

`--format csv` worked fine. JSON should be the reliable machine format — if CSV works but JSON doesn't on the same command, something's broken.

## 5. playwright-cli: no viewport flag

The pixel-perfect-loop chain requires the implementation screenshot to match Figma dimensions. Without viewport control:

```
playwright-cli open --viewport-size 1400,800 → ❌ Unknown option
```

Had to capture at whatever the browser window was, then ImageMagick-crop to the app area. A `--viewport WxH` flag on `open` (or `resize-viewport WxH` command) would eliminate the crop dance.

## 6. ImageMagick `+repage` pitfall

Cropped images retain original canvas geometry:

```
$ magick identify cropped.png
cropped.png PNG 780x73 1918x992+879+96  ← canvas offset preserved
```

Subsequent `magick crop` operations fail silently producing 1×1 outputs:

```
$ magick cropped.png -crop 780x72+0+0 +repage out.png
magick: geometry does not contain image
$ magick identify out.png
out.png PNG 1x1  ← silently destroyed
```

Every crop in the chain needs `+repage` first. `pixel-perfect` could auto-repage input images or warn when canvas offsets are present.

## Related

- `../../plans/pains/figma-pixel-perfect-cli-feedback.md` — broader feedback on both CLIs (export padding, suggested-offset interpretation, visual-context)
- `../../plans/pains/pixel-perfect-probe-scan.md` — earlier probe/scan issues (crop flags now fixed, coordinate overhead, scan verbosity)
