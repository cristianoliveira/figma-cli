---
id: TASK-0011
title: Audit and minimize capability contracts
status: doing
depends_on: [TASK-0006, TASK-0007, TASK-0008, TASK-0009, TASK-0010]
priority: low
tags: [solid, api]
---

# Audit and minimize capability contracts

## Problem
After package extraction, public functions and interfaces can remain broader than their consumers need, preserving accidental coupling despite improved folder boundaries.

## Context
Run this after the structural migrations. Review exported symbols, consumer-owned interfaces, fan-in, and change impact. Go interfaces should be small and defined by consumers; concrete domain values may remain exported when they are stable contracts.

## Acceptance criteria
- [x] Every application port has identified consumers and only methods those consumers require.
- [x] No broad global `Client`, `Repository`, `Service`, or dependency-bag interface crosses capability boundaries.
- [x] Unused exports and compatibility shims created during migration are removed.
- [x] Compile-time interface assertions exist at adapter boundaries where useful.
- [x] A fresh dependency graph has no cycles or reverse imports into commands/infrastructure.
- [x] SOLID review records SRP, OCP, LSP, ISP, and DIP evidence, including principles that do not apply materially in this Go design.

## Notes
Do not manufacture abstractions to satisfy acronyms. LSP work is required only if real substitutability failures appear.
