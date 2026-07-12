# Figma-to-browser pixel-perfect improvement plans

Source: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

Each problem is independently assignable. Implementers own detailed design after inspecting current contracts and creating failing tests.

## Recommended order

### Phase 1: coordinate truth

1. [Figma relative bounds](figma-relative-bounds.md)
2. [Pixel-perfect explicit crops](pixel-perfect-explicit-crops.md)
3. [Figma export metadata](figma-export-metadata.md)
4. [Figma/pixel-perfect handoff manifest](figma-pixel-perfect-handoff-manifest.md)

### Phase 2: review workflow

5. [Pixel-perfect HTML report](pixel-perfect-html-report.md)
6. [Suggested-offset interpretation](pixel-perfect-offset-interpretation.md)
7. [Agent preset](pixel-perfect-agent-preset.md)

### Phase 3: shape and structural context

8. [Figma vector path inspection](figma-vector-path-inspection.md)
9. [Figma component-role hints](figma-component-role-hints.md)
10. [Pixel-perfect shape-mismatch evidence](pixel-perfect-shape-mismatch.md)
11. [Pixel-perfect actionable issue ranking](pixel-perfect-actionable-issue-ranking.md)
12. [Transparent-crop suggestions](pixel-perfect-transparent-crop-suggestions.md)

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
