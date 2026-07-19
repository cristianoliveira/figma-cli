# AXI strict-alignment initiative

Baseline: the completed [AXI follow-up](../../done/axi/README.md) established compact home views, stable JSON contracts, bounded Figma output, validation ordering, exit codes, and drift checks.

This initiative addresses remaining differences between current CLI behavior and the recommendations in the AXI skill.

## Problem

`figma` and `pixel-perfect` are strong agent-facing JSON/CSV CLIs, but strict AXI alignment remains ambiguous:

1. JSON or CSV is default; AXI recommends TOON for structured stdout.
2. usage and operational errors are plain text on stderr; AXI recommends structured errors on stdout.
3. no-argument views do not identify the resolved executable path.
4. `pixel-perfect probe` and `scan` can emit large unbounded output.
5. truncated results usually expose metadata without a copyable recovery hint.
6. raw filesystem or provider failures can expose implementation details.

The [compatibility profile](../../done/axi/06-strict-alignment-contract-profile.md) is decided: v2 makes TOON the default for structured stdout and errors; JSON is explicit compatibility output, while intentional CSV and artifact output remain unchanged.

## Recommended remaining order

The compatibility decision and TOON output migration are complete.

1. Add structured, redacted error output.
2. Bound pixel inspection output and improve truncation recovery.
3. Complete executable discovery and measure agent efficiency.

```text
contract decision ──> TOON migration (complete)
        │
        ├────────────> structured errors
        │
        └────────────> bounds and hints ──> discovery + measurement
```

## Shared delivery rules

- Start each behavior slice with a failing test.
- Preserve existing JSON compatibility unless migration is explicitly versioned.
- Keep domain values independent from TOON, JSON, CSV, or text rendering.
- Use a maintained, specification-compatible TOON library; never hand-roll encoding.
- Validate arguments before network, filesystem, image decoding, or provider calls.
- Keep exit codes stable: `0` success/no-op, `1` operational failure, `2` usage error.
- Never put progress, warnings, or debug logs inside structured result envelopes.
- Redact secrets, raw dependency output, and unnecessary absolute input paths.
- Preserve exact totals, non-null empty collections, and explicit truncation.
- Add hints only when they remove a likely follow-up call.
- Keep Cobra commands thin; rendering and error translation belong at shared boundaries.
- Run binary smoke tests uncached with `-count=1`.

## Initiative success criteria

- An explicit compatibility profile defines default formats, error channels, and migration policy.
- Structured domain values round-trip through TOON and JSON without semantic drift.
- Usage and operational errors are deterministic, actionable, redacted, and tested at process level.
- Every potentially large output is bounded by default or requires explicit selection.
- Truncated output gives total/returned counts and a copyable command to retrieve more.
- No-argument views identify the executable and provide useful state within a measured token budget.
- Representative command outputs require fewer bytes and fewer recovery round trips than baseline.

## Standard verification gate

```bash
goimports -w <changed-go-files>
git diff --check
golangci-lint run ./...
go test -count=1 ./...
go test -cover ./...
```

For executable contract changes, build both binaries and assert stdout, stderr, and exit status directly.
