# layout (responsive specs)

## Problem

Figma auto-layout maps almost 1:1 to CSS Flexbox and Grid. But frontend developers don't
see the auto-layout properties — they see a visual preview and guess:

- "Does this stack vertically on mobile?" → find the mobile variant frame, check manually
- "What's the gap between these cards?" → hover each card, read the auto-layout panel, measure
- "Is this a wrap layout?" → unclear from visual alone, must open the panel
- "What breakpoint does this switch at?" → Figma doesn't natively show breakpoints;
  it shows variant frames for each size

The result: responsive behavior is re-discovered by every developer working on a screen,
or worse, hard-coded with wrong values that look "close enough."

## Success criteria

- Extract auto-layout properties from a frame and map them to CSS (Flexbox or Grid)
- For variant-based responsive designs, show the layout at each breakpoint
- Output is a human-readable summary + CSS-ready declarations
- Optionally detect layout inconsistencies across responsive variants
  (e.g., "this frame uses CENTER in desktop but START in tablet — intentional?")

## API mapping

| Data | API endpoint |
|---|---|
| Node tree with layout properties | `/files/{file_key}/nodes` |
| Auto-layout fields | `layoutMode`, `primaryAxisAlignItems`, `counterAxisAlignItems`, `itemSpacing`, `paddingLeft/Right/Top/Bottom`, `layoutWrap`, `counterAxisSpacing` |
| Responsive variants | Child nodes with names like "Desktop", "Tablet", "Mobile" |

## Figma → CSS mapping

| Figma property | CSS property |
|---|---|
| `layoutMode: HORIZONTAL` | `display: flex; flex-direction: row` |
| `layoutMode: VERTICAL` | `display: flex; flex-direction: column` |
| `primaryAxisAlignItems: MIN` | `justify-content: flex-start` |
| `primaryAxisAlignItems: CENTER` | `justify-content: center` |
| `primaryAxisAlignItems: MAX` | `justify-content: flex-end` |
| `primaryAxisAlignItems: SPACE_BETWEEN` | `justify-content: space-between` |
| `counterAxisAlignItems: MIN` | `align-items: flex-start` |
| `counterAxisAlignItems: CENTER` | `align-items: center` |
| `counterAxisAlignItems: MAX` | `align-items: flex-end` |
| `itemSpacing` | `gap` |
| `paddingLeft/Right/Top/Bottom` | `padding` |
| `layoutWrap: WRAP` | `flex-wrap: wrap` |
| `counterAxisSpacing` | `row-gap` (when wrapping) |
| `layoutSizingHorizontal: FILL` | `width: 100%` |
| `layoutSizingHorizontal: HUG` | `width: fit-content` |
| `layoutSizingHorizontal: FIXED` | `width: <value>px` |

## CLI shape

```
figma layout <node-id-or-name> <file-url> [flags]

Flags:
  --name           Match by layer name (instead of node ID)
  --format         Output format: summary, css, json (default: summary)
  --responsive     Extract layout across responsive variants
  --page           Extract all frames on a page
```

## Output example

```
$ figma layout --name "Card grid" --format css

/* Card grid (20089:685897) */
.card-grid {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  justify-content: flex-start;
  align-items: flex-start;
  gap: 16px;
  row-gap: 24px;
  padding: 20px;
}

.card-grid > * {
  width: 292px;   /* FIXED from layoutSizingHorizontal */
}
```

### Responsive variant output

```
$ figma layout --name "Card grid" --responsive

Desktop breakpoint (1440px):
  flex-direction: row, gap: 16px, 3 columns

Tablet breakpoint (768px):
  flex-direction: row, gap: 12px, 2 columns, flex-wrap: wrap

Mobile breakpoint (375px):
  flex-direction: column, gap: 8px, 1 column
```

## Edge cases

- Nested auto-layout → walk recursively, output nested CSS
- Mixed HUG and FILL children → correctly map `flex: 1` vs `flex: 0 0 auto`
- Absolute-positioned children → note that they're positioned absolutely (not in flex flow)
- Layout grids (CSS Grid-like) → Figma's layout grid system maps to CSS Grid
- Responsive frames named differently → heuristic: match frames whose names contain
  "Desktop", "Tablet", "Mobile", "Responsive", or similar patterns
- Children with `layoutPositioning: ABSOLUTE` → `position: absolute` with constraint-derived
  offsets
