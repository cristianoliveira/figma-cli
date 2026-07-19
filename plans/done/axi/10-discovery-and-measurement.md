# Complete executable discovery and measure agent efficiency

## Problem

No-argument views are compact and useful, but they do not identify the resolved executable path as AXI recommends. Improvements are not yet evaluated by output size or command round trips, so token-efficiency claims would be speculative.

## Proposed outcome

Make no-argument views identify the executable using a resolved absolute path with home collapsed to `~`, show only cheap relevant state, and establish repeatable AXI measurements for representative workflows.

## Pre-analysis

- `cmd/root.go` reports Figma authentication readiness.
- `internal/pixelperfectcmd.NewCommand` reports static usage and next steps.
- root help and local examples are already tested for compactness.
- skill guidance and Cobra drift tests exist from completed AXI work.
- common agent decisions differ: Figma agents choose scope/query; pixel agents choose compare/probe/scan and artifact inspection.

## Project approach

1. Add failing tests for PATH-resolved execution, absolute invocation, symlinked binaries, and home-path collapse.
2. Define cheap live state per executable; never perform network calls on no-argument invocation.
3. Keep views within an explicit byte/token budget.
4. Build fixtures for common workflows: discovery, empty search, truncated search, export, identical image, failed gate, probe, scan, and recovery from invalid input.
5. Measure stdout bytes, stderr bytes, command count, and recovery round trips for default and compatibility formats.
6. Use evidence to decide where additional `--fields`, detail commands, or summaries remove follow-up calls.
7. Synchronize home-view and skill guidance from one source where practical; extend drift checks.

## Acceptance criteria

- no-argument output includes resolved executable path with `$HOME` represented as `~`.
- no-argument invocation performs no network, image decoding, or provider calls.
- state and next steps remain concise and command-specific.
- representative measurements are committed with reproducible commands.
- claims compare actual shapes and do not assert fixed TOON savings.
- any new `--fields` or summary view is justified by measured follow-up reduction.
- stale home/skill examples fail CI.

## Risks

- executable resolution differs across platforms and symlinks.
- dynamic state can make snapshots flaky; inject environment and path resolution.
- optimizing only byte count can remove decision-critical context.

## Implementation freedom

Choose measurement script location and budget after observing current outputs. Prefer a small deterministic test harness over a benchmark framework.


## Completion evidence

- Both no-argument views show canonical executable path with home collapsed to `~`.
- Resolver tests cover PATH-resolved absolute paths, direct absolute invocation, symlinks, and home boundary handling.
- No-argument unit and built-binary tests enforce cheap state, zero stderr, and compact byte budgets.
- `scripts/measure-axi.sh` reproducibly measures stdout/stderr bytes, exit status, commands, and recovery round trips without network calls.
- `docs/axi-measurements.md` records actual TOON/JSON/CSV shapes and decisions without fixed savings claims.
- Figma and pixel-perfect skill drift tests protect home examples and bounded-output facts.
