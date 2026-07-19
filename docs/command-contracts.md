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
| Invalid arguments or flags | structured usage error | empty by default | 2 |
| Dependency or operational failure | structured redacted error | empty by default | 1 |
| Pixel validation gate failed | structured comparison result | failed-metric diagnosis | 1 |
| Intentional grep-style no-match | empty | empty | documented non-zero |

Result and error envelopes belong on stdout. Progress, debug diagnostics, and
the deliberate pixel validation diagnosis belong on stderr. Commands do not
print raw dependency errors, credentials, absolute private paths, or stack traces.

Error envelopes have one stable shape:

```toon
error:
  category: usage
  message: unknown flag: --nide
  input: --nide
  exitCode: 2
  recovery: Run `figma inspect --help` for valid flags.
```

`category` is `usage` or `operational`. `input` appears only when offending input
is safely known. `recovery` is omitted when no specific, safe action is known. Global `--json` renders same fields as compatibility JSON.

## Executable discovery

No-argument views report canonical absolute executable path with home directory
collapsed to `~`. They show only cheap authentication or usage state and perform
no network requests, image decoding, or provider calls. See
[`axi-measurements.md`](axi-measurements.md) for reproducible byte and round-trip evidence.

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
| `pixel-perfect probe`, `pixel-perfect scan` | bounded CSV by default or JSON with `--format json` | `--limit` defaults to 25 points or runs per image; `--full` disables local bound |

For `layout`, the selected root is depth `0`; nodes at `--depth` are included.
`traversal.totalNodes` counts eligible layout nodes before the local depth bound,
while `returnedNodes`, `omittedNodes`, and `truncated` make omissions explicit.
`--full` and an explicit `--depth` are mutually exclusive.

## Migration: bounded collection results

Collection commands listed above now emit at most 100 local results by default.
Use `total` and `returned` where available to distinguish matches from emitted
items. `truncated: true` identifies a partial response; truncated output alone
includes a scope-preserving `hint` command ending in `--full`. Normal output
omits hints.

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
does not override `--depth` or other command-specific bounds. Probe limits
points; scan limits color runs independently for reference and actual images.
Truncated CSV appends one `# total=... returned=... hint=...` metadata line.

## Pixel validation gates

When one or more maximum metric flags are configured, comparison output adds a
`validation` object. It is emitted both when gates pass and when they fail.
On failure, `validation.failed` contains every failed metric, stdout remains
valid TOON (or JSON under `--json`), stderr gives a concise diagnosis, and the command exits `1`.
