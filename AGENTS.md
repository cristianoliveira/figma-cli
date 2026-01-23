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

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
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

## Index of Indexed Documentation (ZK)

Zk indexes documentation for search by tags and content.

**Install & Basic Commands:**
- Install: `brew install zk` or download from [zk-org/zk](https://github.com/zk-org/zk)
- List notes: `zk list`, filter by tags: `zk list --tags "figma-api"`
- Search content: `zk list --match "authentication"`
- Interactive edit: `zk edit --interactive`
- Create note: `zk new --title "Note Title"`

**Tag Categories:** `resources`, `figma-api`, `api-docs`, `github`, `references`

**Example Searches:**
```bash
zk list --tags "resources"
zk list --tags "figma-api" --tags "api-docs"
zk list --match "authentication"
```

**Adding Documentation:**
1. Place markdown files in `docs/` with frontmatter (`title`, `tags`)
2. Run `zk index` (optional, auto‑indexed)

**IMPORTANT:** Use zk to index new documents.

---

## Best Practices for Ordering Agents

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