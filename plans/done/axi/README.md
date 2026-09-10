# Completed agent-interface work

This folder records the completed Agent Experience Interface (AXI) work for
`figma`. It started from commit `67ec36b`.

The task files preserve the problems, decisions, and acceptance criteria from that
work. They are historical records, not a list of current defects. For current
behavior, read [command contracts](../../../docs/command-contracts.md). For open
work, read [the todo index](../../todo/README.md).

## Completed tasks

| Task | Purpose |
| --- | --- |
| [01 — Contract guardrails](01-contract-migration-and-schema-guardrails.md) | Define output compatibility and protect schemas. |
| [02 — Layout bounds](02-layout-tree-bounds.md) | Bound nested layout output and report omissions. |
| [04 — Skill drift checks](04-skill-sync-and-drift-checks.md) | Keep agent guidance aligned with command behavior. |
| [05 — Live Figma smoke tests](05-live-figma-smoke-tests.md) | Add opt-in checks against real Figma endpoints. |
| [06 — Compatibility profile](06-strict-alignment-contract-profile.md) | Choose default formats, error channels, and migration rules. |
| [07 — TOON output](07-toon-output-migration.md) | Use TOON for structured output while retaining JSON compatibility. |
| [08 — Structured errors](08-structured-errors.md) | Return actionable errors without exposing secrets or raw dependency details. |
| [10 — Discovery and measurement](10-discovery-and-measurement.md) | Identify executables and measure output size and recovery steps. |

## Rules to keep when extending the CLI

- Start behavior changes with a failing test.
- Validate options before network, filesystem, image, or provider work.
- Keep structured results and error envelopes on stdout. Put diagnostics on stderr.
- Preserve documented exit codes and command-specific quiet-mode behavior.
- Report empty results, effective scope, pre-limit totals, and truncation explicitly.
- Keep local limits, traversal depth, and API pagination distinct.
- Keep commands focused on wiring. Put extraction, measurement, and rendering in
  their owning packages.
- Make live Figma tests opt-in. Normal tests must not need credentials.

## Verify a follow-up change

Run focused tests for the changed behavior first. Check whitespace with
`git diff --check`. Use the repository watcher or CI for full checks.

For example, from the repository root:

```bash
go test ./internal/cli -count=1
go test ./internal/output -count=1
```

When a command contract changes, also check the built binary's stdout, stderr,
and exit code. Run subprocess-based smoke tests with `-count=1` so cached test
results do not hide a changed binary.
