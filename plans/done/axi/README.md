# AXI follow-up initiative

Baseline: commit `67ec36b` hardened agent-facing command contracts for `figma` and `pixel-perfect`.

This initiative closes remaining Agent Experience Interface gaps without mixing behavior changes, architecture refactors, skill packaging, and live integration work into one risky change.

## Problem

Core AXI contracts now exist, but five risks remain:

1. `figma layout` can emit an unbounded nested tree.
2. newly added output fields and default collection limits need explicit migration and schema guardrails;
3. `pixel-perfect` comparison orchestration is correct but concentrated in one high-complexity Cobra closure;
4. installable Figma skill guidance can drift from live Cobra help and output contracts;
5. mocked HTTP tests do not prove current Figma permissions and live endpoint behavior.

## Recommended order

1. [Contract migration and schema guardrails](01-contract-migration-and-schema-guardrails.md)
2. [Bound `figma layout` trees](02-layout-tree-bounds.md)
3. [Refactor pixel-perfect orchestration](03-pixel-perfect-command-refactor.md)
4. [Synchronize skill guidance and add drift checks](04-skill-sync-and-drift-checks.md)
5. [Add opt-in live Figma smoke tests](05-live-figma-smoke-tests.md)
6. [Choose strict-alignment compatibility profile](06-strict-alignment-contract-profile.md)
7. [Add TOON output boundary](07-toon-output-migration.md)
8. [Add structured, redacted errors](08-structured-errors.md)
9. [Bound output and add truncation recovery](09-output-bounds-and-hints.md)

```text
contract guardrails ──> layout contract ──> skill sync
          │
          └────────────> pixel refactor

all stable local contracts ──────────────> live smoke
```

## Shared delivery rules

- Start each behavior slice with a failing test.
- Validate arguments before network, filesystem, image decoding, or provider calls.
- Preserve stdout for result data and stderr for diagnostics.
- Keep exit codes stable: `0` success/no-op, `1` operational failure, `2` usage error.
- Preserve empty arrays, effective query context, pre-limit totals, and explicit truncation.
- Do not silently reinterpret `--full`, `--limit`, or depth.
- Keep Cobra commands thin. Pure extraction stays in `internal/extract`; image measurement stays in `internal/imagediff`; rendering stays at one output boundary.
- Run smoke tests with `-count=1` because they build binaries in subprocesses.
- Never require live Figma credentials in normal pull-request tests.

## Initiative success criteria

- Every potentially large output is bounded or explicitly paginated by default.
- Output field semantics and compatibility policy are documented and protected by tests.
- Pixel comparison command orchestration is split into named, testable stages without changing output bytes, artifacts, or exit behavior.
- Skill examples and exact flag references cannot silently drift from Cobra contracts.
- Maintainers can run a read-only live smoke workflow without exposing credentials or making ordinary CI flaky.

## Standard verification gate

```bash
goimports -w <changed-go-files>
git diff --check
golangci-lint run ./...
go test -count=1 ./...
go test -cover ./...
```

For changed binaries, also build and probe stdout, stderr, and exit status directly.
