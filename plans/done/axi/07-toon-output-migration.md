# Add a TOON output boundary and migration path

## Problem

Structured results default to JSON, while AXI recommends compact TOON. Adding TOON inside individual commands would duplicate rendering logic and couple domain values to serialization.

## Proposed outcome

Add specification-compatible TOON at the shared output boundary. Preserve JSON as an explicit compatibility format. Change defaults only through the migration policy selected in task 01.

## Pre-analysis

- `internal/output/printer.go` owns common rendering.
- `internal/output/contracts.go` owns generic Figma envelopes.
- pixel comparison still has command-specific output types.
- text and file artifacts need different treatment from structured values.
- current `--json` behavior wraps text/file outputs and must not become ambiguous with `--format`.

## Project approach

1. Read the current TOON specification before selecting a library.
2. Add failing tests for scalars, nested values, empty collections, escaping, declared array lengths, deterministic map ordering, and semantic round-trip.
3. Define one format resolver shared by both binaries.
4. Encode domain values only at the output boundary.
5. Specify interaction between `--format`, existing `--json`, command-specific CSV/text, and file outputs.
6. Add compatibility tests proving JSON schemas remain unchanged.
7. Add process smoke tests for representative Figma query, empty result, truncated result, pixel comparison, and validation-gate failure.
8. Measure TOON and JSON bytes for representative shapes; publish measurements without claiming universal savings.

## Acceptance criteria

- No hand-written TOON concatenation or regex parser exists.
- TOON output is deterministic and semantically round-trips to original domain value.
- Empty arrays, nested values, strings, and counts follow current TOON specification.
- JSON compatibility output remains byte/schema compatible where promised.
- text, CSV, and file artifact behavior is explicit and non-ambiguous.
- progress and diagnostics never contaminate structured stdout.
- help and skill examples document selected default and compatibility flags.

## Risks

- A global format flag can conflict with command-local `--format` flags.
- TOON may be larger for some shapes.
- changing default without versioned rollout can break existing agents.

## Implementation freedom

Choose flag names and migration timing from task 01. Prefer removing ambiguity over preserving overlapping flags indefinitely.


## Completion evidence

- Shared `internal/output.Printer` renders JSON-shaped values as deterministic TOON by default.
- Existing global `--json` remains compatibility opt-in for structured values and text/file envelopes.
- Both binaries use same boundary for structured comparison and query output.
- TOON tests cover nesting, empty arrays, escaping, declared lengths, deterministic output, and semantic round-trip.
- Command tests protect default TOON and compatibility JSON.
- Deterministic 2×2 pixel comparison measured 927 TOON bytes versus 1,254 JSON bytes.
