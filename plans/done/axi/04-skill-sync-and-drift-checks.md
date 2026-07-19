# Synchronize skill guidance and add drift checks

## Problem

The installable Figma skill under `.pi/skills/figma-cli/` routes agents and documents commands. Live Cobra help is command-contract truth, but skill prose and `references/commands.md` can silently retain old flags, unbounded examples, or stale output assumptions.

Recent AXI changes introduced `--limit`, `--full`, totals, query context, compact no-arg readiness, and stricter exit classification. Task 02 will add a layout depth contract. Manual synchronization alone will drift again.

## Proposed outcome

Keep narrative routing hand-written, but derive or validate exact command facts from live CLI behavior. CI fails when skill examples or exact flag references stop matching Cobra.

## Pre-analysis

- Skill entry: `.pi/skills/figma-cli/SKILL.md`.
- Detailed reference: `.pi/skills/figma-cli/references/commands.md`.
- Skill currently recommends `figma layout <url>` without a bound.
- Cobra help already contains examples and flag defaults.
- Root README also describes command families; avoid creating three independent exact flag tables.
- Pi skill trigger language should change only when routing behavior changes, not for every CLI flag.

## Source-of-truth decision

Use this hierarchy:

1. Cobra definitions/tests: command syntax, flags, defaults, examples, exit behavior.
2. Generated or validated command reference: exact facts.
3. SKILL.md: routing, safe workflow, progressive disclosure, domain rules.
4. README: human overview and links.

Do not generate narrative workflow prose from Cobra help.

## Delivery slices

### Slice 1 — drift inventory

Compare every skill/reference command and flag against built `figma --help` and local command help. Classify lines as:

- exact fact suitable for generation/validation;
- narrative recommendation requiring review;
- obsolete command or behavior;
- example requiring a fixture/placeholder check.

### Slice 2 — update AXI workflow guidance

After Task 02 settles, update skill guidance to state:

- collection commands default to bounded output;
- use `--full` only when total/truncation proves it is needed;
- use explicit `--limit` for token budgets;
- use layout/inspect depth before full traversal;
- interpret `total` separately from returned result count;
- preserve query scope and filters when reporting empty results;
- usage errors are exit 2, dependency failures exit 1.

Keep SKILL.md lean; move exact flag tables to reference.

### Slice 3 — implement drift checker

Build both binaries or use Cobra command constructors in a test. Validate at minimum:

- every documented command exists;
- every exact long flag exists on documented command;
- documented default values match help where stated;
- 2–3 primary examples parse to intended command without network access;
- removed/deprecated flags fail the check;
- generated reference is unchanged after regeneration, if generation is chosen.

Prefer structured command inspection in Go over regex-parsing formatted help when practical.

### Slice 4 — CI integration

Add one deterministic, credential-free check to normal CI and pre-commit only if runtime remains small. Provide a regeneration command with actionable failure output.

Example failure:

```text
skill drift: `figma layout --depth` exists in Cobra but layout reference has no depth/default entry
run: go generate ./internal/skilldocs
```

Do not silently rewrite skill files during CI.

### Slice 5 — trigger checks

If skill description/routing triggers change, run trigger tests from skill design guidance. Flag/reference-only updates do not need trigger changes.

## Acceptance criteria

- Skill workflow uses bounded-first commands.
- Exact command/flag/default facts match live CLI.
- One command detects drift locally and in CI.
- Drift error identifies file, command, mismatch, and recovery command.
- No Figma token is required.
- Narrative skill quality remains manually reviewable and compact.

## Risks

- Parsing rendered help is sensitive to Cobra formatting upgrades.
- Generating whole SKILL.md would destroy carefully designed routing language.
- Requiring every new flag in skill creates noise; validate only declared exact references and required AXI facts.
- Project-local skill may differ from separately installed global skill; document packaging path.

## Implementation freedom

A Go test, generator, or small script is acceptable. Justify any new schema by showing which duplicated facts it replaces.
