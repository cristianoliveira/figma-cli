---
id: TASK-0010
title: Separate artifact persistence from output rendering
status: todo
depends_on: [TASK-0002, TASK-0005]
priority: normal
tags: [solid, output]
---

# Separate artifact persistence from output rendering

## Problem
The output package owns both serialization contracts and filesystem creation, while commands call file writes directly, mixing presentation with persistence responsibilities.

## Context
Keep `internal/output` responsible for result envelopes and TOON/JSON/text rendering. Move directory creation and file writes behind an artifact store used by commands/application services. Reuse the asset sink direction established by TASK-0005 where practical.

## Acceptance criteria
- [ ] `internal/output` contains no `os`, `filepath`, or file-permission policy.
- [ ] Artifact persistence is exposed through a narrow consumer-owned port with a filesystem adapter.
- [ ] Token, CSS, and export workflows receive persistence dependencies explicitly.
- [ ] Tests cover parent-directory creation, permissions, write failures, and successful result metadata.
- [ ] Raw stdout and JSON-wrapped file result behavior remain compatible.
- [ ] No generic repository or catch-all I/O interface is introduced.

## Notes
This is SRP at package level: rendering and durable storage change for different reasons.
