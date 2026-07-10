---
name: figma-cli
description: "Use this to answer frontend-relevant questions about Figma designs."
---

# Figma CLI

## Purpose

The `figma` CLI turns Figma designs into frontend-ready output: CSS, design
tokens, images, colors, text, and component inventories. Every command answers
one question a frontend dev asks.

## The "What" Pattern

Every command answers one question. Use this as a lookup table:

| Question | Command |
|---|---|
| What's the CSS for this frame? | `figma css --id <node-id> <url>` |
| What's the CSS for this whole page? | `figma css --recursive --id <node-id> <url>` |
| What are the design tokens? | `figma tokens <file-key>` |
| What are the Tailwind tokens? | `figma tokens --format tailwind <file-key>` |
| What colors are used here? | `figma colors --id <node-id> <url>` |
| What images/icons can I download? | `figma assets --output ./dir <url>?node-id=X` |
| How is this frame structured? | `figma layout <url-with-node-id>` |
| What's all the copy in this frame? | `figma texts <url-with-node-id>` |
| What's the text in layers with this name? | `figma texts --layer "Name" <url>` |
| What components exist here? | `figma components --id <node-id> <url>` |
| What layers match this name/type? | `figma find --name "icon" --type INSTANCE <url>` |
| What's this node? (details) | `figma inspect <url-with-node-id>` |
| What's this copied Figma comment? | `figma comments <url-with-comment-hash>` |
| What unresolved feedback affects this node? | `figma comments --include-ancestors --unresolved-only <url>` |
| What changed in the copy? | `figma diff text --from v1 --to v2 <url>` |
| What files are in this project? | `figma files <project-id-or-url>` |
| What projects exist? | `figma projects` |
| Is my token working? | `figma me` |

---

## Command Details

### `figma css` — Generate CSS

```bash
# Single frame
figma css --id 42:1 "https://www.figma.com/design/abc123/Name?node-id=42-1"

# Recursive (every child node gets a rule)
figma css --recursive --id 42:1 "url"

# Write to file
figma css --output styles.css --id 42:1 "url"
```

Layout maps to flexbox: `layoutMode → display:flex`, `itemSpacing → gap`,
`padding → padding`, fills → `background-color`. Text nodes emit
`font-family`, `font-size`, `font-weight`, `line-height`, `color`.
Properties are sorted deterministically.

### `figma tokens` — Design tokens

```bash
# CSS custom properties (default)
figma tokens "abc123"

# Tailwind theme.extend
figma tokens --format tailwind "abc123"

# JSON for style-dictionary
figma tokens --format json "abc123"

# With prefix
figma tokens --prefix brand- "abc123"  # → --brand-color-primary, etc.

# Pin source for CI determinism
figma tokens --source variables "abc123"   # Figma Variables
figma tokens --source styles "abc123"      # Published Styles
figma tokens --source scan "abc123"        # Raw node scan
```

Source order (`--source auto`, default): Variables → Styles → scan fallback.
When neither Variables nor Styles exist, scanning raw fills still produces
tokens (named by hex value). Pass `--scan-fallback=false` for named-only.

### `figma colors` — Color palette

```bash
figma colors --id 42:1 "url"
# → JSON: { "scope": {...}, "results": [{ "hex": "#FF0000", "opacity": 1.0, "name": "Red bg", "nodeId": "42:5" }] }
```

### `figma assets` — Download images, vectors, instances

```bash
# Everything
figma assets --output ./assets "url?node-id=42:1"

# Only vectors (icons)
figma assets --kind vector --format svg --output ./icons "url?node-id=42:1"

# Only images
figma assets --kind image --format png --output ./img "url?node-id=42:1"

# With JSON manifest
figma assets --json --output ./assets "url?node-id=42:1"
```

`--kind`: `all`, `image`, `instance`, `vector`.
`--format`: `auto` (default), `png`, `svg`.

Filenames: `lowercase-name_nodeid.png`.

### `figma export` — Export one node

```bash
figma export --format png --id 42:1 "url"
figma export --format svg --output ./icons/star.svg --id 42:1 "url"
```

Formats: `png`, `jpg`, `svg`, `pdf`.

### `figma layout` — Inspect frame structure

```bash
figma layout "https://www.figma.com/design/abc/Name?node-id=42-1"
# → { "scope": {...}, "result": <ordered tree with id, name, type, layoutMode, gap, padding, text, and children> }

figma layout --measure-spacing "url?node-id=42-1"
# → additionally reports measured spacing between adjacent layout children
```

