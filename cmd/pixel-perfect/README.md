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
pixel-perfect reference.png implementation.png --threshold 8
```

The command prints JSON metrics to stdout and writes a transparent diff mask by default beside the actual image as `<actual>.diff.png`. Use `--output diff-mask.png` to choose a different mask path, and `--overlay diff-overlay.png` for a directional overlay.

Both inputs must be equal-sized PNGs. Unequal dimensions fail instead of producing invalid metrics.

## Pixel probe

```bash
pixel-perfect probe reference.png implementation.png --at 316,300
```

`probe` inspects one pixel in both equal-sized PNGs and prints RGBA, hex, and per-channel delta JSON. Use it when debugging exact colors, edge transitions, or when an agent would otherwise reach for ImageMagick pixel sampling. The point must be in bounds. It accepts the same input-preparation flags as diff: `--reference-crop`, `--actual-crop`, and `--reference-metadata`; coordinates are in comparison/cropped space. When crops are used, `inputPoint` maps the probed point back to original reference/actual coordinates.

For repeated boundary checks, scan one row or column into compact color runs:

```bash
pixel-perfect scan reference.png implementation.png --y 300  # horizontal row
pixel-perfect scan reference.png implementation.png --x 316  # vertical column
```

`scan` prints reference and actual runs with `start`, `end`, `length`, `rgba`, and `hex`, making edge transitions visible without N separate probe calls. It accepts `--reference-crop`, `--actual-crop`, and `--reference-metadata`; row/column indexes are in comparison/cropped space. When crops are used, `inputLine` maps the scanned row/column back to original reference/actual coordinates.

## Optional visual context

```bash
pixel-perfect reference.png implementation.png \
  --output diff-mask.png \
  --visual-context
```

`--visual-context` adds advisory appearance descriptions for deterministic changed regions. It does not alter metrics, classifications, or validation gates. OpenRouter remains default; select OpenAI with `--visual-context-provider openai`. Use `--visual-context-model` for per-call model override.

Configuration is based on Pi Spectacles at `~/.pi/agent/pi-spectacles.json`. OpenRouter uses its existing top-level settings and environment variables. OpenAI uses `OPENAI_API_KEY`, `OPENAI_VISION_MODEL`, and `OPENAI_BASE_URL`, or an optional nested configuration:

```json
{
  "openai": {
    "apiKey": "...",
    "model": "gpt-5.4-mini",
    "baseUrl": "https://api.openai.com/v1"
  }
}
```

The screenshots are sent only when `--visual-context` is present. If selected provider is not configured, comparison succeeds and `visualContext.disclaimer` explains why advisory context is unavailable.

## Focus and validation

```bash
pixel-perfect reference.png implementation.png \
  --region 522,282,200,48 \
  --threshold 8 \
  --output button-mask.png \
  --max-rmse 0.03 \
  --max-changed-ratio 0.02 \
  --max-perceptual-changed-ratio 0.01
```

- `--region x,y,width,height` compares one area while retaining absolute coordinates in JSON.
- `--threshold` ignores channel differences at or below the supplied value.
- `--max-rmse`, `--max-changed-ratio`, and `--max-perceptual-changed-ratio` make the process exit non-zero when a limit is exceeded.

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
  "antialiasedPixels": 1840,
  "evidence": {
    "rawOnlyPixels": 7119,
    "perceptualOnlyPixels": 0,
    "rawAndPerceptualPixels": 2310
  },
  "bounds": {"x": 12, "y": 80, "width": 520, "height": 310},
  "changedRows": [80, 81, 120, 121],
  "regions": [
    {
      "bounds": {"x": 522, "y": 282, "width": 200, "height": 48},
      "changedPixels": 9429,
      "changedRatio": 0.9822,
      "rmse": 0.143,
      "edgeRmse": 0.031,
      "perceptualRmse": 0.122,
      "perceptualChangedPixels": 7214,
      "perceptualChangedRatio": 0.7515,
      "antialiasedPixels": 318,
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
- `perceptualChangedPixels` and `perceptualChangedRatio`: pixels above the reported `perceptualThreshold`; configure it with `--perceptual-threshold` (default `0.1`; any finite non-negative HyAB distance is accepted). Raw changed-pixel evidence and `--threshold` semantics remain unchanged.
- `antialiasedPixels`: raw changed pixels that match neighborhood ramp evidence. This is report-only and never silently removes changes.
- `evidence`: exact overlap between raw and perceptual changed-pixel sets, exposing disagreement without forcing an interpretation.
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
