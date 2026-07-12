# pixel-perfect scan/probe pain points

**Date:** 2026-07-12  
**Context:** pixel-perfect-loop on Figma customization page card (575×477)

**Update:** `--reference-crop` / `--actual-crop` / `--reference-metadata` are now available on `probe` and `scan`. The pain below was from an outdated binary. Leaving the rest as a note on the workflow friction that existed.

## 1. `probe` and `scan` lacked crop flags (FIXED)

The main `pixel-perfect` command supported `--reference-crop` and `--actual-crop`, but an older version of `probe` and `scan` did not. This forced pre-cropping with ImageMagick.

**Impact:** When the reference has different dimensions than the implementation (e.g., reference 591×493 with export padding, implementation 575×477), you can't just `probe --at x,y` — the coordinate systems don't match. You're forced to pre-crop with ImageMagick or manually translate coordinates between images.

**What would help:** `probe --reference-crop x,y,w,h --actual-crop x,y,w,h` matching the compare command's interface. Or alternatively, a `pixel-perfect crop` subcommand that writes pre-cropped PNGs to disk for downstream use.

## 2. Coordinate system mental overhead

When comparing cropped regions, every `probe` coordinate is in the cropped image's space, not the original image. This means:
- Checking Figma absolute bounds requires manually subtracting crop offsets
- Switching between full-image and cropped regions requires recalculating all coordinates
- Error-prone when iterating between `compare` (which accepts crops) and `probe` (which doesn't)

**What would help:** `probe` printing both the cropped coordinate AND the original-image coordinate in the output. Or a `--coordinate-origin` flag that accepts a coordinate system name.

## 3. `scan` output verbosity

`scan` dumps every pixel in the row/column as separate color entries. For a 536px row, that's 536 lines. There's no compact/run-length output mode.

**What would help:** `scan --compact` that outputs run-length-encoded color runs (e.g., `[0-50]: #FFFFFF, [51-55]: #D8D8D8`).

## Workaround used

Resorted to ImageMagick `convert -crop` to pre-crop images so `probe` could work against aligned same-sized crops. Clunky but functional.
