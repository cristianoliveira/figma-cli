# changes (structural diff)

## Problem

Every frontend PR that follows a design update starts with the same question:
**"What actually changed in the Figma file?"**

Currently, the answer comes from:
- Designers manually listing changes in Slack/Notion (incomplete, vague)
- Devs visually comparing two frames side-by-side (slow, misses non-visual changes)
- The Figma version history UI — which shows a pixel diff but not *structural* changes
  (did a component get swapped? did a layer get renamed? did layout constraints change?)

The `diff text` command we already have only compares text content. Real design changes
include: node additions/deletions, component swaps, auto-layout property changes, style
bindings, layer reordering, renamed frames.

## Success criteria

- Given two versions of a Figma file, output a list of *structural changes*:
  - Nodes added, removed, or renamed
  - Component instances whose source changed
  - Auto-layout property changes (layoutMode, padding, gap, wrapping)
  - Style binding changes (a TEXT node changed from "Heading 1" to "Heading 2")
  - Bounds/position changes (did something move or resize?)
- Output is CI-readable (JSON) so it can be posted to a PR as a comment
- Performance: diff a large file (50+ frames) in under 10 seconds

## API mapping

| Data source | API endpoint |
|---|---|
| Version A of file | `/files/{key}?version=v1` |
| Version B of file | `/files/{key}?version=v2` |
| Node tree (both versions) | Parsed from the response's `document` node |

## CLI shape

```
figma changes <file-url> [flags]

Flags:
  --from        Version ID or date (default: previous version)
  --to          Version ID or date (default: current version)
  --format      Output format: summary, json (default: summary)
  --node-id     Diff only a specific node/frame (optional)
  --terse       Only show changed node paths, no property details
```

## Output example

```json
{
  "from": "2026-07-01",
  "to": "2026-07-07",
  "changes": [
    {
      "path": "Pages/Customization page",
      "type": "modified",
      "changes": [
        { "property": "children[3].name", "from": "Change status", "to": "Update status" },
        { "property": "children[3].fills[0].color", "from": "#0667C8", "to": "#0954A5" },
        { "property": "children[5].componentId", "from": "104:15187", "to": "104:15200" }
      ]
    },
    {
      "path": "Pages/Settings",
      "type": "added"
    }
  ]
}
```

## Approach

Two strategies, tradeoff between accuracy and speed:

### A. Structural diff (fast, pragmatic)
Walk both node trees in parallel. Compare node-by-node on:
- name, type, componentId, componentSetId
- Scalar properties: opacity, cornerRadius, layoutMode, padding, gap, fills
- Children: added/removed/reordered (by id, not index)
- Skip: exact pixel positions (noise), `guid` (irrelevant)

### B. JSON diff (brute force, accurate)
Fetch both versions, dump the full node tree as JSON, run a recursive JSON diff.
Accurate but noisy (every position change shows up), slower on large files.

**Recommendation:** Strategy A with a `--full` flag that falls back to B for power users.

## Edge cases

- Empty files (just created or archived) → "No previous version"
- Same version compared → "No changes"
- Large diffs (>1000 changes) → paginate or summarize by page
- Deleted nodes → reference children by ID, not index, to avoid misattribution
- Component instances in multiple places → group changes by component, not by location