Use this for copy/layout alignment when generated CSS is too implementation-oriented. Add `--measure-spacing` only when exact geometric sibling gaps are needed; default output stays compact. The URL node is inferred; use `--id` only with a bare file key or to override URL scope.

### `figma texts` — Extract text content

```bash
figma texts "https://www.figma.com/design/abc/Name?node-id=42-1"
figma texts --id 42:1 "abc123"
# → ordered JSON: { "nodeId": "42:1", "texts": [{ "id", "name", "text", "depth", "order", "parentName" }] }

figma texts --layer "Hero Title" "abc123"
figma texts --layer "Button Label" --recursive "abc123"
```

A selected node recursively returns all descendant text in Figma tree order. Without a node ID, `--layer` is required; `--recursive` includes descendants of each named layer.

### `figma find` — Search layers

```bash
# By name
figma find --name "button" "abc123"

# By type
figma find --type "COMPONENT" "abc123"

# Text matches include both layer name and actual copy
figma find --type "TEXT" "url?node-id=42-1"
# → { "scope": {...}, "results": [{ "id": "42:2", "name": "Stale layer name", "type": "TEXT", "text": "Actual copy" }] }

# Both, scoped to a node
figma find --id 42:1 --name "icon" --type "INSTANCE" "abc123"
```

At least one of `--name` or `--type` is required. Types: `FRAME`, `COMPONENT`,
`INSTANCE`, `SECTION`, `TEXT`, etc. Matching text nodes include `text`, so consumers should prioritize actual copy over potentially stale layer names.

### `figma components` — List components

```bash
figma components --id 42:1 "abc123"
figma components --id 42:1 --name "Button" "abc123"    # filter
figma components --id 42:1 --raw "abc123"               # raw nodes in the same {scope, results} envelope
```

### `figma inspect` — Node summary

```bash
figma inspect "https://www.figma.com/design/abc/Name?node-id=42-1"
figma inspect --id 42:1 "abc123" # explicit scope for a bare file key
# → JSON: { "scope": {...}, "result": { type, name, bounds, fills, strokes, text, children summary } }
```

### `figma comments` — Review feedback

```bash
# A copied comment URL returns that exact comment, regardless of node scope
figma comments "https://www.figma.com/design/abc?node-id=42-1&m=dev#1838610593"

# Include unresolved feedback attached to parent frames
figma comments --include-ancestors --unresolved-only "url?node-id=42-2"
```

Each comment includes a direct Figma `url`. `--recursive=false` limits normal node lookup to the selected node; hash lookup by comment ID takes precedence over node filtering.

### `figma diff text` — Copy changes between versions

```bash
figma diff text --from <version-id> --to <version-id> "abc123"
# → JSON: { added: [...], removed: [...], changed: [...] }
```

`--quiet` gives grep-style exit code: 0 = changes exist, 1 = none.

### Discovery commands

```bash
figma me                           # verify auth
figma projects                     # list projects
figma projects --team-id <id>      # filter by team
figma files <project-id-or-url>    # list files in a project
figma files --branches <id>        # include branch data
figma versions "abc123"            # file version history
```

---

## Workflow Recipes

### "I need CSS for this component"
```bash
figma css --output component.css "url?node-id=42:1"
```

### "I need the design tokens as CSS variables"
```bash
figma tokens --prefix brand- "abc123" > tokens.css
```

### "Download all the icons from this frame"
```bash
figma assets --kind vector --format svg --output ./icons "url?node-id=99:1"
```

### "Did the copy change since last sprint?"
```bash
figma diff text --from <sprint-start> --to <latest> "abc123"
```

### "Pre-commit: did copy change?"
```bash
figma diff text --from <last-version> --to <latest> --quiet "abc123"
# exit 0 → changes exist → don't merge blindly
```

### "List all button components in the design system"
```bash
figma find --type COMPONENT --name "button" "abc123"
```

---

## Node Scope Convention

Node-scoped commands infer `node-id` from a Figma URL. Use optional `--id` with a bare file key or to override URL scope. Single-node commands reject multiple IDs instead of silently choosing one; multi-node commands preserve selection order.

## Output Convention

All commands produce structured JSON (except `css`, `tokens`, `export`, and
non-`--json` `assets` which produce text). Pipe into `jq` for filtering:

```bash
figma inspect "https://www.figma.com/design/abc/Name?node-id=42-1" | jq '.result.fills'
figma find --name "hero" "abc123" | jq '.matches[].id'
```
