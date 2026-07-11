# pixel-perfect

Deterministic PNG comparison for visual-regression loops, CI, and screenshot diagnosis.

`pixel-perfect` measures raster differences and points to likely causes without silently resizing or aligning either image. It is generic and does not require Figma credentials.

## Install

```bash
# Run from this repository
nix run .#pixel-perfect -- --help

# Build
nix build .#pixel-perfect

# Or with Go
go install github.com/cristianoliveira/figma-cli/cmd/pixel-perfect@latest
```

## Quick start

```bash
pixel-perfect reference.png implementation.png \
  --output diff-mask.png \
  --overlay diff-overlay.png
```

Both inputs must be equal-sized PNGs. Unequal dimensions fail instead of producing invalid metrics.

## Focus and validation

```bash
pixel-perfect reference.png implementation.png \
  --region 522,282,200,48 \
  --threshold 8 \
  --output button-mask.png \
  --max-rmse 0.03 \
  --max-changed-ratio 0.02
```

- `--region x,y,width,height` compares one area while retaining absolute coordinates in JSON.
- `--threshold` ignores channel differences at or below the supplied value.
- `--max-rmse` and `--max-changed-ratio` make the process exit non-zero when a limit is exceeded.

## Diagnosis

```bash
pixel-perfect reference.png implementation.png \
  --output mask.png \
  --overlay overlay.png \
  --suggest-offset 5 \
  --region-gap 8 \
  --min-region-pixels 12
```

- `--overlay`: red is stronger/present in reference; green is stronger/present in implementation.
- `--suggest-offset`: reports best translation within radius but never applies it.
- `--region-gap`: groups nearby clusters, such as glyphs in one text block.
- `--min-region-pixels`: removes insignificant clusters from region reporting without changing global metrics.

## Ignore known differences

```bash
pixel-perfect reference.png implementation.png \
  --output mask.png \
  --ignore-region 0,0,100,40 \
  --ignore-region 500,200,80,80 \
  --mask comparison-mask.png
```

`--ignore-region` is repeatable. A comparison mask must match full reference dimensions: visible non-black pixels are included; black or transparent pixels are ignored.

## JSON result

```json
{
  "width": 575,
  "height": 477,
  "changedPixels": 9429,
  "comparedPixels": 274275,
  "changedRatio": 0.03437,
  "rmse": 0.093,
  "rgbRmse": 0.087,
  "luminanceRmse": 0.061,
  "alphaRmse": 0.012,
  "edgeRmse": 0.048,
  "perceptualRmse": 0.052,
  "perceptualChangedPixels": 2310,
  "perceptualChangedRatio": 0.0084,
  "perceptualThreshold": 0.1,
  "bounds": {"x": 12, "y": 80, "width": 520, "height": 310},
  "changedRows": [80, 81, 120, 121],
  "regions": [
    {
      "bounds": {"x": 522, "y": 282, "width": 200, "height": 48},
      "changedPixels": 9429,
      "changedRatio": 0.9822,
      "rmse": 0.143,
      "edgeRmse": 0.031,
      "dominantColorPairs": [
        {"reference": "#0667C8", "actual": "#1676D2", "pixels": 7214}
      ],
      "classification": "solid-fill"
    }
  ],
  "suggestedOffset": {"x": -2, "y": 1, "rmse": 0.041},
  "mask": "mask.png",
  "overlay": "overlay.png"
}
```

### Metrics

- `rmse`: normalized RGB or RGBA error, matching image transparency semantics.
- `rgbRmse`: color-channel difference.
- `luminanceRmse`: brightness difference.
- `alphaRmse`: transparency and effect difference.
- `edgeRmse`: local visible-luminance gradient difference; useful for geometry.
- `perceptualRmse`: OKLab HyAB color distance after alpha compositing; useful for human-visible color change.
- `perceptualChangedPixels` and `perceptualChangedRatio`: pixels above the reported `perceptualThreshold`; configure it with `--perceptual-threshold` (default `0.1`, range `0..1`). Raw changed-pixel evidence and `--threshold` semantics remain unchanged.
- `changedRatio`: thresholded changed pixels divided by compared pixels.
- `bounds`: smallest absolute rectangle containing all changed pixels.
- `changedRows`: sorted absolute row indexes containing changed pixels; useful as compact localization evidence.

### Region classifications

Classifications are deterministic hints; raw metrics remain authoritative.

- `solid-fill`: concentrated color mismatch with mostly aligned edges.
- `geometry`: edge difference dominates.
- `sparse-raster`: sparse change, often text or antialiasing.
- `mixed`: no dominant signal.

## Stable screenshots

For reproducible comparisons, keep these fixed:

- viewport and `deviceScaleFactor: 1`
- browser engine, OS, zoom, and screenshot method
- installed font files; wait for `document.fonts.ready`
- page background and transparency
- animation, transition, caret, and cursor state
- effect padding around shadows; element screenshots often clip them

Same dimensions are necessary but do not prove that screenshots share one coordinate system. Use DOM/design bounds and `--suggest-offset` to diagnose alignment.

## Limitations

`pixel-perfect` is a measurement and diagnostic primitive, not a CSS debugger. Raster results remain sensitive to fonts, browser/OS rendering, alpha, capture bounds, and antialiasing. Region classifications are heuristics and cannot reliably name the exact CSS property to change.
