# Research notes

Use this directory for reviewed findings that future contributors need. Put
current usage instructions in the [project README](../README.md) or
[command contracts](../docs/command-contracts.md), not in research notes.

## Choose a location

| Material | Location |
| --- | --- |
| Durable findings and decision evidence | `research/` |
| Task reports | `${AGENT_WORKSPACE:-$PWD/.tmp}/reports/<dd-mm-yy>/` |
| Temporary experiments and raw output | A local, untracked workspace such as `.tmp/researches/` |
| Proposed implementation work | [`plans/todo/`](../plans/todo/README.md) |

Resolve report paths from the repository root. Keep temporary artifacts untracked;
do not force-add ignored files. Reports and raw output do not belong in this
folder unless they have been reviewed and are useful as permanent evidence.

## Write a useful note

Include:

1. The question and date.
2. Relevant versions, inputs, and sources.
3. What you observed, separate from assumptions or interpretations.
4. Commands or steps needed to reproduce the finding.
5. Limits, unresolved questions, and any resulting decision.

Use a descriptive filename, short sections, and relative links to repository
files. Link external claims to their sources. Remove credentials, private design
data, and machine-specific paths before committing.

## Search

Search tracked notes from the repository root:

```bash
rg -n 'search term' research docs
```

There is no repository-managed automatic research index. If you use a local
search tool such as qmd, configure its collections yourself. Do not assume another
contributor has the same index or local report directory.
