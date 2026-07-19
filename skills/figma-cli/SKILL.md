---
name: figma-cli
description: >
  Query a Figma URL or file for layout, CSS, assets, text, components, comments, tokens, or history.
  Use for requests like "inspect this Figma"; requires FIGMA_ACCESS_TOKEN.
  Not for Figma editing, browser interaction, screenshot comparison, or iterative UI matching.
---

# Figma CLI

## Objective

Answer one concrete design question with structured, frontend-ready Figma data.

## Workflow

1. Verify authentication with `figma me` when token state is unknown.
2. Identify user question and choose smallest command that answers it:
   - implementation handoff: `figma inspect --handoff --depth <n> <url>`
   - bounded recursive node details: `figma inspect --recursive --depth <n> <url>`
   - token-light selected outline: `figma inspect --recursive --depth <n> --format text --fields name,type,relativeBounds,layout.mode,layout.gap,fills <url>`
   - pixel-diff coordinate context: `figma inspect --recursive --annotations-output <path> <url>`
   - exact vector contour truth: `figma inspect --include-vector-paths <vector-url>`
   - structure and spacing: `figma layout --depth 4 <url>`; use `--full` only after traversal metadata proves omitted detail is needed
   - generated styles: `figma css <url>`
   - copy: `figma texts <url>`
   - layer search: `figma find --name <name> <url>`
   - components: `figma components <url>`
   - assets: `figma assets --output <dir> <url>`
   - selected-node raster export: `figma export --format png --output <path> <url-with-node-id>`
   - screenshot-aligned raster export: `figma export --format png --width <pixels> --output <path> <url-with-node-id>`
   - tokens or colors: `figma tokens <file>` / `figma colors <url>`
   - comments: `figma comments <url>`
   - history: `figma versions <url>`, `figma changes`, or `figma diff text`
3. Prefer node-scoped URLs to reduce output and requests.
4. Prefer `inspect --format text --fields ...` over ad hoc Python/JQ tree formatting when selected recursive node properties are enough. Otherwise filter structured JSON with `jq` only when narrower output helps.
5. Report command, scope, result, and any missing permissions or ambiguity.

Read [command reference](references/commands.md) only when exact flags, output shape, or advanced behavior are needed.

## Scope and Output Rules

- Accept Figma file key or full URL where command supports both.
- Infer one node ID from URL. Use `--id` or `--node` for bare keys or explicit override. `figma export` exports that selected node directly; do not export a parent frame and manually calculate a child crop when child node ID is available.
- Reject unsupported multiple node IDs; never silently choose one.
- User-facing node IDs use `1-2`; API-facing IDs use `1:2`.
- Preserve stable JSON envelopes. Collection `total` is pre-limit count, distinct from returned result count; use `--full` only when `truncated: true` proves it is needed, otherwise set `--limit` for a token budget. CSS, tokens, exports, and downloaded assets may produce deterministic text/files.
- Prefer bounded `layout --depth` and `inspect --depth` before full traversal. Empty queries retain scope and effective filters; usage errors exit 2 and dependency failures exit 1.
- For PNG/JPG exports, use either `--scale` or target `--width`; the latter derives valid Figma scale from node bounds. Do not combine them.
- Use `inspect --include-vector-paths` when implementation depends on exact vector contour. Preserve returned fill/stroke path commands and winding rules; do not approximate shape from bounds or normalize path data. Recursive vector inspection requires explicit `--depth` because geometry payloads are large.
- Export metadata uses pixel-aligned `logicalCrop` and `contentInset` relative to the exported image. Prefer these values over Figma canvas coordinates when preparing screenshot comparisons.
- For pixel-perfect loops, write generic annotations from same selected frame with `inspect --recursive --annotations-output`. Pass artifact to `pixel-perfect --annotations`; annotations add Figma node IDs/names to mismatch regions without changing metrics. Ensure annotation coordinate-space dimensions match prepared reference image.
- Layer names are not unique; return every match with node ID.
- Exact comment ID lookup takes precedence over node filtering.
- Never edit Figma or claim CLI can mutate design files.

## Routing Checks

Should trigger:
- “Inspect this Figma frame and give me implementation specs.”
- “Download SVG icons from this Figma URL.”
- “What copy changed between these Figma versions?”
- “List components used by this Figma screen.”
- “Generate tokens from this Figma file.”

Should not trigger:
- “Compare these two PNG screenshots.” → use `pixel-perfect`.
- “Keep changing this page until it matches Figma.” → use `figma-pixel-perfect-loop`.
- “Open this website and click the login button.” → use browser tooling.
- “How should I structure generic React CSS?”

## Validation

- Command answers user question without unrelated full-file output.
- Node scope and ID normalization are explicit.
- Output or written files exist and are readable.
- Authentication/API failures are surfaced, never omitted.
