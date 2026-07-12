# Reproducible comparison profiles

**Status:** Completed 12-07-2026.

## Problem

Repeated pixel-perfect invocations required verbose flags, but one universal `--agent` preset would encode arbitrary defaults that do not fit every screenshot.

## Outcome

Added versioned JSON profiles:

```bash
pixel-perfect reference.png actual.png --profile pixel-perfect.json
```

```json
{
  "version": 1,
  "suggestOffset": 8,
  "regionGap": 2,
  "minRegionPixels": 4
}
```

Precedence is:

```text
built-in defaults < profile values < explicit CLI flags
```

Resolved values and sources appear in output JSON. Unknown fields, unsupported versions, and invalid values fail explicitly. Profiles do not own input, crop, mask, report, overlay, or output paths.

## Verification

Package and built-CLI smoke tests cover profile loading, strict validation, and explicit flag precedence—including explicit values equal to built-in defaults.
