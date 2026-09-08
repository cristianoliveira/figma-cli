# Purpose

`scripts/` contains repository maintenance utilities: API generation support, live smoke orchestration, and measurement helpers.

# Boundaries

Scripts coordinate external tools and repository workflows; they are not runtime CLI capabilities. Generated API changes remain owned by [Figma transport](internal/figma/AGENTS.md) and its [generated package](internal/figma/api/AGENTS.md).

# Connections

- [Figma transport](internal/figma/AGENTS.md): consumes generated API artifacts produced by the generation workflow.
- [Internal packages](internal/AGENTS.md): smoke and measurement scripts exercise package boundaries without moving policy into scripts.

# Placement

Add a script only for repeatable repository maintenance or isolated smoke support. Keep reusable behavior in Go packages so commands and tests can share it.
