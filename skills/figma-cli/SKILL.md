---
name: figma-cli
description: >
  Query a Figma URL or file for layout, CSS, assets, text, components, comments, tokens, or history.
  Use for requests like "inspect this Figma"; requires FIGMA_ACCESS_TOKEN.
  Not for Figma editing, browser interaction, or implementing a UI from a screenshot.
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
   - exact vector contour truth: `figma inspect --include-vector-paths <vector-url>`
   - structure and spacing: `figma layout --depth 4 <url>`; use `--full` only after traversal metadata proves omitted detail is needed
   - generated styles: `figma css <url>`
   - copy: `figma texts <url>`
   - layer search: `figma find --name <name> <url>`
   - components: `figma components <url>`
   - assets: `figma assets --output <dir> <url>`
   - selected-node raster export: `figma export --format png --output <path> <url-with-node-id>`
   - raster export at a target width: `figma export --format png --width <pixels> --output <path> <url-with-node-id>`
   - tokens or colors: `figma tokens <file>` / `figma colors <url>`
   - comments: `figma comments <url>`
   - history: `figma versions <url>`, `figma changes`, or `figma diff text`
3. Prefer node-scoped URLs to reduce output and requests.
4. Prefer `inspect --format text --fields ...` when selected recursive node properties are enough. Structured output is TOON by default; add global `--json` only for interoperability or `jq` pipelines.
5. Report command, scope, result, and any missing permissions or ambiguity.

Read [command reference](references/commands.md) only when exact flags, output shape, or advanced behavior are needed.

## Scope and Output Rules

- Accept Figma file key or full URL where command supports both.
- Infer one node ID from URL. Use `--id` or `--node` for bare keys or explicit override. `figma export` exports that selected node directly; do not export a parent frame and manually calculate a child crop when child node ID is available.
- Reject unsupported multiple node IDs; never silently choose one.
- User-facing node IDs use `1-2`; API-facing IDs use `1:2`.
- Prefer default TOON structured envelopes. Global `--json` preserves compatibility JSON when needed. Collection `total` is pre-limit count, distinct from returned result count; use emitted copyable `hint` only when `truncated: true`, otherwise set `--limit` for a token budget. CSS, tokens, exports, and downloaded assets may produce deterministic text/files.
- Prefer bounded `layout --depth` and `inspect --depth` before full traversal. Empty queries retain scope and effective filters. Structured stdout errors use category `usage` with exit 2 or `operational` with exit 1; follow single `recovery` step when present.
- For PNG/JPG exports, use either `--scale` or target `--width`; the latter derives valid Figma scale from node bounds. Do not combine them.
- Use `inspect --include-vector-paths` when implementation depends on exact vector contour. Preserve returned fill/stroke path commands and winding rules; do not approximate shape from bounds or normalize path data. Recursive vector inspection requires explicit `--depth` because geometry payloads are large.
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
- “Open this website and click the login button.” → use browser tooling.
- “How should I structure generic React CSS?”

## Validation

- Command answers user question without unrelated full-file output.
- Node scope and ID normalization are explicit.
- Output or written files exist and are readable.
- Authentication/API failures are surfaced as redacted structured stdout errors, never omitted. Do not expect raw provider responses or credentials.
