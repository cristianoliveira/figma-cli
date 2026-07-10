# Stable output contracts

## Problem

Frontend scripts and agents need predictable JSON. `texts` changes top-level shape between layer and node modes, asset flows report failures differently by output mode, and CSS/tokens use text while most query commands use JSON. Every special case adds parsing branches and makes automation fragile.

## Solution

Give each command family a stable envelope while preserving appropriate payload formats. Query commands should have consistent metadata plus one result collection. Generated artifacts may remain text/files, but `--json` must always provide deterministic manifest metadata.

## How

- Inventory current output schemas and mark public fields.
- Define envelopes such as `{scope, matches}` for queries and `{artifact, manifest, errors}` for generation.
- Unify `texts` layer/node modes behind one result field with scope metadata.
- Make partial failures structured and visible in asset JSON manifests.
- Document schemas and add golden JSON tests before migration.
- Treat incompatible schema changes as intentional CLI contract changes, not accidental refactors.

## Success criteria

- One parser works across modes of same command.
- Failures cannot disappear from machine-readable output.
- Every command has a command-level output contract test.
