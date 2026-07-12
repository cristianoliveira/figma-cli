# Pixel-Perfect CLI + Figma Pain Points

## 1. Crop dimensions must match exactly
Figma frames are fixed-width (e.g. 780px). Implementation is flex-width.
Without viewport control on the capture tool, you can't resize to match.
The whole `--reference-crop`/`--actual-crop` comparison loop becomes unusable for flex-layout UIs.

Fix: `playwright-cli` needs `--viewport` or a `resize` command.
Or: `pixel-perfect` should accept proportional/percentage crops.

## 2. Finding crop coordinates is manual math
`figma export --scale 1` gives a full-frame PNG. To crop a sub-region like
"just the channel list" you do: (node_x - frame_x, node_y - frame_y, w, h).
Then add 1px for frame borders. Error-prone and slow.

Fix: `figma export --node-id <id>` that exports ONLY that node at its native bounds.
Or: `figma export --region x,y,w,h` in canvas coordinates.

## 3. media_describe is unreliable for pixel-level metrics
It said send button = 48px (actual: 40px), border = #E2E2E2 (actual: #DCE0E3).
Fine for layout discovery, harmful for pixel decisions.

Fix: never use media_describe for measurements. Only use scan/probe.

## 4. Probe/scan coordinate system is confusing with crops
When using `--reference-crop`, probe coordinates are in cropped space
(not original image). Easy to get wrong and waste iterations.

Fix: add a `--verbose` that prints the effective coordinate mapping.
Or: accept coordinates in original-image space with automatic translation.

## 5. The loop is too many tool calls per iteration
Export figma → capture browser → probe/scan → fix code → re-capture.
That's 4-5 tool calls per pixel-fix cycle. With rate limits and network latency,
this is ~30-60 seconds per iteration.

Fix: a combined command like:
`pixel-perfect compare --figma-url <url> --url <live-url> --selector ".class"`
that handles export, capture, cropping, diff, and probe in one shot.

## 6. Figma node scoping is hit-or-miss
`figma inspect` on component instances returns the instance wrapper,
not the resolved children. Need `--recursive` or separate calls for each child.
Made it hard to get full specs for nested components like the input bar icons.

Fix: `figma inspect --depth 2` or always resolve instance children.
