---
id: TASK-0003
title: Move inspect workflow behind application ports
status: done
depends_on: [TASK-0001, TASK-0002]
priority: high
tags: [architecture, inspect]
---

# Move inspect workflow behind application ports

## Problem
The inspect Cobra handler owns application orchestration and depends on a concrete Figma client, which couples reusable behavior to delivery and transport details.

## Context
Move fetch, scope, extraction, style/variable enrichment, limiting, and failure policy into an inspect application service. Keep flag parsing and rendering in Cobra. Define narrow interfaces in the inspect consumer package; adapt Figma at the composition root.

## Acceptance criteria
- [ ] An inspect service exposes a context-aware request/result API with no Cobra dependency.
- [ ] The service depends on consumer-owned ports, not `*figma.Client`.
- [ ] Tests cover single, recursive, handoff, missing-node, transport-failure, and optional variable-enrichment failure paths.
- [ ] `cmd/inspect.go` only maps flags/arguments to a request and renders the returned result.
- [ ] Inspect output, result limiting, and current enrichment behavior remain compatible.
- [ ] Inspect service tests require no HTTP server and command tests require no Figma adapter.

## Notes
Preserve current policy where variable enrichment is best-effort unless a separate product decision changes it.
