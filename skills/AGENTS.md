# Purpose

`skills/` contains agent-facing workflows that teach reliable use of this repository's CLIs for design exploration and visual implementation.

# Boundaries

Skills are guidance and evaluation inputs, not production runtime code. They may point agents to commands and artifacts, but must not become a second implementation of CLI behavior.

# Connections

- [Commands](cmd/AGENTS.md): supplies the executable behavior skills describe.
- [Documentation](docs/AGENTS.md): supplies user-facing contract context.
- [Pixel-perfect skill](skills/pixel-perfect/AGENTS.md): owns the screenshot implementation workflow.
- [Evaluation fixtures](tests/evals/AGENTS.md): evaluates skill behavior outside the skill content.

# Placement

Put reusable agent procedures in the skill that owns the workflow. Put production behavior in Go packages and holdout answers or grader controls in `tests/evals/`.
