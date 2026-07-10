# Frontend command contract tests

## Problem

Workflow-heavy commands lack command-level tests. Extractor tests protect traversal logic but do not catch broken flags, URL inference, envelopes, help text, printer behavior, or accidental output changes.

## Solution

Add deterministic command contract tests around user-facing frontend workflows using injected/mock HTTP clients and golden JSON where appropriate.

## How

- Build reusable Cobra test harness with isolated flags, stdout/stderr, environment, and test server.
- Cover `css`, `export`, `comments`, `find`, `layout`, `components`, `colors`, and `tokens`.
- Test happy and unhappy paths: URL node inference, `--id` override, missing scope, multiple nodes, API errors, empty results, file output, and JSON mode.
- Keep API fixtures minimal and deterministic.
- Run command tests in CI and measure changed-package coverage without relying on live Figma token.

## Success criteria

- Every user-facing command has at least one happy and one failure contract test.
- No test requires network or real token.
- Public output changes require intentional golden updates.
