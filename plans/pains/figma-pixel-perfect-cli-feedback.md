# Figma CLI + pixel-perfect CLI feedback

Target used during feedback loop:

```bash
https://www.figma.com/design/exampleFileKey123/Example-Design?node-id=13576-15248&m=dev
```

Local implementation:

```bash
http://localhost:5173/
```

Main problematic area:

- sidebar node: `13576:15248`
- shape: `Rectangle Copy 13`
- grabber icon: `grabber left`
- selected menu item: `Sidemenu/Menu item/selected`

## Summary

Both tools were useful, but the loop still had too much agent guessing.

Best workflow discovered:

1. Use Figma CLI to identify real node hierarchy and exact vector/background shapes.
2. Export SVG for complex vectors before approximating CSS.
3. Use pixel-perfect with `--visual-context` on one small region.
4. Only then edit CSS.

The main mistake in this session was guessing the grabber shape as circle/half-tab/custom wave before checking `Rectangle Copy 13` SVG path.

---

# Figma CLI feedback

## 1. `inspect` gives bounds, but not enough shape truth

Command:

```bash
figma find --id 13576:15248 --name "Rectangle Copy 13" exampleFileKey123
```

Output:

```json
{
  "id": "I13576:15248;0:86",
  "name": "Rectangle Copy 13",
  "type": "VECTOR"
}
```

Then:

```bash
figma inspect --id 'I13576:15248;0:86' exampleFileKey123
```

Useful output:

```json
{
  "bounds": {
    "x": 136,
    "y": 562,
    "width": 320,
    "height": 166
  },
  "effects": [
    {
      "type": "DROP_SHADOW",
      "color": "#17181A",
      "radius": 8,
      "offsetY": 2
    }
  ]
}
```

But the real implementation detail needed was the vector path. That only appeared after:

```bash
figma export --format svg --id 'I13576:15248;0:86' \
  --output rectangle-copy-13.svg exampleFileKey123
```

The SVG showed the truth:

```svg
<path d="M8 6H320C320 6 320 29.8986 320 51C320 60 328 63.935 328 75C328 86.065 320 88.9604 320 99C320 113.466 320 172 320 172H8L8 6Z" />
```

That path explained the sidebar grabber shape. Before seeing it, I guessed wrong: circle, half-tab, custom wave.

### Proposed improvement

Add vector path details to inspect output for VECTOR nodes:

```bash
figma inspect --id 'I13576:15248;0:86' exampleFileKey123
```

Expected addition:

```json
{
  "vector": {
    "viewBox": "0 0 336 182",
    "paths": [
      {
        "d": "M8 6H320C320 ...",
        "fill": "#FFFFFF",
        "fillRule": "evenodd",
        "clipRule": "evenodd"
      }
    ],
    "exportPadding": {
      "left": 8,
      "top": 6,
      "right": 8,
      "bottom": 10,
      "reason": "drop-shadow"
    }
  }
}
```

This would prevent agents from approximating complex vector shapes.

---

## 2. Absolute bounds made implementation harder

Figma output gave absolute canvas coordinates:

```json
{
  "name": "grabber left",
  "bounds": {
    "x": 435.4644546508789,
    "y": 623,
    "width": 9.071067810058594,
    "height": 16
  }
}
```

The selected sidebar node has:

```json
{
  "name": "Sidemenu",
  "bounds": {
    "x": 136,
    "y": 562,
    "width": 320,
    "height": 1675
  }
}
```

To implement CSS, I had to manually subtract:

```txt
grabber relative x = 435.464 - 136 = 299.464
grabber relative y = 623 - 562 = 61
```

### Proposed improvement

For scoped inspect/find commands, include relative bounds:

```json
{
  "name": "grabber left",
  "bounds": {
    "x": 435.464,
    "y": 623,
    "width": 9.071,
    "height": 16
  },
  "relativeBounds": {
    "x": 299.464,
    "y": 61,
    "width": 9.071,
    "height": 16,
    "relativeTo": "13576:15248"
  }
}
```

Developer verification command:

```bash
figma inspect --recursive --id 13576:15248 exampleFileKey123 \
  | jq '.results[] | select(.name=="grabber left") | .relativeBounds'
```

