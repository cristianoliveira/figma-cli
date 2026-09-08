# Open improvement plans

These plans cover remaining gaps in Figma-to-browser visual checks. They describe
work to investigate, not features that are already available.

Source: [workflow feedback](../pains/figma-pixel-perfect-cli-feedback.md).
Completed work lives in [done](../done/), including crops, export metadata,
coordinate annotations, profiles, and reports. The
[AXI command-interface work](../done/axi/README.md) is complete.

## Suggested order

| Plan | Problem to solve |
| --- | --- |
| [Shape-mismatch evidence](pixel-perfect-shape-mismatch.md) | A lower raster score can hide a worse shape. Test whether reliable evidence can expose it. |
| [Actionable issue ranking](pixel-perfect-actionable-issue-ranking.md) | Agents need help choosing which mismatch to fix first. |
| [Transparent-crop suggestions](pixel-perfect-transparent-crop-suggestions.md) | Transparent padding can make capture bounds hard to compare. |

This is a suggested investigation order, not a declaration that each task blocks
the next. Read the task and current implementation before starting. Keep unfinished
tasks here and move verified work to `../done/`; update links in the same change.

## Write a useful plan

Each task should state:

1. **Problem:** observed pain and supporting evidence.
2. **Outcome:** the behavior or experiment needed to address it.
3. **Current ownership:** relevant packages, dependencies, and risks.
4. **Approach:** the smallest testable change, without fixing the design too early.
5. **Acceptance criteria:** behavior another developer can verify.
6. **Open decisions:** choices that depend on implementation evidence.

Do not treat a proposed flag or example output as an existing command contract.
Use [command contracts](../../docs/command-contracts.md) and executable help for
current behavior.

## Shared constraints

- Use Figma structure for node identity and design coordinates.
- Keep image measurements and coordinate transformations deterministic.
- Keep model descriptions optional and separate from validation gates.
- Never resize, align, or crop silently.
- Preserve JSON compatibility unless a change has an explicit migration.
- Use generated fixtures or local test servers instead of live credentials.
- Prove behavior with focused tests before moving a task to done.

## Completion goal

A developer can inspect a scoped design, export its logical bounds, compare a
browser capture, and trace each coordinate transformation in the result. Reports
must distinguish measured differences from suggested fixes and visual acceptance.
