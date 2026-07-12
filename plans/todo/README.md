# Figma-to-browser pixel-perfect improvement plans

Source: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

Completed coordinate and workflow foundations live in [`../done/`](../done/), including relative bounds, explicit crops, export metadata, HTML reports, offset interpretation, comparison profiles, and generic coordinate annotations.

## Recommended remaining order

1. [Figma component-role hints](figma-component-role-hints.md)
2. [Pixel-perfect shape-mismatch evidence](pixel-perfect-shape-mismatch.md)
3. [Pixel-perfect actionable issue ranking](pixel-perfect-actionable-issue-ranking.md)
4. [Transparent-crop suggestions](pixel-perfect-transparent-crop-suggestions.md)

## Task-plan pattern

Every task describes:

1. **Problem** — observed developer pain, not a feature-shaped assumption.
2. **Proposed outcome** — user-facing contract or experiment boundary.
3. **Pre-analysis** — current project ownership, likely dependencies, and risks to inspect.
4. **Project approach** — TDD and package-boundary expectations without prescribing exact implementation.
5. **Acceptance criteria** — behavior another developer can verify.
6. **Implementation freedom** — decisions intentionally left to implementer after evidence.

## Shared constraints

- Figma structure and metadata are authoritative for node identity.
- Pixel metrics and transformations are deterministic and reproducible.
- LLM descriptions are optional, advisory, and excluded from CI gates.
- No silent resizing, alignment, or cropping.
- Preserve existing JSON compatibility unless change is explicitly versioned.
- Tests use generated fixtures or local test servers, never live credentials.

## Initiative completion criteria

A developer can inspect scoped Figma structure without manual coordinate subtraction, understand logical versus exported bounds, compare browser captures without external cropping tools, review all evidence in one report, and reproduce every coordinate transformation from output JSON.
