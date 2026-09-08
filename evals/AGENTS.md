# Purpose

`evals/` contains offline evaluation inputs and coordinator-owned evidence for agent workflows, especially the pixel-perfect implementation task.

# Boundaries

Evaluation fixtures are not production code and must not be copied into skills as answers. They exercise real applications and CLIs under controlled inputs; claims require independently generated evidence.

# Connections

- [Pixel-perfect skill](skills/pixel-perfect/AGENTS.md): provides the workflow under evaluation.
- [Skill evaluation helpers](skills/pixel-perfect/evals/AGENTS.md): validates skill packaging and protocol inputs.
- [Upload-panel controls](evals/pixel-perfect/upload-panel/AGENTS.md): owns the accepted snapshot and defect controls for one evaluation case.

# Placement

Add a fixture here when it tests a stable agent workflow without becoming runtime code. Keep reusable evaluator mechanics in the nearest evaluation helper package.
