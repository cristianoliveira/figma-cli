# specs (dev handoff)

## Problem

The Figma Inspect panel is the de facto handoff mechanism. But:

- **It's interactive** — you click one element at a time. Repetitive across 30+ components.
- **It's visual** — you read numbers off a panel and type them into code. Error-prone.
- **It's ephemeral** — specs live in Figma, not in the codebase. No version history, no PR review.
- **It's incomplete** — auto-layout, constraints, and spacing aren't surfaced in a single view.

The result: devs either copy specs manually (slow) or skip specs entirely and eyeball it
(wrong). There's no single command to say "give me everything Code needs to implement
this component."

## Success criteria

- One command extracts all implementation-relevant properties of a single node
  (or all nodes matching a name)
- Output includes: dimensions, auto-layout, fills, strokes, typography, component properties,
  constraints, spacing, corner radius, opacity
- Optional code-facing output formats: TypeScript props interface, Storybook args,
  CSS module skeleton
- Feeds directly into code generation or PR description templates

## API mapping

| Data | API endpoint |
|---|---|
| Node properties | `/files/{file_key}/nodes?ids=...` |
| Component properties (variants, boolean, text props) | Component data in node tree |
| Style references | Referenced style IDs resolved via `styles` map in response |
| Dev resources (annotations) | `/files/{file_key}/dev_resources` |

## CLI shape

```
figma specs <node-id-or-name> <file-url> [flags]

Flags:
  --name            Match by layer name (instead of node ID)
  --format          Output format: json, typescript, storybook, markdown (default: json)
  --recursive       Include children specs
  --page            Extract all components on a page
```

## Output examples

### JSON (default)
```json
{
  "name": "Button / Primary",
  "id": "104:15187",
  "type": "COMPONENT",
  "bounds": { "width": 200, "height": 48 },
  "layout": {
    "mode": "HORIZONTAL",
    "padding": { "top": 11, "bottom": 11, "left": 20, "right": 20 },
    "gap": 8,
    "primaryAxisAlignItems": "CENTER",
    "counterAxisAlignItems": "CENTER"
  },
  "fills": [{ "color": "#0667C8", "opacity": 1 }],
  "strokes": [],
  "cornerRadius": 12,
  "typography": { "fontFamily": "Inter", "fontSize": 16, "fontWeight": 600 },
  "componentProperties": {
    "variant": { "type": "VARIANT", "value": "Primary" },
    "disabled": { "type": "BOOLEAN", "value": false }
  }
}
```

### TypeScript
```typescript
interface ButtonProps {
  /** Variant of the button */
  variant: "Primary" | "Secondary" | "Tertiary";
  /** Whether the button is disabled */
  disabled: boolean;
  /** Button label */
  label: string;
}
```

### Markdown (for PR descriptions)
```markdown
## Button / Primary
- **Size:** 200×48px
- **Fill:** #0667C8
- **Corner:** 12px
- **Font:** Inter 16px / 600
- **Layout:** Horizontal, center-center, gap 8px, padding 20/11
```

## Edge cases

- Node is an INSTANCE (not COMPONENT) → resolve to main component, show both
- Auto-layout with wrapping → show `layoutWrap: "WRAP"` (maps to CSS flex-wrap)
- Text with mixed styles (rich text) → show range-based formatting
- Node has no fills (inherits from parent) → note "inherited" or resolve parent
- Component properties with no value set → show default from main component
