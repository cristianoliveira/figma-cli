# Spacing verification

## Problem

Frontend implementation often needs exact distance between sibling text layers and frames. Existing commands expose declared auto-layout rules and node bounds, but developer must combine `css --recursive` with several `inspect` calls and calculate:

```text
previous edge = previous position + previous size
gap = next position - previous edge
```

This is slow and fragile. Declared auto-layout `itemSpacing` is also not always same as measured visual distance when nodes use absolute positioning, negative spacing, wrapping, or non-layout containers.

## Solution

Extend `figma layout` with opt-in computed sibling spacing rather than introduce another overlapping tree command:

```bash
figma layout --measure-spacing <frame-url>
```

Default layout extraction remains unchanged and avoids bounds calculations. With measurement enabled, return both:

- declared parent gap from Figma auto-layout;
- measured gap from previous sibling using node bounds and parent axis.

Add focused pair comparison only if real usage shows tree output is insufficient:

```bash
figma spacing --from 4707:15506 --to 4707:15508 <file-url>
```

Do not overload repeatable `--id`; use explicit `--from` and `--to` roles.

## How

- Add `--measure-spacing`, wired to `ExtractLayout(..., LayoutOptions{MeasureSpacing: true})`.
- Preserve `absoluteBoundingBox` long enough during opt-in layout extraction to calculate sibling edges.
- For vertical parent layout:
  `next.y - (previous.y + previous.height)`.
- For horizontal parent layout:
  `next.x - (previous.x + previous.width)`.
- Add `gapFromPrevious` to each child after first; omit when bounds or axis are unavailable.
- Keep parent `gap` as declared `itemSpacing`, so output can expose mismatches instead of hiding them.
- Include child padding and layout properties already available in compact layout tree.
- Mark absolutely positioned children and avoid presenting their measured distance as auto-layout gap.
- Handle negative gaps, fractional values, hidden children, wrapping rows, and mixed absolute/flow children explicitly.
- Add deterministic fixtures for vertical, horizontal, absent bounds, negative spacing, and declared-versus-measured mismatch.
- Validate against WPB-27007 nodes `4707:15506` and `4707:15508`, expecting measured vertical gap `16`.

## Example

```json
{
  "id": "4707:15504",
  "layoutMode": "VERTICAL",
  "gap": 16,
  "padding": {"top": 40, "right": 40, "bottom": 40, "left": 40},
  "children": [
    {"id": "4707:15506", "name": "Sending and receivin", "type": "TEXT"},
    {"id": "4707:15508", "name": "Frame 435", "type": "FRAME", "gapFromPrevious": 16}
  ]
}
```

## Success criteria

- One `figma layout --measure-spacing <frame-url>` call reports sibling spacing in tree order.
- Default `figma layout` output and computational path remain lightweight.
- Developer can distinguish declared auto-layout gap from measured geometric gap.
- Missing/absolute/wrapped layout cases are explicit, never guessed.
- WPB-27007 spacing can be verified without manual arithmetic.
