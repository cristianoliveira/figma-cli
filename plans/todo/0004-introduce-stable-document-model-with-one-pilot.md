---
id: TASK-0004
title: Introduce stable document model with one pilot
status: doing
depends_on: [TASK-0001]
priority: high
tags: [architecture, domain]
---

# Introduce stable document model with one pilot

## Problem
Pure extraction code reads Figma's untyped JSON shape directly, so external schema changes spread into the core and malformed data can fail silently.

## Context
Introduce only the node fields needed by one low-risk pilot, preferably `find`. The Figma adapter maps generated responses to the stable model. Expand the model from real use cases rather than reproducing the full OpenAPI schema.

## Acceptance criteria
- [ ] A small `internal/document` model represents the pilot's required node fields and children.
- [ ] Mapping from Figma API values happens inside `internal/figma` and reports malformed required data explicitly.
- [ ] The pilot capability accepts the stable model and contains no `map[string]any` document traversal.
- [ ] Fixture-backed mapper tests cover nested nodes, text, missing optional fields, and malformed required fields.
- [ ] Pilot command output and ordering remain unchanged.
- [ ] The model does not import generated API, HTTP, Cobra, filesystem, or output packages.

## Notes
This is a vertical slice, not a mandate to type every Figma field. Stop when the pilot is isolated and measured.
