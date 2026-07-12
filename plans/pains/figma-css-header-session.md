# Figma CLI: CSS output + inspect gaps from header session

**Date:** 2026-07-12
**Context:** pixel-perfect-loop on Figma Header component (780×72, node `9078:238320`)

## 1. `figma css` output is near-unusable for implementation

The Recursive CSS output for this 780×72 header produced 22 flat classes. Here's what they map to:

| CSS class | Figma layer | Real element |
|-----------|------------|--------------|
| `.header` | Header (INSTANCE) | — wrapper |
| `.header-2` | Header (FRAME) | top bar |
| `.button-icon-button` | Button / Icon Button | Search button |
| `.button-icon-button-2` | Button / Icon Button | Phone button |
| `.button-icon-button-3` | Button / Icon Button | Info button |
| `.icon`, `.icon-2`, `.icon-3` | Icon | duplicate icon classes |
| `.color`, `.color-2`, `.color-3` | Color (boolean op) | icon color fill |
| `.subtract`, `.oval-7`, `.oval-7-2`, `.rectangle`, `.oval` | — | **Info icon internals** |
| `.rectangle-9`, `.rectangle-9-2` | Rectangle 9 | **design artifact** |
| `.name` | Name (TEXT) | title placeholder |
| `.frame-1`, `.frame-520` | auto-frames | layout wrappers |
| `.navigation` | navigation | tab bar |
| `.conversation`, `.shared-drive` | TEXT | tab labels |

Three problems here:

### 1a. Auto-named duplicates with numeric suffixes

Three identical Icon Button instances → `.button-icon-button`, `.button-icon-button-2`, `.button-icon-button-3`. Identical CSS repeated three times. The agent has to mentally map `button-icon-button-2` → "that's the Phone button, and -3 is Info". No semantic mapping, just positional guessing.

### 1b. Component internals leaked as CSS classes

The Info icon is a single Figma boolean operation (Subtract). Its internals — two ovals, a rectangle, an ellipse — become five separate CSS classes (`.subtract`, `.oval-7`, `.oval-7-2`, `.rectangle`, `.oval`). These should be one SVG export, not five CSS declarations for individual vector pieces.

```css
.subtract   { background: #DADADA; }
.oval-7     { background: #33373A; }
.oval-7-2   { background: #33373A; }
.rectangle  { background: #33373A; border-radius: 1px; }
.oval       { background: #33373A; }
```

### 1c. Design artifacts treated as real elements

`Rectangle 9` appears twice — a 627×44 rectangle with `#EDEFF0` fill, once in the header bar and once in the tab bar. These are likely measurement overlays or component default backgrounds set to invisible. The CSS output gives them equal standing with real UI:

```css
.rectangle-9   { background: #EDEFF0; height: 44px; width: 627px; }
.rectangle-9-2 { background: #EDEFF0; height: 44px; width: 627px; }
```

In the pixel-perfect comparison, these rectangles showed up as `#EDEFF0` pixels in the reference where the implementation had `#FAFAFA` — creating phantom "mismatches" that aren't real design differences.

### What would help

**Option A — Filtering:** `figma css --skip-vectors --skip-hidden` to exclude boolean operation internals and invisible layers.

**Option B — Grouping:** `figma css --group-by-component` merges identical instances under one class and references it from the parent. Three Icon Buttons → one `.button-icon-button` class.

**Option C — Hierarchy output:** Nested CSS with parent selectors reflecting the actual Figma tree, not flat peer classes.

At minimum, excluding boolean operation children (`.subtract`, `.oval-7`) and invisible/hidden layers from CSS output would cut the noise by ~40% in this frame.

## 2. `characters: ""` for component-instance text

All three text nodes in the header had empty `characters`:

```json
{"type": "TEXT", "name": "Name",          "characters": ""}
{"type": "TEXT", "name": "Conversation",  "characters": ""}
{"type": "TEXT", "name": "Shared Drive",  "characters": ""}
```

These are component instances where text is bound via `componentProperties`. The actual text values are in the component property bindings, not in `characters`. The layer name ("Name") is misleading — it's the Figma layer name, not the rendered text.

To discover the real text content I had to:
1. See that `characters` is empty
2. Check `componentProperties` for text bindings
3. Guess from the layer name

**Fix:** When `characters` is empty and the node has `componentProperties` with text bindings, resolve and include the bound value. Even a `resolvedText: "Marketing Team"` field would make inspect self-contained.

## 3. CSS output is flat — no nesting, no hierarchy

```css
.header { ... }       /* wrapper */
.header-2 { ... }     /* inner frame */
.button-icon-button { ... }  /* child of header-2 */
```

The parent-child relationship isn't expressed. If `.header-2` has `padding: 8px`, and `.button-icon-button` is a child, that relationship only exists in the inspect JSON — the CSS file makes them peers. Implementing from CSS alone means guessing which class nests inside which.

**Fix:** `figma css --nested` emitting actual nested selectors:

```css
.header {}
.header .header-2 {}
.header .header-2 .button-icon-button {}
```

Or at minimum, a comment block showing the tree structure before the CSS rules.

## Summary

`figma inspect --recursive` was excellent — the right data, well-structured. `figma css --recursive` was the weak link. Three concrete improvements would make it implementation-ready:

1. **Exclude vector internals** — boolean operation children shouldn't become CSS classes
2. **Exclude invisible/hidden layers** — design artifacts pollute the output
3. **Resolve component-property text** — empty `characters` needs a fallback to the bound value

With these, the CSS output goes from "skimmable reference" to "copy-paste starting point."
