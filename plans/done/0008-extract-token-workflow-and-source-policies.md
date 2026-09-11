---
id: TASK-0008
title: Extract token workflow and source policies
status: done
depends_on: [TASK-0002, TASK-0003, TASK-0004]
priority: normal
tags: [solid, tokens]
---

# Extract token workflow and source policies

## Problem
Token source selection, fallback policy, diagnostics, Figma access, formatting, file writing, and Cobra handling are combined in one command package, giving the workflow several reasons to change.

## Context
Create a token application service with narrow variable, style, node, diagnostic, formatter, and artifact ports. Model source selection as explicit policy so adding a source does not expand the Cobra handler. Preserve deterministic Variables -> Styles -> optional scan behavior.

## Acceptance criteria
- [ ] Token source selection and fallback policy live outside `cmd` and have no Cobra dependency.
- [ ] The application service does not depend on `*figma.Client`, `os.Stderr`, or direct filesystem functions.
- [ ] Explicit source and auto-fallback paths are tested independently, including provider failure and empty results.
- [ ] Diagnostics are returned as structured application results rather than written as hidden side effects.
- [ ] Adding a token source requires an adapter/policy registration, not edits across command, transport, and formatting code.
- [ ] CSS, Tailwind, JSON, stdout, and file output contracts remain compatible.

## Notes
OCP does not mean removing every switch. A closed switch over stable output formats is acceptable; isolate only source policy that is expected to grow.
