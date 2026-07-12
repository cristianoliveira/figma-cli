# Vector path truth

**Status:** Completed 12-07-2026.

## Problem

Vector inspection exposed bounds and effects but not exact contours, encouraging agents to approximate icons and illustrations.

## Outcome

Added explicit API geometry opt-in:

```bash
figma inspect --include-vector-paths --id <vector-node> <file>
```

Output preserves:

- original fill and stroke path command strings;
- winding rules;
- optional fill override IDs and override table;
- relative transform;
- vector size.

Figma API supplies this through `geometry=paths`; SVG export and XML parsing are unnecessary.

## Guardrails

- Default inspect request and JSON remain unchanged.
- Geometry requests are explicit and depth-bounded.
- Non-recursive inspection requests depth 1.
- Recursive use requires explicit `--depth`.
- Path commands are preserved without parsing or normalization.
- Node IDs containing colons and semicolons remain URL-query encoded.

## Verification

URL tests cover geometry/depth query and composite IDs. Pure extraction tests cover multiple fill paths, stroke geometry, winding rule, override ID/table, transform, size, and exact path preservation. Command integration tests verify opt-in request and recursive depth guard.
