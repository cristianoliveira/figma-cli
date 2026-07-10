# Copy list semantics

## Problem

Raw `characters` text does not tell frontend developer whether copy is one multiline block, separate list items, or styled ordered/unordered lines. This causes bullet intent and content structure to be missed during copy alignment.

## Solution

Enrich ordered copy output with Figma-provided line/list metadata without guessing semantic HTML. Preserve distinction between separate text nodes and lines inside one text node.

## How

- Extract `lineTypes`, line indentation, paragraph spacing, style overrides, and related text metadata available in API response.
- Return normalized per-line data only when metadata exists.
- Include `nodeKind: textBlock` and stable line indexes.
- Never infer list semantics from glyphs like `•` when Figma metadata is absent.
- Add fixtures for unordered list, ordered list, plain multiline text, mixed styles, and separate list-item nodes.
- Validate on real frame containing bullets.

## Success criteria

- Consumers can distinguish list block from separate text nodes.
- Ordered/unordered intent is explicit when Figma provides it.
- Plain text output remains concise.
