# Generic coordinate annotations

**Status:** Completed 12-07-2026 as a decoupled alternative to a Figma-specific handoff manifest.

## Problem

Pixel mismatch coordinates did not identify likely owning design elements. A shared Figma handoff manifest would couple generic pixel comparison to Figma concepts.

## Outcome

Added neutral, versioned coordinate annotations:

```bash
figma inspect --recursive --annotations-output frame.annotations.json <figma-url>
pixel-perfect reference.png actual.png --annotations frame.annotations.json
```

Figma maps selected-node descendants into screenshot-relative bounds. Pixel-perfect deterministically intersects mismatch regions and returns annotation IDs, labels, overlap ratios, and opaque metadata.

## Boundaries

- Pixel-perfect does not know about Figma.
- Annotations never change metrics, region detection, gates, or exit status.
- Schema dimensions must match prepared comparison image.
- Invalid versions, fields, IDs, dimensions, and bounds fail explicitly.
- Other producers such as browser or accessibility tooling may emit same schema.

## Verification

Pure schema/intersection tests, Figma extraction tests, mocked command integration tests, and built pixel-perfect smoke tests cover producer and consumer paths.
