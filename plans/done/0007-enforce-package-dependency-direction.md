---
id: TASK-0007
title: Enforce package dependency direction
status: done
depends_on: [TASK-0001, TASK-0002]
priority: high
tags: [architecture, solid]
---

# Enforce package dependency direction

## Problem
Clean and SOLID boundaries are documented but not broadly executable, so future imports can silently reintroduce framework, transport, or generated-code coupling.

## Context
Add architecture fitness tests over Go imports after the new composition boundary exists. Enforce dependency direction, not folder naming. Generated code needs an explicit exemption because it is not hand-designed.

## Acceptance criteria
- [ ] Tests fail when domain/application packages import Cobra, HTTP, filesystem, environment, TOON, or generated Figma API packages.
- [ ] Tests fail when `internal/figma/api` is imported outside the Figma adapter.
- [ ] Tests fail when any internal package imports `cmd`.
- [ ] Allowed dependency rules are short, explicit, and documented in root `AGENTS.md`.
- [ ] Generated code and tests have narrow, justified exemptions.
- [ ] The check runs in the existing local/CI guardrail path.

## Notes
This protects DIP and package-level SRP. Do not encode a rigid layer taxonomy that blocks capability-oriented packages.
