# Command contracts

This is the versioned public-output reference for `figma` and `pixel-perfect`.
Source code and command tests remain executable contract truth; this document
states the stable semantics consumers may rely on.

## Compatibility policy

- Structured results emit TOON by default. Global `--json` selects the established indented JSON contract without changing domain fields.
- Intentional text, CSV, and file artifact outputs remain unchanged unless `--json` explicitly requests an existing compatibility envelope.
- Adding an optional field is backward compatible.
- Removing, renaming, changing a field type, or reusing a field with a new
  meaning requires an explicit versioned migration.
- A changed default result limit requires a release note.
- Consumers must not derive `total` from the returned collection length.
- `--full` is the compatibility escape hatch when a collection command has a
  local result limit. It does **not** disable API pagination or traversal-depth
  controls.

## Exit and stream convention

| Outcome | stdout | stderr | exit |
| --- | --- | --- | --- |
| Success, including empty query | command result | diagnostics only | 0 |
| Invalid arguments or flags | no result data | concise correction | 2 |
| Dependency or operational failure | no result data, unless a result was deliberately rendered first | actionable diagnosis | 1 |
| Pixel validation gate failed | structured comparison result | failed-metric diagnosis | 1 |

Result data belongs on stdout. Diagnostics, including Cobra errors, belong on
stderr. Commands do not print raw dependency errors, credentials, or progress
in structured result output.

## TOON migration

Structured domain values pass through one output boundary. TOON is default;
global `--json` retains prior JSON bytes, field names, omission rules, and text/file
envelopes. Command-local `--format` flags still select intentional artifacts or
views such as CSV, CSS, image formats, and recursive inspect text.

Measured on a deterministic 2×2 pixel comparison with one changed pixel, default
TOON was 927 bytes and compatibility JSON was 1,254 bytes (26% fewer bytes for
this shape). Savings depend on data shape; no fixed reduction is guaranteed.

## Collection envelope

Commands using the generic query envelope return this domain shape (shown in default TOON):

```toon
scope:
  fileKey: abc
  nodeIds[1]: "1:2"
query:
  name: Button
total: 125
truncated: true
results[0]:
```

| Field | Meaning |
| --- | --- |
| `scope.fileKey` | Figma file actually queried. |
| `scope.nodeIds` | Normalized API-form node IDs actually queried; always an array. |
| `query` | Effective filters after defaults and aliases resolve; omitted only where no filter applies. |
| `total` | Number of local matches before the command's result limit. Never lower than `results.length`. |
| `truncated` | Present and `true` only when local limiting returned fewer values than `total`. |
| `results` | Returned values after local limiting; always an array. |

A filtered empty result therefore has `total: 0`, `results: []`, and its
effective `query`; it is a successful result, not an absent response.

## Commands and output families

| Commands | Output family | Local limit / relevant controls |
| --- | --- | --- |
| `me`, `meta`, `inspect`, `layout compare`, `diff blame` | structured detail | command-specific node, version, or frame scope |
| `layout` | bounded structured tree (`scope`, effective `query`, `traversal`, `result`) | `--depth` defaults to 4; `--full` disables only the local tree-depth bound |
| `find`, `colors`, `components`, `comments`, `frames`, `inspect`, `texts` | structured query | `--limit` defaults to 100; `--full` disables only that local limit |
| `projects`, `files` | legacy structured collection (`projects`/`files`, `total`, optional `truncated`) | `--limit` defaults to 100; `--full` disables only that local limit |
| `versions` | structured pagination | `count` is page count; cursor/page navigation is distinct from local query totals |
| `assets`, `export` | files plus metadata JSON where requested | explicit output path and node scope |
| `css`, `tokens` | deterministic text or generated file; global `--json` envelope available | format-specific controls |
| `changes`, `diff text` | structured change analysis | explicit `--from` and `--to` versions |
| `pixel-perfect <reference> <actual>` | structured comparison and optional image/report artifacts | explicit crops, masks, profiles, and optional validation gates |
| `pixel-perfect probe`, `pixel-perfect scan` | CSV by default or JSON with `--format json` | explicit points, rows, columns, and sampling controls |

For `layout`, the selected root is depth `0`; nodes at `--depth` are included.
`traversal.totalNodes` counts eligible layout nodes before the local depth bound,
while `returnedNodes`, `omittedNodes`, and `truncated` make omissions explicit.
`--full` and an explicit `--depth` are mutually exclusive.

## Migration: bounded collection results

Collection commands listed above now emit at most 100 local results by default.
Use `total` to learn how many matched, `truncated: true` to identify a partial
response, and `--full` when a caller intentionally needs every locally matched
value.

Before, consumers commonly assumed every match was returned:

```bash
figma find --name Button <figma-url>
```

After migration, choose an explicit bounded or complete request:

```bash
# Default: at most 100 results, with total and truncation evidence.
figma find --name Button <figma-url>

# Compatibility escape hatch: disable only local result limiting.
figma find --full --name Button <figma-url>
```

Do not combine `--full` with `--limit`; this is a usage error. `--full` does
not retrieve API pages that the underlying endpoint has not supplied, and it
does not override `--depth` or other command-specific bounds.

## Pixel validation gates

When one or more maximum metric flags are configured, comparison output adds a
`validation` object. It is emitted both when gates pass and when they fail.
On failure, `validation.failed` contains every failed metric, stdout remains
valid TOON (or JSON under `--json`), stderr gives a concise diagnosis, and the command exits `1`.
