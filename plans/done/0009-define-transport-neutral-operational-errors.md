---
id: TASK-0009
title: Define transport-neutral operational errors
status: done
depends_on: [TASK-0003, TASK-0005]
priority: normal
tags: [solid, errors]
---

# Define transport-neutral operational errors

## Problem
CLI error rendering imports concrete Figma, HTTP, environment, and filesystem errors, so presentation policy changes when infrastructure implementations change.

## Context
`internal/cli/error_output.go` currently recognizes `figma.ResponseError`, HTTP statuses, `env.ErrTokenNotSet`, and `os.PathError`. Define a small application-facing classification contract and translate infrastructure failures at adapter boundaries.

## Acceptance criteria
- [ ] CLI error rendering does not import `internal/figma`, `net/http`, `internal/env`, or filesystem error types.
- [ ] Adapter errors retain causes for `errors.Is`/`errors.As` and safe diagnostics never expose tokens or response bodies.
- [ ] Stable categories cover authentication, authorization, rate limiting, unavailable dependency, invalid input, and artifact access.
- [ ] Unknown failures retain the generic operational fallback.
- [ ] Tests cover every classification and wrapped-error behavior.
- [ ] Existing exit codes, messages, recovery hints, and structured envelopes remain compatible.

## Notes
Keep classification data-focused. Do not create an inheritance hierarchy or provider-specific error types in the application core.
