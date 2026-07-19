# Bound pixel inspection output and improve truncation recovery

## Problem

Most Figma collections and layout trees are bounded, and pixel mismatch regions default to 20. However, `pixel-perfect probe` and `scan` can still emit output proportional to selected points or image color runs. Truncated structured responses usually report truncation without a copyable recovery command.

## Proposed outcome

Make every potentially large output bounded or explicitly selected. Return total, returned count, and truncation state. When truncation occurs, include one scope-preserving command that retrieves more or all results.

## Pre-analysis

- shared Figma limits live in `cmd/result_limit.go` and `internal/output/contracts.go`.
- layout uses depth and traversal metadata.
- comparison regions use `--max-regions` and `--full`.
- probe supports repeated points, lines, step, and radius.
- scan compresses pixels into runs, but alternating colors can still produce one run per pixel.
- inspect text already emits a useful `--full` hint and can guide the structured contract.

## Project approach

1. Add failing tests for exact-limit, over-limit, empty/no-change, `--full`, invalid limit, and mutually exclusive flags.
2. Define totals before local limiting for probe points and scan runs.
3. Add bounded defaults based on measured representative screenshots.
4. Preserve axis, crop, and input-coordinate metadata after limiting.
5. Generate recovery commands from parsed command scope rather than raw shell strings.
6. Preserve relevant flags and use placeholders only for unknown values.
7. Apply same conditional hint contract to truncated Figma collections and comparison regions.
8. Keep normal, non-truncated output free from hints.

## Acceptance criteria

- probe and scan cannot flood stdout by default.
- output distinguishes total matches/runs from returned count.
- truncation is explicit and includes a copyable recovery command.
- recovery command preserves input scope, filters, crop, axis/point selection, and format.
- `--full` or explicit limit behavior is validated before image decoding/network access.
- empty and no-change outputs remain explicit.
- existing untruncated JSON/CSV compatibility is covered according to task 01.

## Risks

- limiting scan runs can hide transitions; totals and continuation/full behavior must remain clear.
- embedding shell commands requires safe quoting.
- hints can bloat output; emit only on truncation.

## Implementation freedom

Choose pagination, `--limit`, or compact summaries after measuring output and common agent decisions. Do not silently sample data.
