# Figma CLI decision reference

## Recommendation

Use this reference to choose between related commands. Use live help for exact arguments and flags:

```bash
figma <command> --help
```

Do not read every section before acting. Start with the user's concrete question.

## Choose the narrowest read

### Implementation

- `inspect`: implementation properties for one node, a bounded handoff, selected recursive fields, or exact vector paths.
- `layout`: ordered hierarchy, auto-layout behavior, measured spacing, or explicit responsive-frame comparison.
- `frames`: cheap screen-level frame discovery before deeper inspection.
- `css`: deterministic generated CSS when implementation-oriented output is the goal.

Prefer `inspect --handoff` for a bounded design-to-code summary. Prefer selected inspect fields when only a few properties matter. Use layout instead of CSS when the question is about structure rather than generated declarations.

### Content and discovery

- `texts`: ordered copy and Figma-provided list semantics.
- `find`: layers matching a name or type; text matches include actual copy.
- `components`: components, sets, instances, or usage grouped by exact component ID.
- `colors`: colors used in a selected tree.
- `tokens`: reusable variables, styles, or scan-derived tokens.
- `meta`, `projects`, `files`: navigate account and file hierarchy.

Use `frames` before recursive layout when the user first needs to identify a screen. Use `find` when the user knows a layer name or type but not its node ID.

### Assets

- `assets`: discover and download many images, instances, or vectors from a selected tree.
- `export`: render one exact selected node as PNG, JPG, SVG, or PDF.

Export the child node directly when its node ID is available. Do not export a parent and calculate a manual crop. Use either raster scale or target width, never both.

### Review and history

- `comments`: exact copied comment, unresolved feedback, authors, or node-scoped threads.
- `versions`: version IDs and pagination before a comparison.
- `changes`: frontend-relevant structural differences.
- `diff text`: copy differences.
- `diff blame`: version that introduced selected text.

Exact comment-ID lookup takes precedence over node filtering. Prefer node-scoped URLs for change analysis to reduce requests and output.

## Scope rules

- A full Figma URL can carry file and node scope.
- In URLs, use Figma's dash form: `?node-id=42-1`.
- With a bare file key, use explicit API-form scope such as `--id 42:1` when required.
- `--node` is an alias for `--id` where exposed by live help.
- Single-node commands reject ambiguous multiple IDs. Multi-node commands preserve caller order.
- Layer names can collide; keep all returned matches until identity is resolved by node ID.

## Output rules

- Structured stdout defaults to TOON.
- Use global `--json` for JSON-only tools:

```bash
figma --json inspect "https://www.figma.com/design/abc/Name?node-id=42-1" | jq '.result.fills'
figma --json find --name hero "abc123" | jq '.results[].id'
```

- Collection `total` describes matches before the local limit; it is not `results` length.
- `truncated: true` means output is incomplete. Follow its scope-preserving hint or choose an explicit limit.
- Bounded tree output reports traversal metadata separately from collection limits.
- CSS, token, export, and asset commands can intentionally write text or files.
- Structured errors use exit `2` for usage and exit `1` for operational failure.
- Quiet diff commands have documented grep-style status. Read their live help before using them in automation.

## High-value examples

### Inspect implementation properties

```bash
figma inspect --handoff --depth 4 "https://www.figma.com/design/abc/Name?node-id=42-1"
```

### Discover screens before deeper inspection

```bash
figma frames "https://www.figma.com/design/abc/Name?node-id=1-2"
```

### Download named vector icons

```bash
figma assets --kind vector --format svg --output ./icons \
  "https://www.figma.com/design/abc/Name?node-id=99-1"
```

### Compare copy between known versions

```bash
figma diff text --from <older-version-id> --to <newer-version-id> "abc123"
```

### Resolve authentication after a classified failure

```bash
figma me
```

A successful account lookup does not prove access to a specific file, team, comments endpoint, or Variables API.

## Non-negotiable interpretation rules

- Preserve returned vector path commands and winding rules; do not approximate them.
- Do not infer list semantics from glyphs when Figma supplies no list metadata.
- Do not guess responsive breakpoints from frame width; report observed frame evidence.
- Do not guess component identity from names when exact component IDs are absent.
- Do not claim complete results when truncation or bounded traversal says otherwise.
- Do not expose tokens, provider response bodies, stack traces, or private absolute paths.
