# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Code Quality Hooks

This project uses [lefthook](https://github.com/evilmartians/lefthook) for pre‑commit code quality checks. Hooks automatically run on `git commit` and check formatting, linting, tests, and more.

**Install hooks**:
```bash
make install-hooks
```

**Manual run**:
```bash
lefthook run pre-commit
```

**Bypass hooks** (emergencies only):
```bash
git commit --no-verify
```

See [PRE_COMMIT_HOOKS.md](docs/PRE_COMMIT_HOOKS.md) for full details.

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Ensure pre‑commit hooks pass (formatting, linting, tests). Run `lefthook run pre-commit` if needed.
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds

---

## Research Collection for Agent Outputs\n\nThis project uses **qmd** (Librarian CLI) for indexing agent reports and research findings.\n\n**Where to write outputs**:\n- **Reports**: Write task completion reports to `.tmp/reports/` with `.md` extension\n- **Research**: Write research notes and findings to `.tmp/researches/` with `.md` extension\n\n**Important**: Always write reports to `.tmp/reports/<task>-report.md` as specified in agent instructions.\n\nFor detailed qmd usage, see [Research Collection Documentation](docs/research-collection.md).\n\n---\n\n## Best Practices for Ordering Agents

See [Best Practices for Ordering Agents](docs/best-practices-agent-ordering.md) for detailed guidelines on agent sequencing.

---

## Temporary File Handling

When creating temporary files, always use the local `.tmp/` directory in the project root instead of system `/tmp/`. This ensures:

- Temporary files stay within project scope
- Better security and predictability  
- Consistency across all agents and sessions

Example: `.tmp/reports/task-report.md`

---

## Using Cobra CLI for Command Output

See [Using Cobra CLI for Command Output](docs/cobra-cli-output-methods.md) for proper output methods in Cobra-based applications.

<!-- bv-agent-instructions-v1 -->

---

## Beads Workflow Integration

This project uses [beads_viewer](https://github.com/Dicklesworthstone/beads_viewer) for issue tracking. Issues are stored in `.beads/` and tracked in git.

### Essential Commands

```bash
# View issues (launches TUI - avoid in automated sessions)
bv

# CLI commands for agents (use these instead)
bd ready              # Show issues ready to work (no blockers)
bd list --status=open # All open issues
bd show <id>          # Full issue details with dependencies
bd create --title="..." --type=task --priority=2
bd update <id> --status=in_progress
bd close <id> --reason="Completed"
bd close <id1> <id2>  # Close multiple issues at once
bd sync               # Commit and push changes
```

### Workflow Pattern

1. **Start**: Run `bd ready` to find actionable work
2. **Claim**: Use `bd update <id> --status=in_progress`
3. **Work**: Implement the task
4. **Complete**: Use `bd close <id>`
5. **Sync**: Always run `bd sync` at session end

### Key Concepts

- **Dependencies**: Issues can block other issues. `bd ready` shows only unblocked work.
- **Priority**: P0=critical, P1=high, P2=medium, P3=low, P4=backlog (use numbers, not words)
- **Types**: task, bug, feature, epic, question, docs
- **Blocking**: `bd dep add <issue> <depends-on>` to add dependencies

### Session Protocol

**Before ending any session, run this checklist:**

```bash
git status              # Check what changed
git add <files>         # Stage code changes
bd sync                 # Commit beads changes
git commit -m "..."     # Commit code
bd sync                 # Commit any new beads changes
git push                # Push to remote
```

### Best Practices

- Check `bd ready` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `bd create` when you discover tasks
- Use descriptive titles and set appropriate priority/type
- Always `bd sync` before ending session
<!-- end-bv-agent-instructions -->
