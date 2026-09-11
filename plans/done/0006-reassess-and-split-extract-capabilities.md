---
id: TASK-0006
title: Reassess and split extract capabilities
status: done
depends_on: [TASK-0003, TASK-0004, TASK-0005]
priority: normal
tags: [architecture, cohesion]
---

# Reassess and split extract capabilities

## Problem
The extract package owns many unrelated transformations and has broad fan-in, but splitting it before stable ports and a document model would create churn without improving dependency direction.

## Context
After the pilot services and model exist, measure actual shared code and dependency direction. Split only capabilities with independent reasons to change. Shared typed traversal may move to `internal/document`; capability policy should remain vertical.

## Acceptance criteria
- [ ] A fresh import/fan-in and complexity report identifies real split candidates with evidence.
- [ ] Proposed package boundaries keep dependencies acyclic and point toward stable models/ports.
- [ ] At least the inspect, asset, and pilot capability boundaries have explicit ownership.
- [ ] Any selected move is behavior-preserving and performed in small commits with focused tests.
- [ ] `internal/extract` retains only cohesive shared transforms, or the report justifies keeping remaining code together.
- [ ] AGENTS.md architecture guidance is updated to match implemented boundaries.

## Notes
A package move alone is not success. Success is lower external coupling and clearer reasons for each package to change.
