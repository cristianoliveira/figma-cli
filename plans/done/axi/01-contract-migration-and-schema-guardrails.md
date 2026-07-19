# Contract migration and schema guardrails

## Problem

Commit `67ec36b` intentionally changed agent-visible contracts:

- collection envelopes gained `query`, `total`, and conditional `truncated`;
- collection commands default to 100 returned results and accept `--full`;
- pixel validation-gate failures now emit JSON on stdout before exiting 1;
- no-argument commands now emit compact readiness output;
- unknown flags now return compact nearest-flag guidance.

Tests cover implementations, but consumers do not yet have one concise contract reference or an explicit compatibility policy. A later refactor could accidentally change field meaning while keeping tests locally green.

## Proposed outcome

Create a versioned, human-readable command-contract reference plus focused schema fixtures that define semantics rather than incidental formatting.

### Collection contract

```json
{
  "scope": {"fileKey": "abc", "nodeIds": ["1:2"]},
  "query": {"name": "Button"},
  "total": 125,
  "truncated": true,
  "results": []
}
```

Required meanings:

- `scope`: normalized Figma scope actually queried;
- `query`: effective filters after defaults and aliases resolve;
- `total`: matched results before local output limiting;
- `results`: returned results after limiting;
- `truncated`: present and true only when returned count is lower than total;
- `--full`: disables local result limiting, not API pagination or unrelated depth controls.

### Pixel gate contract

- stdout remains valid comparison JSON on pass and fail;
- configured gates add `validation`;
- `validation.failed` includes every failed metric;
- stderr contains concise failure diagnosis;
- exit is 1 when any configured gate fails.

## Pre-analysis

- Generic envelopes live in `internal/output/contracts.go`.
- Command-specific envelopes still exist for texts, projects, files, versions, changes, and pixel output.
- Full-output assertions exist across `cmd/*_test.go`; use them selectively because giant snapshots are expensive to maintain.
- Binary contracts are covered in `tests/smoke`, which must run uncached.

## Delivery slices

### Slice 1 — inventory public contracts

1. List every command and classify output as query, detail, text, file, CSV, or gate result.
2. Record default limit/pagination/depth behavior.
3. Record exit behavior for success, empty, usage failure, operational failure, and gate/no-match semantics.
4. Identify intentional exceptions, such as grep-style quiet exits.

Deliverable: one table under `docs/` or command documentation linked from root README.

### Slice 2 — semantic contract fixtures

1. Add small golden JSON fixtures for generic query, filtered empty query, truncated query, and pixel gate failure.
2. Assert required fields and types.
3. Avoid snapshots of volatile URLs, timestamps, or complete raw Figma nodes.
4. Add a helper that rejects `null` where arrays are required.

### Slice 3 — compatibility policy

Document:

- additive fields are backward compatible;
- field removal, rename, type change, or semantic reuse requires an explicit versioned migration;
- default limit changes require release notes;
- consumers must not infer `total` from `len(results)`;
- `--full` is the compatibility escape hatch for previous unbounded behavior.

### Slice 4 — release note

Add a concise migration note with before/after examples and copyable commands. State which commands gained local limits.

## Test matrix

| Path | Required assertion |
|---|---|
| non-empty query | `total == len(results)` when not truncated |
| empty filtered query | exact effective query plus `total: 0`, `results: []` |
| truncated query | `total > len(results)`, `truncated: true` |
| `--full` | all local matches returned, no truncation |
| paginated API result | page cursor remains distinct from local total |
| pixel gate pass | JSON parseable, exit 0 |
| pixel gate fail | JSON parseable, stderr diagnostic, exit 1 |
| malformed CLI input | empty stdout, exit 2 |
| dependency failure | empty stdout unless domain contract says otherwise, exit 1 |

## Acceptance criteria

- One linked reference defines every public output family and exit convention.
- Contract tests fail on field type/meaning regressions, not whitespace.
- Migration note names default-limited commands and `--full` escape hatch.
- `go test -count=1 ./...` and lint pass.

## Risks

- Overspecified snapshots can freeze irrelevant representation details.
- Calling a page count `total` can mislead users; retain `count` for paginated version pages.
- Documentation can become a second source of truth. Task 04 must automate drift detection for exact flags/examples.

## Implementation freedom

Choose JSON Schema, typed fixture assertions, or compact golden files after measuring maintenance cost. Do not add a schema generator unless it removes real duplication.
