# Open improvement plans

These plans cover remaining gaps in Figma inspection and implementation handoff.
They describe work to investigate, not features that are already available.

The [AXI command-interface work](../done/axi/README.md) is complete.

There are currently no open plans.


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
- Preserve JSON compatibility unless a change has an explicit migration.
- Use generated fixtures or local test servers instead of live credentials.
- Prove behavior with focused tests before moving a task to done.

## Completion goal

A developer can inspect a scoped design and export the implementation inputs needed
for a Figma-based UI. Results must preserve scope and distinguish design facts from
agent interpretation.
