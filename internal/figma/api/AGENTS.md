# Purpose

`internal/figma/api` contains generated Go models and clients derived from the repository's Figma OpenAPI definition.

# Boundaries

Treat this package as an API schema boundary, not a place for CLI or domain policy. Consumers should normally enter through [Figma transport](internal/figma/AGENTS.md), which maps generated values before passing them to other modules.

# Connections

- [Figma transport](internal/figma/AGENTS.md): owns use of generated responses and isolates schema churn from callers.

# Placement

Change the OpenAPI source or generation configuration when the contract changes, then regenerate this package. Do not hand-edit generated Go output or add business workflows here.
