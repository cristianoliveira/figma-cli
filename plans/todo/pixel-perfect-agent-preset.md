# Agent preset

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

Good agent-oriented flags are verbose and inconsistently selected.

## Proposed outcome

```bash
pixel-perfect reference.png actual.png --agent --output comparison.mask.png
```

Preset expands to documented defaults, for example threshold, region grouping, minimum pixels, offset search, overlay, report, and optional visual context. Explicit user flags always override preset values.

Output records resolved values and which came from preset.

## Pre-analysis

- `internal/pixelperfectcmd`: resolve configuration before execution rather than scattering conditional defaults.
- Consider a typed options structure to stop `RunE` from growing.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

- Test every preset value.
- Test explicit override precedence.
- Artifact names derive safely from `--output`; no accidental overwrite.
- Missing provider configuration produces disclaimer, not comparison failure.

## Guardrails

`--agent` is convenience only. It must not change metric algorithms or turn LLM output into validation evidence.

---

