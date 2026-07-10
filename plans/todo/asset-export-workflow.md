# Consistent asset export workflow

## Problem

Single `export` and bulk `assets` support different formats and multiple-node behavior. Bulk failures may only appear on stderr while stdout lists successful files, making incomplete handoff easy to miss.

## Solution

Align format rules and make every export auditable through deterministic manifest containing successes, failures, requested node IDs, output paths, and formats.

## How

- Share format validation between `export` and `assets`.
- Explicitly support or reject JPG/PDF in bulk with documented rationale.
- Apply shared node-scope contract; never silently export first selected node.
- Return non-zero status for partial failure unless explicit `--allow-partial` is provided.
- Always write/emit manifest in JSON mode; keep useful paths in human mode plus clear failure summary.
- Add mocked download tests for full success, partial failure, unsupported format, and filename collisions.

## Success criteria

- Same format vocabulary where API supports it.
- Partial exports cannot look fully successful.
- Re-running produces deterministic filenames and manifest.
