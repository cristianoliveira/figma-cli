# Bound `figma layout` trees

## Problem

`figma layout` recursively converts the selected Figma subtree with `extract.ExtractLayout`. Unlike collection commands, nested layout output has no depth bound, truncation evidence, or `--full` escape hatch. A large frame can flood agent context even when user needs only first implementation decisions.

Flat `--limit` behavior is wrong for a tree: slicing nodes destroys parent/child meaning and can hide where omission occurred.

## Proposed command contract

```bash
figma layout --depth 4 <url>   # default behavior
figma layout --depth 2 <url>   # shallower explicit view
figma layout --full <url>      # unbounded local traversal
```

Rules:

- selected root is depth `0`;
- default maximum descendant depth is `4`;
- `--depth` accepts zero or greater;
- `--full` and explicit `--depth` are mutually exclusive;
- validation occurs before loading Figma client;
- no silent omission: bounded output reports traversal metadata.

Suggested JSON:

```json
{
  "scope": {"fileKey": "abc", "nodeIds": ["1:2"]},
  "query": {"maxDepth": 4},
  "traversal": {
    "returnedNodes": 18,
    "totalNodes": 57,
    "truncated": true,
    "omittedNodes": 39
  },
  "result": {"id": "1:2", "children": []}
}
```

`totalNodes` counts nodes eligible for layout output before depth bounding, not every raw Figma node. `returnedNodes` counts serialized layout nodes including root.

## Pre-analysis

- Command orchestration: `cmd/layout.go`.
- Pure traversal: `internal/extract/layout.go`.
- Current recursion: `extractLayoutNode` has no depth parameter.
- `meaningfulLayoutNode` filters raw nodes after recursive child extraction, so counting/truncation must preserve existing meaningful-node semantics.
- Full Figma subtree is already fetched. Initial goal is token-bounded output, not reduced network payload.
- Existing layout spacing behavior must remain identical inside returned depth.

## Project approach

### Slice 1 — characterize current traversal

Write tests first for:

- root depth convention;
- nested meaningful nodes separated by non-meaningful containers;
- text leaf at exact depth boundary;
- spacing calculation among siblings inside bound;
- full extraction byte-equivalent to current `ExtractLayout` output.

Do not change command yet.

### Slice 2 — add pure bounded extraction

Introduce one pure extraction result containing tree plus traversal statistics. Preserve existing `ExtractLayout` as compatibility wrapper if internal callers need it.

Design constraints:

- one traversal should produce tree and counts when practical;
- depth omission must not fabricate children;
- meaningful descendants behind a non-meaningful container still affect total/omitted counts consistently;
- sibling spacing is calculated only between siblings actually represented at same returned level;
- no Cobra or output package dependency in `internal/extract`.

### Slice 3 — expose flags and output envelope

1. Add `--depth` default 4 and `--full` to `figma layout`.
2. Validate negative depth and conflicting flags before client load.
3. Map extraction result to stable command output.
4. Keep `--measure-spacing` orthogonal.
5. Add local examples showing bounded default and full escape.

### Slice 4 — binary contract and measurements

Create deterministic fixture with depth greater than default. Measure:

- default output bytes;
- `--depth 1` output bytes;
- `--full` output bytes;
- command round trips needed to discover omitted detail.

Default should materially reduce bytes while one copyable `--full` or deeper command recovers detail.

### Slice 5 — docs and skill handoff

Update command docs and leave exact skill/reference synchronization to Task 04 after contract settles.

## Test matrix

| Case | Expected |
|---|---|
| leaf root | one returned/total node, not truncated |
| child at max depth | child included |
| grandchild beyond max | omitted and counted |
| `--depth 0` | root only, explicit truncation when descendants eligible |
| `--full` | all eligible nodes, not truncated |
| `--full --depth 2` | usage error, exit 2, client not loaded |
| `--depth -1` | usage error, exit 2, client not loaded |
| spacing enabled | unchanged metrics inside returned tree |
| repeated run | byte-identical JSON |

## Acceptance criteria

- Default layout output is bounded at documented depth.
- Output explicitly states returned, total, omitted, and truncation.
- `--full` recovers previous unbounded local behavior.
- Depth and full errors occur before network access.
- Existing layout semantics and spacing tests remain green.
- Command help includes 2–3 copyable examples.

## Risks

- Counting raw nodes instead of meaningful output nodes makes metadata misleading.
- Pruning before evaluating meaningful descendants may incorrectly remove container paths.
- Traversing twice for total and output may increase CPU, though document is already local; measure before optimizing.
- This is a public output-envelope change and must follow Task 01 migration rules.

## Implementation freedom

Exact type names and whether extraction uses one or two passes are open. Do not flatten tree or reuse collection `--limit` simply for code reuse.
