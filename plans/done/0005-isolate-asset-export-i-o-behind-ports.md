---
id: TASK-0005
title: Isolate asset export I/O behind ports
status: done
depends_on: [TASK-0002]
priority: normal
tags: [architecture, assets]
---

# Isolate asset export I/O behind ports

## Problem
Asset export mixes filtering and naming policy with Figma transport, raw HTTP downloads, and filesystem creation, preventing independent reuse and substitution.

## Context
Split asset selection and deterministic manifest policy from infrastructure. Use narrow `AssetSource`, `ExportURLSource`, and `AssetSink` contracts owned by the asset application package. Keep Figma, HTTP, and filesystem implementations at the edge.

## Acceptance criteria
- [ ] Asset application workflow does not import `net/http`, `os`, or the concrete Figma client.
- [ ] Figma source and filesystem/HTTP sink implement consumer-owned ports.
- [ ] Unit tests cover filtering, URL failure, write failure, filename collisions, and request-order preservation using fakes.
- [ ] Empty requests perform no I/O and return a non-null empty manifest.
- [ ] Existing export paths, manifest shape, and partial-failure behavior remain compatible.
- [ ] Focused asset command, application, and adapter tests pass.

## Notes
Keep naming and collision policy in the application capability because they define deterministic user-visible behavior.
