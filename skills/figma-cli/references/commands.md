# Figma CLI command reference

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
| Which component instances does this screen use? | `figma components --usage <url-with-node-id>` |
| What layers match this name/type? | `figma find --name "icon" --type INSTANCE <url>` |
| What's this node? (details) | `figma inspect <url-with-node-id>` |
| What should I implement from this node? | `figma inspect --handoff <url-with-node-id>` |
| What's this copied Figma comment? | `figma comments <url-with-comment-hash>` |
| What unresolved feedback affects this node? | `figma comments --include-ancestors --state open <url>` |
| What changed structurally? | `figma changes --from v1 --to v2 <url>` |
| What changed in the copy? | `figma diff text --from v1 --to v2 <url>` |
| How different are two PNG screenshots? | `pixel-perfect reference.png actual.png --output diff.png` |
| What is this file about? | `figma meta <url>` |
| What versions exist? | `figma versions <url>` |
| When did this text appear? | `figma diff blame --to <version> <url>` |
| What files are in this project? | `figma files <project-id-or-url>` |
| What projects exist? | `figma projects [team-url-or-id]` |
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
`padding → padding`, solid/linear-gradient fills → `background`, paint alpha →
CSS color alpha, node opacity → `opacity`, clipped content → `overflow:hidden`,
and effects → `box-shadow`, `filter`, or `backdrop-filter`. Text nodes emit
`font-family`, `font-size`, `font-weight`, `line-height`, `color`, and
`text-decoration`.
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

# Discover and export icon instances/vectors with clean names
figma assets --kind icon --name "icon" --format svg --filename name \
  --trim-name-prefix "Iconography / " --output ./icons "url?node-id=42:1"

# Only standalone vectors
figma assets --kind vector --format svg --output ./icons "url?node-id=42:1"

# Only images
figma assets --kind image --format png --output ./img "url?node-id=42:1"

# With JSON manifest
figma assets --json --output ./assets "url?node-id=42:1"

# Keep successful downloads when only some exports fail
figma assets --allow-partial --output ./assets "url?node-id=42:1"
```

`--kind`: `all`, `icon` (instances + vectors), `image`, `instance`, `vector`.
`--format`: `auto` (default), `png`, `jpg`, `svg`, `pdf`.
`--name` filters layer names case-insensitively. `--filename name` removes node IDs; collisions receive deterministic `-2`, `-3` suffixes. Use explicit `--trim-name-prefix` rather than relying on guessed naming conventions.

Default filenames remain `lowercase-name_nodeid.ext`.

### `figma export` — Export one node

```bash
figma export --format png --id 42:1 "url"
figma export --format png --scale 2 --id 42:1 --output ./node@2x.png "url"
figma export --format svg --output ./icons/star.svg --id 42:1 "url"
figma export --format png --id 42:1 --output ./node.png --metadata ./node.export.json "url"
```

Formats: `png`, `jpg`, `svg`, `pdf`. Use `--scale` only for raster exports (`png`/`jpg`), default `1`, range `0.01-4`. Use `--metadata` for visual-diff workflows. The sidecar records `nodeBounds`, measured `exportBounds`, `scale`, `dimensionDelta`, `paddingEvidence`, and when derivable `logicalCrop` / `exportPadding`. For PNG exports with vector/effect padding, metadata may derive logical crop from a temporary SVG export so `pixel-perfect --reference-metadata` can avoid manual crop math.

### `figma layout` — Inspect frame structure

```bash
figma layout "https://www.figma.com/design/abc/Name?node-id=42-1"
# → { "scope": {...}, "result": <ordered tree with id, name, type, layoutMode, gap, padding, text, and children> }

figma layout --measure-spacing "url?node-id=42-1"
# → additionally reports measured spacing between adjacent layout children

# Compare explicit responsive frames in supplied order
figma layout compare "abc123" --id 100:1 --id 200:1 --id 300:1

