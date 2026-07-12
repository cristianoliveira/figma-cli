# pixel-perfect color probe and scan

**Problem:** `pixel-perfect` reports where images differ, but cannot inspect absolute colors at arbitrary coordinates. Agents fall back to ImageMagick for pixel sampling and edge/color transition checks.

## Increment 1: point probe

Add a probe command:

```bash
pixel-perfect probe <reference.png> <actual.png> --at 316,300
```

Return JSON:

```json
{
  "point": {"x":316,"y":300},
  "reference": {"rgba":[255,255,255,255], "hex":"#FFFFFF"},
  "actual": {"rgba":[244,244,244,255], "hex":"#F4F4F4"},
  "delta": {"r":11,"g":11,"b":11,"a":0}
}
```

Requirements:
- PNG inputs only.
- Validate equal dimensions.
- Validate point is in bounds.
- Include RGBA, hex, and per-channel delta.
- Keep stdout JSON stable.

## Increment 2: line/profile scan

Add scan command or flags:

```bash
pixel-perfect scan <reference.png> <actual.png> --x 316
pixel-perfect scan <reference.png> <actual.png> --y 300
```

Return compact 1D color runs / transition points along the selected row or column. This answers “where does white become gray?” without external tools.

## Design note
Prefer subcommands over adding probe flags to the diff command: probing/scanning is image inspection, not diff mask generation.
