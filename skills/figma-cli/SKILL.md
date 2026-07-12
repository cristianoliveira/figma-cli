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
   - implementation handoff: `figma inspect --handoff <url>`
   - node details: `figma inspect <url>`
   - structure and spacing: `figma layout <url>`
   - generated styles: `figma css <url>`
   - copy: `figma texts <url>`
   - layer search: `figma find --name <name> <url>`
   - components: `figma components <url>`
   - assets: `figma assets --output <dir> <url>`
   - screenshot-aligned raster export: `figma export --format png --width <pixels> --output <path> <url>`
   - tokens or colors: `figma tokens <file>` / `figma colors <url>`
   - comments: `figma comments <url>`
   - history: `figma versions <url>`, `figma changes`, or `figma diff text`
3. Prefer node-scoped URLs to reduce output and requests.
4. Run command and filter structured JSON with `jq` only when narrower output helps.
5. Report command, scope, result, and any missing permissions or ambiguity.

Read [command reference](references/commands.md) only when exact flags, output shape, or advanced behavior are needed.

## Scope and Output Rules

- Accept Figma file key or full URL where command supports both.
- Infer one node ID from URL. Use `--id` for bare keys or explicit override.
- Reject unsupported multiple node IDs; never silently choose one.
- User-facing node IDs use `1-2`; API-facing IDs use `1:2`.
- Preserve stable JSON envelopes. CSS, tokens, exports, and downloaded assets may produce deterministic text/files.
- For PNG/JPG exports, use either `--scale` or target `--width`; the latter derives valid Figma scale from node bounds. Do not combine them.
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
