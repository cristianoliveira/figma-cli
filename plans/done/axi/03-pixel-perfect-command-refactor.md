# Refactor pixel-perfect comparison orchestration

## Problem

`internal/pixelperfectcmd.newCommand` owns option parsing, validation, profile resolution, input preparation, comparison, annotations, overlays, offset/movement evidence, region enrichment, gates, reports, visual context, and rendering in one closure.

Current behavior is well tested, but function complexity is high and every new option risks changing validation order, side effects, or exit behavior.

## Proposed outcome

Split orchestration into named, testable stages while preserving CLI bytes, files, provider calls, and exit codes.

Target shape:

```text
Cobra flags
  -> parse + validate request        (pure, usage errors)
  -> prepare inputs                  (filesystem boundary)
  -> deterministic comparison       (imagediff boundary)
  -> enrich deterministic evidence  (annotations/regions/movement)
  -> evaluate gates                 (pure)
  -> write optional artifacts       (report/overlay)
  -> optional visual context        (provider boundary)
  -> render once                     (stdout boundary)
```

## Non-goals

- no new metrics or flags;
- no output-schema changes;
- no change to default mask naming;
- no change to advisory visual-context semantics;
- no movement of image algorithms into command package;
- no generic framework for all Cobra commands.

## Pre-analysis

- `newCommand` is roughly 300 lines with measured complexity near 92.
- Pure image behavior already belongs to `internal/imagediff`.
- Report rendering belongs to `internal/pixelperfectreport`.
- Provider behavior belongs to `internal/imagecontext`.
- Existing injectable `imageComparer` supports command tests.
- Validation-gate failure intentionally writes JSON before returning exit 1.
- Side-effect order matters: invalid options must not decode images or write artifacts.

## Delivery slices

### Slice 1 — characterization tests

Before extraction, lock down:

- successful JSON bytes for identical and changed fixtures;
- default mask path and explicit paths;
- invalid option causes no image read/artifact;
- malformed image remains operational exit 1;
- gate failure emits JSON then returns operational error;
- report/overlay collision errors occur before writes;
- visual provider invalid/unconfigured/configured paths;
- stdout/stderr cleanliness at real binary level.

Prefer small typed assertions plus a few golden outputs.

### Slice 2 — comparison request parser

Extract a pure function that reads resolved Cobra values into one request/options type and validates relationships.

It should own:

- thresholds and gate limits;
- region/ignore-region syntax;
- output/report/overlay path relationships;
- provider enum;
- region count controls;
- crop flag relationships;
- effective profile values.

It must return typed usage errors and perform no filesystem/network/image work.

### Slice 3 — input preparation boundary

Separate:

- loading optional metadata/profile files;
- image dimensions;
- crop validation and temporary files;
- cleanup ownership.

Cleanup must be idempotent and called exactly once. Missing files remain operational failures; malformed profile schema/version remains usage failure.

### Slice 4 — deterministic runner

Create a small orchestrator for deterministic stages only:

- compare;
- load/validate annotations;
- region grouping/filtering/limiting;
- region metrics/classification;
- suggested offset/movement;
- gate evaluation.

Dependencies should be configured once at composition root. Do not pass per-call configuration that belongs in injected clients.

### Slice 5 — optional side effects

Extract report, overlay, and visual-context stages behind narrow functions. Preserve ordering explicitly in tests.

Decision to verify: gate failures currently skip report and visual context but preserve comparison mask. Keep this behavior unless a separate contract change is approved.

### Slice 6 — thin Cobra closure

Final `RunE` should read like orchestration with early returns and one rendering decision. Suggested quality target:

- `newCommand` complexity below 25;
- main `RunE` under about 100 lines;
- no direct image algorithm loops in command setup;
- no duplicate option source of truth.

Treat targets as feedback, not reason to create empty pass-through abstractions.

## Test matrix

| Stage | Happy | Unhappy |
|---|---|---|
| request parse | defaults/profile/explicit precedence | invalid enum, ratio, path collision |
| input prep | same-size, valid crops, metadata | missing file, malformed metadata, crop bounds |
| deterministic run | identical, regions, offset/movement | unequal dimensions, bad mask/annotations |
| gates | pass at boundary | one and multiple failures with JSON stdout |
| artifacts | mask, overlay, report | unwritable path, collisions |
| visual context | configured and unconfigured advisory | unsupported provider before comparison |
| rendering | valid JSON | writer failure returned operationally |

## Acceptance criteria

- Existing command and smoke tests pass without expected-output rewrites except test organization.
- Runtime probes before/after show identical stdout, stderr, exit code, and artifacts for representative commands.
- Validation remains before dependencies.
- Complexity is materially reduced.
- Package direction remains command -> owning internal packages.
- No new abstraction lacks a stated behavior or test boundary.

## Risks

- Reordering side effects can break CI or leave partial artifacts.
- Over-generalizing request types may hide which stage owns validation.
- Testing only function return values can miss real Cobra stdout/stderr behavior.
- Giant golden files can make safe internal changes painful.

## Implementation freedom

Keep extracted orchestration in `internal/pixelperfectcmd` unless a reusable domain concept clearly belongs elsewhere. Prefer a few cohesive types over many one-method interfaces.
