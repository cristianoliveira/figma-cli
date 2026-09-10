# Purpose

`skills/` contains agent-facing workflows that teach reliable use of this repository's Figma CLI for design exploration and implementation.

# Boundaries

Skills are guidance and evaluation inputs, not production runtime code. They may point agents to commands and artifacts, but must not become a second implementation of CLI behavior.

# Connections

- [Commands](cmd/AGENTS.md): supplies the executable behavior skills describe.
- [Documentation](docs/AGENTS.md): supplies user-facing contract context.

# Placement

Put reusable agent procedures in the skill that owns the workflow. Put production behavior in Go packages; keep evaluation inputs outside the installable skill content.
