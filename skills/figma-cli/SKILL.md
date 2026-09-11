---
name: figma-cli
description: >
  Query a Figma URL or file for layout, CSS, assets, text, components, comments, tokens, or history.
  Use for requests like "inspect this Figma"; requires FIGMA_ACCESS_TOKEN.
  Not for Figma editing, browser interaction, or implementing a UI from a screenshot.
---

# Figma CLI

## Recommendation

Start with the CLI's current command index:

```bash
figma --help
```

Choose the smallest command that answers the question, then inspect only its live contract:

```bash
figma <command> --help
```

Run that command directly. Do not call `figma me` as a routine preflight.

## Why this works

1. **Small commands reduce output.** Ask one design question per call.
2. **Narrow scope reduces requests.** Prefer a URL containing the target `node-id`.
3. **Live help prevents drift.** Treat Cobra help as syntax truth; use this skill for routing and cross-command decisions.

## Choose by outcome

| Needed outcome | Start with |
| --- | --- |
| Implementation handoff or node properties | `figma inspect` |
| Ordered structure, spacing, or responsive comparison | `figma layout` |
| Screen-level frame discovery | `figma frames` |
| Generated CSS | `figma css` |
| Copy or list semantics | `figma texts` |
| Layer search | `figma find` |
| Components or instance usage | `figma components` |
| Downloadable assets | `figma assets` |
| One rendered node | `figma export` |
| Design tokens or color inventory | `figma tokens` or `figma colors` |
| Review comments | `figma comments` |
| Version history or design changes | `figma versions`, `figma changes`, or `figma diff` |
| File, project, or account discovery | `figma meta`, `figma files`, `figma projects`, or `figma me` |

Read [decision reference](references/commands.md) only when choosing between related commands or handling scope, output, and failure semantics.

## Decision rules

### Minimize scope and output

- Prefer node-scoped Figma URLs.
- Use bounded depth or result limits before requesting full traversal.
- Use `inspect --format text --fields ...` when selected recursive properties are enough.
- Trust `total`, `truncated`, traversal metadata, and emitted hints before requesting more data.
- Layer names are not unique. Preserve every match and its node ID.

### Preserve Figma identity

- URL node IDs use Figma's `1-2` form. Explicit `--id` and API output may use `1:2`.
- Never silently choose one node when a command rejects multiple IDs.
- Use exact vector-path inspection when contour matters. Do not infer paths from bounds.
- Never edit Figma or claim this CLI can mutate a design file.

### Handle output deliberately

- Structured output is TOON by default.
- Put global `--json` before the command when JSON tooling is required: `figma --json inspect ...`.
- Text and file-producing commands may intentionally return artifacts instead of structured stdout.
- Check command exit status before consuming output. Respect documented grep-style non-zero results.

### Recover only after failure

- Run the selected command first.
- After an authentication failure, use `figma me` only to isolate token configuration from resource permissions.
- Follow one structured `recovery` step when present.
- Report safe error category and missing permission; never expose credentials or raw provider responses.

## Validation

- Result answers the user's question without unrelated full-file output.
- Scope and normalized node IDs are explicit.
- Truncation is acknowledged before treating results as complete.
- Written artifacts exist and are readable.
- Usage and operational failures are not interpreted as domain results.
