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

See [PRE_COMMIT_HOOKS.md](docs/PRE_COMMIT_HOOKS.md) for full details. For comprehensive code quality standards, see [Code Quality Best Practices](#code-quality-best-practices).

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

## Code Quality Best Practices

Based on comprehensive code reviews and architecture investigations, these are the established best practices for this project:

### Critical Practices (DO NOT Violate)

#### 1. Code Duplication is P0
- **Rule**: Any code duplication > 5 lines must be eliminated
- **Tools**: Use jscpd for detection, enforced in CI
- **Solutions**: Generic wrappers, code generation, reflection, or DRY principles
- **Example**: Instead of copying retry logic 30 times, create a generic retry wrapper
- **Priority**: P0 - This is critical for maintainability

#### 2. Complete Implementation
- **Rule**: Never commit "not implemented" stubs
- **Exception**: Only in prototype phase, documented in issue
- **Requirement**: All methods must have proper implementation or be removed from interface
- **Priority**: P1 - Blocking functionality

#### 3. Error Handling
- **Rule**: Always check errors from flag retrieval and all API calls
- **Pattern**: If a function returns error, MUST handle it
- **No silent failures**: Panic or explicit error handling, never ignore errors
- **Priority**: P1 - Could cause crashes

#### 4. Interface Segregation
- **Rule**: Interfaces should have < 10 methods
- **Principle**: Clients should only depend on methods they use
- **Split large interfaces**: Group by domain (FileClient, CommentClient, etc.)
- **Priority**: P1 - SOLID violation

### Important Practices

#### 5. Token Type-Aware Configuration
- **Rule**: Validation must differ for PAT vs OAuth tokens
- **Required fields**:
  - PAT: token only
  - OAuth: client_id, client_secret, redirect_uri
- **Action**: Add token type field, validate accordingly

#### 6. Configuration Simplicity
- **Rule**: Config merging should be < 100 lines, readable
- **Avoid**: Nested if-else chains for config merging
- **Prefer**: Pluggable ConfigSource interface or reflection

#### 7. Test Coverage
- **Target**: 70%+ coverage on critical packages
- **Must test**:
  - API client (with httptest.Server)
  - Configuration loading
  - CLI commands
  - Error paths
- **Avoid**: Testing private functions, over-mocking

#### 8. TODO Management
- **Rule**: No TODO comments in production code
- **Process**:
  - Implement TODO (if < 2 hours)
  - File issue in beads (if > 2 hours)
  - Remove if obsolete
- **Enforcement**: linter should flag TODOs

### Nice to Have

#### 9. Code Cleanliness
- **Rule**: No unused variables or imports
- **Tools**: goimports, golangci-lint unused checker
- **Enforcement**: Pre-commit hooks

#### 10. Long Functions
- **Rule**: Functions should be < 20 lines
- **Split**: Into smaller, single-purpose functions
- **Name**: Descriptive names that explain purpose

### Architecture Best Practices

#### Clean Architecture
- **Layers**: Entities → Use Cases → Interface Adapters → Frameworks
- **Dependency Rule**: Dependencies point inward
- **No cross-layer**: Frameworks can't depend on use cases

#### SOLID Principles
- **SRP**: One reason to change per component
- **OCP**: Open for extension, closed for modification
- **LSP**: Subtypes must be substitutable
- **ISP**: Interfaces should be small and focused
- **DIP**: Depend on abstractions, not concretions

### Tooling Requirements

#### Pre-Commit Hooks (Mandatory)
- `go fmt` - Format on save
- `goimports` - Fix imports
- `golangci-lint` - Lint and fix
- `go test -short` - Quick test run

#### CI Enforcements
- Code duplication check (jscpd)
- Test coverage minimum (60-70%)
- Linting must pass
- No TODOs in production code

### Anti-Patterns to Avoid

#### ❌ Code Duplication
- Copy-pasting retry logic
- Similar functions with small differences
- Repeated error handling patterns
- **Fix**: Extract to reusable components

#### ❌ Over-Mocking
- Mocking your own code
- Testing implementation details
- Tests for private functions
- **Fix**: Use real implementations where possible

#### ❌ "Not Implemented" Stubs
- Returning "not implemented" errors
- Empty function bodies
- **Fix**: Implement properly or remove from interface

#### ❌ Ignoring Errors
- `_ = someFunc()` (ignoring error)
- No error checks from flag functions
- **Fix**: Always handle errors explicitly

### Code Review Checklist

Before submitting code or merging PR, verify:

- [ ] No code duplication (> 5 lines)
- [ ] All methods fully implemented (no stubs)
- [ ] All errors handled
- [ ] Interfaces have < 10 methods
- [ ] Functions < 20 lines
- [ ] Test coverage > 60% for changed files
- [ ] No TODO comments
- [ ] No unused variables/imports
- [ ] Pre-commit hooks pass
- [ ] CI checks pass

## Recent Investigation Summaries

### Code Quality Investigation (Overall: 6/10)
- **Strengths**: Clean architecture, comprehensive logging, good error types
- **Critical Issues**: Code duplication (400+ lines), incomplete implementations, missing error handling
- **Priority Fixes**: P0 (duplication), P1 (unimplemented methods, error handling, OAuth refresh)
- **Report**: `.tmp/reports/code-quality-investigation-1.md`

### Architecture & Design Investigation (Overall: 7/10)
- **Strengths**: Clear layering, good configuration management, structured logging
- **Critical Issues**: Incomplete HTTP client, interface bloat, config validation gaps
- **Priority Fixes**: P1 (implement methods, split interface), P2 (config validation, retry duplication)
- **Report**: `.tmp/reports/architecture-quality-investigation-1.md`

### Key Takeaways
- Foundation is solid, but focus on completeness and eliminating duplication
- Implement missing API methods to unblock functionality
- Split large interfaces to improve testability
- Add comprehensive test coverage
