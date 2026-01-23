---
title: "Best Practices for Ordering Agents"
tags: ["agents", "best-practices", "ordering"]
---
# Best Practices for Ordering Agents

Effective multi‑agent workflows require thoughtful sequencing of specialized agents. Follow these guidelines to maximize productivity and code quality.

## Choosing the Right Agent
- **Task‑Specialist Agents**: Use agents with specific skills (e.g., `db‑explorer`, `logcli‑logs`) for domain‑specific tasks
- **General‑Purpose Agents**: Use default agents for broad implementation tasks
- **Review Agents**: Consider using `gh‑address‑comments` for handling GitHub PR feedback
- **Validation Agents**: Use `land‑the‑plane` for pre‑merge CI validation

## Common Patterns for Agent Ordering
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

## Example Sequences
- **New Feature**: `db‑explorer` → General agent → `land‑the‑plane` → `gh‑address‑comments`
- **Debugging**: `look‑at‑the‑logs` → `db‑explorer` → General agent → `land‑the‑plane`
- **Documentation**: General agent → `land‑the‑plane` → `gh‑address‑comments`

## Key Principles
- **Minimal Changes**: Make smallest possible change that moves task forward
- **Follow Conventions**: Adhere to existing code style and project patterns
- **Clear Hand‑offs**: Leave clear notes/issues for next agent
- **Atomic Tasks**: Break work into small, well‑defined deliverables
- **Document Assumptions**: Document assumptions and decisions in reports
- **Verify Continuously**: Run validation steps after each major change
- **Ask Early**: If ambiguous, ask for clarification before proceeding

By following these practices, you can create efficient, reliable multi‑agent workflows that produce high‑quality results.