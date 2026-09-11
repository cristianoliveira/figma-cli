# Purpose

`internal/document` owns the stable, transport-neutral Figma document tree model and pure tree-level value coercion.

# Boundaries

Keep this package independent of Figma HTTP/API clients, commands, output rendering, and persistence. `Node` validates and normalizes the mapped tree. `StringValue`, `NumberValue`, `OptionalNumber`, `NumberSlice`, and `MapValue` provide generic `any`-tree coercion shared by extract capabilities.

# Connections

- [Figma boundary](../figma/AGENTS.md): maps API payloads into document nodes before capability extraction.
- [Extraction](../extract/AGENTS.md): consumes this model and coercion helpers; document must not depend on extraction output types.

# Placement

Add stable tree invariants and generic traversal primitives here. Keep capability-specific output shaping in `internal/extract` and transport mapping in `internal/figma`.
