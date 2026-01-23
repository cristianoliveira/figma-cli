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

Effective multi‑agent workflows require thoughtful sequencing of specialized agents. Follow these guidelines to maximize productivity and code quality.

### Choosing the Right Agent
- **Task‑Specialist Agents**: Use agents with specific skills (e.g., `db‑explorer`, `logcli‑logs`) for domain‑specific tasks
- **General‑Purpose Agents**: Use default agents for broad implementation tasks
- **Review Agents**: Consider using `gh‑address‑comments` for handling GitHub PR feedback
- **Validation Agents**: Use `land‑the‑plane` for pre‑merge CI validation

### Common Patterns for Agent Ordering
1. **Explore → Implement → Validate**
   - Start with exploratory agents (`db‑explorer`, `look‑at‑the‑logs`) to understand context
   - Follow with implementation agents to make changes
   - Finish with validation agents (`land‑the‑plane`) to ensure quality

2. **Parallel Specialization**
   - Run multiple specialist agents concurrently when tasks are independent
   - Example: `db‑explorer` and `logcli‑logs` can run simultaneously to gather different data

3. **Feedback Loop**
   - Use `gh‑address‑comments` to process review feedback
   - Follow with implementation agents to apply changes
   - Re‑run validation agents after updates

### Example Sequences
- **New Feature**: `db‑explorer` → General agent → `land‑the‑plane` → `gh‑address‑comments`
- **Debugging**: `look‑at‑the‑logs` → `db‑explorer` → General agent → `land‑the‑plane`
- **Documentation**: General agent → `land‑the‑plane` → `gh‑address‑comments`

### Key Principles
- **Minimal Changes**: Make smallest possible change that moves task forward
- **Follow Conventions**: Adhere to existing code style and project patterns
- **Clear Hand‑offs**: Leave clear notes/issues for next agent
- **Atomic Tasks**: Break work into small, well‑defined deliverables
- **Document Assumptions**: Document assumptions and decisions in reports
- **Verify Continuously**: Run validation steps after each major change
- **Ask Early**: If ambiguous, ask for clarification before proceeding

By following these practices, you can create efficient, reliable multi‑agent workflows that produce high‑quality results.

---

## Temporary File Handling

When creating temporary files, always use the local `.tmp/` directory in the project root instead of system `/tmp/`. This ensures:

- Temporary files stay within project scope
- Better security and predictability  
- Consistency across all agents and sessions

Example: `.tmp/reports/task-report.md`

---

## Using Cobra CLI for Command Output

**Problem**: Agents are using pure `fmt.Print*` statements for CLI output, which is incorrect for Cobra-based applications.

**Solution**: Use Cobra's built-in output methods that respect the command's configured output writers and enable proper testing and output redirection.

### Cobra Output Methods
Cobra provides six output methods on the `Command` struct:

- `Print`, `Println`, `Printf` – for standard output
- `PrintErr`, `PrintErrln`, `PrintErrf` – for error output

### Correct Usage Example
**Instead of** `fmt.Println("Starting process...")` **use** `cmd.Println("Starting process...")`

**Correct pattern:**
```go
func Run(cmd *Command, args []string) {
    cmd.Println("Starting process...")
    cmd.Printf("Processing %d items\n", count)
    if err != nil {
        cmd.PrintErrf("Error: %v\n", err)
    }
}
```

### Best Practices
1. **Never use `fmt.Print*`** in command `Run`/`RunE` functions
2. **Use `cmd.Print*` for normal output** (information, results, progress)
3. **Use `cmd.PrintErr*` for errors, warnings, and diagnostic messages**
4. **Be consistent** – Use the same pattern across all commands

### Error Handling Pattern
When returning errors from `RunE`, use Cobra's error output for user-facing messages:
```go
RunE: func(cmd *Command, args []string) error {
    if err := validate(args); err != nil {
        cmd.PrintErrln("Validation failed:", err)
        return err
    }
    cmd.Println("Operation successful")
    return nil
}
```

### Notes
- The project uses Go's `flag` package but includes `cobra-cli` in dev environment
- When adding new commands, prefer Cobra patterns
- Follow existing patterns in codebase where Cobra is gradually adopted
