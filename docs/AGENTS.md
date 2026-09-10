# Purpose

`docs/` contains user-facing command contracts and workflow explanations for the Figma CLI. Documentation should describe the stable interface exposed by the source and live Cobra help.

# Boundaries

Keep architecture and agent navigation in repository and module `AGENTS.md` files. Keep implementation details in source. When prose and command behavior diverge, source, tests, and live help are authoritative.

# Connections

- [Commands](cmd/AGENTS.md): defines the command surface documented here.
- [Internal capabilities](internal/AGENTS.md): defines output and behavior that documentation must represent without duplicating implementation.

# Placement

Add a document here when it explains a user-visible workflow or contract spanning commands. Put package ownership guidance in the nearest module guide instead.