# Resolve exact, unique sibling/descendant names under a selected section
figma layout compare "section-url?node-id=50-1" --name Desktop --name Tablet --name Mobile
# → { "scope": {...}, "variants": [...], "transitions": [{ "fromId", "toId", "changes" }] }
```

Use this for copy/layout alignment when generated CSS is too implementation-oriented. Add `--measure-spacing` only when exact geometric sibling gaps are needed; default output stays compact. Compare never guesses breakpoints: caller order is authoritative, names must match uniquely, and reported frame widths are design evidence rather than declared CSS breakpoints. The URL node is inferred; use `--id` only with a bare file key or to override URL scope.

### `figma texts` — Extract text content

```bash
figma texts "https://www.figma.com/design/abc/Name?node-id=42-1"
figma texts --id 42:1 "abc123"
# → ordered JSON: { "scope": {...}, "results": [{ "id", "name", "text", "nodeKind", "depth", "order", "parentName", "lines" }] }

figma texts --layer "Hero Title" "abc123"
figma texts --layer "Button Label" --recursive "abc123"
```

A selected node recursively returns all descendant text in Figma tree order. `lines` appears only when Figma provides line/list metadata and includes stable indexes, list type, and indentation; glyphs such as `•` are never treated as inferred lists. Mixed text exposes style override IDs and their typography metadata. Without a node ID, `--layer` is required; `--recursive` includes descendants of each named layer.

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
figma components --id 42:1 --name "Button" "abc123"    # filter by name
figma components --id 42:1 --kind instance "abc123"     # component | set | instance
figma components --id 42:1 --raw "abc123"               # all raw descendant nodes
figma components --usage "url?node-id=42-1"              # group instances by exact component ID
```

Default results contain only `COMPONENT`, `COMPONENT_SET`, and `INSTANCE` nodes. Results include node path, variant properties, instance properties, property definitions, component ID, and component-set ID when available. `--usage` returns counts and occurrence details grouped strictly by `componentId`; detached instances without an ID are omitted rather than guessed by name. `--raw` without `--kind` preserves full node traversal.

### `figma inspect` — Node summary

```bash
figma inspect "https://www.figma.com/design/abc/Name?node-id=42-1"
figma inspect --id 42:1 "abc123" # explicit scope for a bare file key
figma inspect --recursive "url?node-id=42-1" # implementation specs for entire selected tree with relativeBounds
figma inspect --recursive --annotations-output frame.annotations.json "url?node-id=42-1"
figma inspect --handoff "url?node-id=42-1"   # bounded implementation specs + component usage
figma inspect --handoff --depth 2 --include-hidden "url?node-id=42-1"
# → default JSON: { "scope": {...}, "result": { type, name, bounds, paints, layout, effects, componentProperties, propertyDefinitions, styleBindings, resolvedStyles, variableBindings, resolvedVariables } }
# → recursive JSON: { "scope": {...}, "results": [{...}, {...}] }
# → handoff JSON: { "scope": {...}, "result": { "nodes": [{...}], "components": [{ "name", "componentId", "count" }] } }
```

Use `--handoff` as the design-to-code default: it limits traversal to depth 4, excludes invisible descendants, and summarizes repeated component instances. `--recursive` remains the unbounded flat implementation inventory and cannot be combined with `--handoff`. Scoped recursive inspect includes `relativeBounds` measured from the requested scope root while preserving absolute `bounds`; use these for local CSS coordinates instead of manual subtraction. Recursive inspect also includes `spacingFromPrevious` for measured auto-layout gaps between direct visible non-absolute siblings. It reports `parentId`, `previousId`, `axis`, `measured`, `declared`, and `matchesDeclared`; missing Figma `itemSpacing` is treated as declared `0`. Use it to catch collapsed text heights or missing DOM spacing before chasing raster offsets.

`--annotations-output` requires `--recursive` and writes neutral screenshot-relative bounds for selected scope and descendants. Use output with `pixel-perfect --annotations <path>` to attach Figma node IDs and labels to intersecting mismatch regions. This is optional context: it must not change pixel metrics, gates, or exit status. Explicit `--depth` and `--include-hidden` also control annotation traversal.