---

## 3. Export dimensions vs logical dimensions were confusing

Logical node:

```json
{
  "width": 320,
  "height": 166
}
```

But SVG export:

```svg
<svg width="336" height="182" viewBox="0 0 336 182">
```

The extra pixels are effect/shadow padding. I lost time comparing `320x166` DOM against a `336x182` export.

### Proposed improvement

Make export output optionally return metadata:

```bash
figma export --format png --json --id 'I13576:15248;0:86' exampleFileKey123
```

Suggested JSON:

```json
{
  "output": "rectangle-copy-13.png",
  "nodeBounds": { "width": 320, "height": 166 },
  "exportBounds": { "width": 336, "height": 182 },
  "padding": { "left": 8, "top": 6, "right": 8, "bottom": 10 },
  "paddingReason": ["DROP_SHADOW radius=8 offsetY=2"]
}
```

This makes visual diff setup deterministic.

---

## 4. Need component-role hints

In the sidebar, several nodes matter differently:

- `Rectangle Copy 13`: background/container shape
- `Rectangle 15`: team avatar
- `grabber left`: icon
- `Title Copy 20`: text
- `Rectangle Copy 3`: selected nav background

The CLI lists them, but does not say which ones define structure.

### Proposed improvement

Add a handoff grouping mode:

```bash
figma inspect --handoff --id 13576:15248 exampleFileKey123
```

Suggested output:

```json
{
  "layers": {
    "containers": [
      {
        "name": "Rectangle Copy 13",
        "id": "I13576:15248;0:86",
        "reason": "large vector behind text and menu items"
      },
      {
        "name": "Rectangle Copy 3",
        "id": "I13576:15248;0:89;0:69",
        "reason": "selected menu item background"
      }
    ],
    "icons": [
      {
        "name": "grabber left",
        "id": "I13576:15248;0:99"
      }
    ],
    "text": [
      {
        "name": "Team Name",
        "characters": "Team Name"
      }
    ]
  }
}
```

---

# Pixel-perfect CLI feedback

## 1. `--visual-context` should be first-class for agents

Without visual context, metrics looked like:

```json
{
  "changedRatio": 0.08667,
  "rmse": 0.19192,
  "regions": [
    {
      "bounds": { "x": 40, "y": 30, "width": 40, "height": 40 },
      "classification": "solid-fill"
    }
  ]
}
```

Useful, but not enough.

With:

```bash
pixel-perfect reference.png actual.png \
  --region 0,0,320,166 \
  --threshold 8 \
  --visual-context \
  --output mask.png \
  --overlay overlay.png
```

It said:

```json
{
  "visualContext": {
    "regions": [
      {
        "referenceAppearance": "A dark grey square icon is visible.",
        "actualAppearance": "No icon is visible; the area is empty and white.",
        "visualContext": "Above the 'Team Name' text heading."
      }
    ]
  }
}
```

That was immediately actionable.

### Proposed improvement

Add an agent-focused preset:

```bash
pixel-perfect reference.png actual.png --agent --region 0,0,320,166
```

Equivalent to:

```bash
--threshold 8
--visual-context
--overlay overlay.png
--region-gap 8
--min-region-pixels 12
--suggest-offset 5
```

And emits:

```json
{
  "topActionableIssues": [
    {
      "type": "missing-element",
      "confidence": 0.91,
      "summary": "Reference has grey square avatar; actual is blank.",
      "bounds": { "x": 40, "y": 30, "width": 40, "height": 40 }
    }
  ]
}
```

---

## 2. Metrics improved while visual design got worse

I accidentally made the grabber a floating circle. Some metrics got better, but visually it was wrong.

The CLI could warn when shape semantics diverge:

```json
{
  "warning": {
    "type": "shape-mismatch",
    "reference": "continuous protruding wave/tab",
    "actual": "separate floating circle",
    "message": "Pixel score improved, but visual structure changed."
  }
}
```

### Developer-verifiable test case

Create two fixtures:

- reference: sidebar edge with wave tab
- actual: sidebar edge with circle button

Run:

