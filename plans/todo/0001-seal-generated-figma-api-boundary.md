---
id: TASK-0001
title: Seal generated Figma API boundary
status: todo
depends_on: []
priority: high
tags: [architecture, figma]
---

# Seal generated Figma API boundary

## Problem
Generated OpenAPI types leak into commands and capability packages, so upstream schema regeneration can force changes outside the Figma adapter.

## Context
`internal/figma/api` is generated infrastructure. Production imports currently escape into `cmd/files.go`, `cmd/projects.go`, `cmd/versions.go`, and `internal/comments`. Map responses to stable application-facing DTOs inside `internal/figma`; do not change CLI output contracts.

## Acceptance criteria
- [ ] No production Go file outside `internal/figma/**` imports `internal/figma/api`.
- [ ] `files`, `projects`, `versions`, and comments consume stable non-generated values.
- [ ] Mapper tests cover optional fields and timestamp formatting at the adapter boundary.
- [ ] Existing command output remains byte-compatible in JSON and semantically compatible in TOON.
- [ ] An automated architecture test fails if a generated API import escapes again.
- [ ] Focused `cmd`, `internal/comments`, and `internal/figma` tests pass.

## Notes
Keep generated files unchanged. This task creates the first enforceable dependency rule and should land before domain-model work.
