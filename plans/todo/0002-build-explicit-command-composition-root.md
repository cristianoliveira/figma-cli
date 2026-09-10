---
id: TASK-0002
title: Build explicit command composition root
status: doing
depends_on: []
priority: high
tags: [architecture, cli]
---

# Build explicit command composition root

## Problem
Package-level commands, init registration, and mutable flag globals distribute dependency wiring and make isolated construction harder.

## Context
`cmd` currently wires commands through package globals and `init()`. Create dependencies in `cmd/figma/main.go`, pass them into a root factory, and let each command own instance-local flag state. This is composition work, not a command behavior rewrite.

## Acceptance criteria
- [ ] `cmd/figma/main.go` is the sole production composition root.
- [ ] Root construction receives explicit dependencies and returns a complete command tree.
- [ ] Production command registration does not depend on Go `init()` ordering.
- [ ] Command flags do not bind to mutable package-level variables.
- [ ] Two independently constructed roots can execute in one test without state leakage.
- [ ] Existing help, exit-code, error, and smoke contracts remain unchanged.

## Notes
Migrate command factories mechanically. Do not introduce a service locator or one broad dependency interface.
