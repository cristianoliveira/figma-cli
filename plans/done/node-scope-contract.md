# Consistent node scope

## Problem

Frontend developers move between pasted Figma URLs and node IDs copied from other commands. Node-scoped commands currently differ on URL inference, `--id` overrides, missing scope, and multiple node IDs. Some reject multiple nodes while `texts` and `export` can silently use only the first, risking work against wrong design node.

## Solution

Define one node-scope contract for every relevant command:

1. Infer `node-id` from URL.
2. Let optional `--id` override URL scope for exploration with bare file key.
3. Declare command as single-node or multi-node.
4. Reject unsupported multiple selections instead of silently dropping nodes.
5. Use same errors and help language.

## How

- Add shared resolver returning explicit single/multiple scope types in `internal/figma`.
- Migrate `inspect`, `layout`, `texts`, `find`, `components`, `colors`, `css`, `assets`, `export`, comments, and diff commands.
- For single-node commands, return `requires exactly one node ID`.
- For multi-node commands, preserve caller node order and identify node in each result.
- Add table-driven contract tests for URL inference, override, missing scope, hyphen normalization, and multiple IDs.

## Success criteria

- Same URL/`--id` behavior across all node-scoped commands.
- No command silently chooses first node.
- Root and skill documentation describe one reusable rule.
