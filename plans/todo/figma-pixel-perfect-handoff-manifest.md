# Consume export metadata and define a handoff manifest

Source feedback: [`../pains/figma-pixel-perfect-cli-feedback.md`](../pains/figma-pixel-perfect-cli-feedback.md)

This task defines desired behavior and verification, not a mandatory implementation. Inspect current code, write failing tests first, and preserve package boundaries described in repository guidance. Deterministic evidence must remain separate from advisory interpretation.

## Problem

Separate CLI outputs force consumers to reconstruct coordinate relationships.

## First increment

```bash
pixel-perfect reference.png actual.png \
  --reference-metadata rectangle.export.json
```

Use metadata only to apply an explicit, validated logical crop. Record applied transformation in result JSON.

## Longer-term outcome

Add a versioned Figma handoff artifact containing:

- selected node ID and absolute bounds;
- descendants with absolute and scope-relative bounds;
- export records and asset paths;
- effects and padding evidence;
- optional vector assets.

Possible command:

```bash
figma handoff --id 13576:15248 --output sidebar.handoff.json <file>
```

Pixel-perfect may use this to attach known Figma node identities to intersecting deterministic regions. This is metadata joining, not image-based semantic inference.

## Pre-analysis

- Inspect current inspect/export JSON contracts and pixel-perfect input orchestration before defining shared schema.
- Identify one owning package for manifest types so Figma and pixel-perfect do not duplicate schema.
- Treat path resolution, schema versioning, and coordinate transforms as compatibility boundaries.

## Project approach

- Start with a failing package-level test reproducing the feedback case.
- Keep command code responsible for parsing and orchestration; place behavior in the owning internal package.
- Preserve current output by default and make new fields or behavior explicit.
- Verify focused tests first, then run `go test ./...` and `golangci-lint run ./...`.
- Record any coordinate convention, schema decision, or heuristic limitation in command documentation.

## Acceptance criteria

- Schema version is mandatory and unsupported versions fail.
- Paths resolve relative to manifest location.
- Metadata dimensions must match actual input files.
- Every coordinate transformation appears in output JSON.
- Pixel metrics remain identical to equivalent explicit crop invocation.

## Implementation freedom

The implementer should decide whether `handoff` is a new command or an export/inspect mode after checking command conventions. Keep schema in one package rather than duplicating structs across CLIs.

---