```bash
pixel-perfect wave-reference.png circle-actual.png \
  --region 280,40,64,80 \
  --visual-context
```

Expected: visual context should explicitly classify as shape mismatch, not just `geometry`.

---

## 3. Need side-by-side HTML report

During debugging, I had:

- reference crop
- actual crop
- mask
- overlay
- JSON metrics

But I had to open them manually.

### Proposed improvement

Add:

```bash
pixel-perfect reference.png actual.png \
  --region 280,44,56,64 \
  --visual-context \
  --report report.html
```

Report should show:

1. reference crop
2. actual crop
3. overlay
4. mask
5. metric summary
6. visual-context comments
7. suggested offset

This would make review much easier for developers and designers.

---

## 4. Region coordinate mismatch needs stronger help

Pixel-perfect correctly rejects different dimensions:

```txt
error: image dimensions differ: reference is 336x1691, actual is 320x1675
```

Good. But next step is unclear.

In this workflow, Figma export included shadow padding. I manually cropped:

```bash
magick sidebar-reference.png -crop 320x1675+8+8 +repage sidebar-reference-crop.png
```

### Proposed improvement

Add an option:

```bash
pixel-perfect reference.png actual.png \
  --reference-crop 8,8,320,1675
```

or:

```bash
pixel-perfect reference.png actual.png \
  --auto-crop-transparent-padding
```

Even better: accept Figma export metadata:

```bash
pixel-perfect reference.png actual.png \
  --figma-export-metadata rectangle-copy-13.export.json
```

---

## 5. `suggestedOffset` is useful but needs interpretation

Example:

```json
{
  "suggestedOffset": {
    "x": -5,
    "y": -1,
    "rmse": 0.15643
  }
}
```

Useful, but risky. Agents may blindly shift UI.

### Proposed improvement

Add explanation:

```json
{
  "suggestedOffset": {
    "x": -5,
    "y": -1,
    "rmse": 0.15643,
    "interpretation": "Most changed pixels look like a translation. Verify DOM bounds before applying.",
    "affectedRegions": ["text", "icon"]
  }
}
```

---

# Concrete reproduction script

Developers can replay a similar loop:

```bash
mkdir -p output/debug-sidebar

figma inspect --recursive --id 13576:15248 exampleFileKey123 \
  > output/debug-sidebar/sidebar.json

figma export --format png --id 13576:15248 \
  --output output/debug-sidebar/sidebar-reference.png \
  exampleFileKey123

figma export --format svg --id 'I13576:15248;0:86' \
  --output output/debug-sidebar/rectangle-copy-13.svg \
  exampleFileKey123

# Crop reference to logical sidebar bounds when export includes shadow padding.
magick output/debug-sidebar/sidebar-reference.png \
  -crop 320x1675+8+8 +repage \
  output/debug-sidebar/sidebar-reference-crop.png

# Capture implementation however your app does it.
playwright-cli open http://localhost:5173/
playwright-cli screenshot '.side-menu' \
  --filename output/debug-sidebar/sidebar-actual.png

pixel-perfect \
  output/debug-sidebar/sidebar-reference-crop.png \
  output/debug-sidebar/sidebar-actual.png \
  --region 0,0,320,180 \
  --threshold 8 \
  --suggest-offset 5 \
  --visual-context \
  --output output/debug-sidebar/mask.png \
  --overlay output/debug-sidebar/overlay.png \
  > output/debug-sidebar/metrics.json
```

Then inspect:

```bash
jq '.visualContext.regions' output/debug-sidebar/metrics.json
jq '.regions[0:5]' output/debug-sidebar/metrics.json
```

---

# Biggest product lesson

For design implementation, numeric diff is necessary but not sufficient.

The best agent workflow needs three aligned truths:

1. Figma structural truth
   - bounds
   - relative bounds
   - vector paths
   - effect/export padding

2. Pixel visual truth
   - RMSE
   - changed regions
   - overlays/masks
   - suggested offsets

3. Semantic visual truth
   - missing avatar
   - wrong shape: circle vs wave
   - selected row too narrow
   - font weight differs

Right now both CLIs provide parts of this. The main improvement is to connect them so the next agent does less guessing and more verifying.
