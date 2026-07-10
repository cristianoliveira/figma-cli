# JSON style inspection

## Problem

Generated CSS is useful for implementation but class-oriented output derived from stale layer names is awkward for inspection. Frontend developer often needs raw design intent: layout, typography, colors, radius, effects, sizing, and variable/style bindings—without generated selectors.

## Solution

Add JSON-first style inspection for selected node/tree. Keep CSS generation as artifact command, and expose design properties in stable structured model for agents and frontend tooling.

## How

- Extend `inspect` or add `styles` command after evaluating command cohesion.
- Include typography, fills, strokes, radius, effects, opacity, sizing, alignment, style IDs, and variable bindings.
- Reuse extraction logic used by CSS/tokens rather than duplicate conversions.
- Include source node ID/name and omit absent fields.
- Remove or implement exposed `tokens --team`; do not advertise hard-failing feature.
- Add typed fixtures for text, frame, bound variable, shadow, and mixed fills.

## Success criteria

- Developer can consume design style without parsing CSS.
- CSS, tokens, and JSON inspection share one source of truth.
- No unsupported flags appear in help.