Raw binding IDs are always preserved. `resolvedStyles` adds style name/type from node metadata. `resolvedVariables` adds variable and collection names when the Variables API is accessible; it is omitted without failing when metadata access is unavailable. Components and instances expose variants, property values, and property definitions. Mixed text exposes style override IDs and typography metadata. Prefer this over a separate handoff/spec command so implementation properties keep one source of truth.

### `figma comments` — Review feedback

```bash
# A copied comment URL returns that exact comment, regardless of node scope
figma comments "https://www.figma.com/design/abc?node-id=42-1&m=dev#1838610593"

# Include unresolved feedback attached to parent frames
figma comments --include-ancestors --state open "url?node-id=42-2"
figma comments --state resolved --author "Ada" <file-url>
figma comments --after 2026-01-01T00:00:00Z <file-url>
```

Comments use the stable `{scope, results}` envelope. Each result is a thread with `root` and chronologically ordered `replies`; anchored roots include `node_path` when node scope is available. Each comment includes a direct Figma `url`. `--recursive=false` limits normal node lookup to the selected node; hash lookup by comment ID takes precedence over node filtering.

### `figma changes` — Structural changes between versions

```bash
figma changes --from <version-id> --to <version-id> "abc123"
figma changes --from <version-id> --to <version-id> "url?node-id=42-1"
figma changes --from <version-id> --to <version-id> --terse --limit 100 "abc123"
figma changes --from <version-id> --to <version-id> --quiet "abc123"
# → JSON: { "scope": {...}, "from": "...", "to": "...", "total": 42, "truncated": 0, "changes": [{ "id", "path", "type", "nodeType", "changes": [...] }] }
```

Reports added/removed/renamed nodes, component swaps, auto-layout changes, style binding changes, bounds changes, and child reordering. `--terse` omits property details. `--limit` bounds emitted nodes while `total` and `truncated` preserve diff size. `--quiet` uses grep-style status: 0 means changes exist, 1 means none. Prefer a node-scoped URL for focused PR review and faster requests.

### `figma diff text` — Copy changes between versions

```bash
figma diff text --from <version-id> --to <version-id> "abc123"
# → JSON: { added: [...], removed: [...], changed: [...] }
# Added/removed/changed entries include readable parent `path` when available.
```

`--quiet` gives grep-style exit code: 0 = changes exist, 1 = none.

### `figma diff blame` — Explain text provenance

```bash
figma diff blame --to <version-id> "url?node-id=42-1"
figma diff blame --from <older-version-id> --to <version-id> "url?node-id=42-1"
```

Reports the earliest inspected version where each current text value appeared.
Use a node-scoped URL for focused results. `--from` bounds how far back to search.

### Discovery commands

```bash
figma me                              # verify auth
figma projects                        # list projects from authenticated teams
figma projects <team-url-or-team-id>  # select one team
figma files <project-id-or-url>       # list files in a project
figma files --branches <id>           # include branch data
figma meta "abc123"                   # compact file metadata
figma versions "abc123"               # file version history
figma versions --page-size 50 "abc123"
figma versions --after <version-id> "abc123"   # older versions
figma versions --before <version-id> "abc123"  # newer versions
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

Node-scoped commands infer `node-id` from a Figma URL. Use optional `--id` with a bare file key or to override URL scope; `--node` is accepted as a cross-command alias for `--id`. Single-node commands reject multiple IDs instead of silently choosing one; multi-node commands preserve selection order.

## Output Convention

All commands produce structured JSON (except `css`, `tokens`, `export`, and
non-`--json` `assets` which produce text). Pipe into `jq` for filtering:

```bash
figma inspect "https://www.figma.com/design/abc/Name?node-id=42-1" | jq '.result.fills'
figma find --name "hero" "abc123" | jq '.results[].id'
```
