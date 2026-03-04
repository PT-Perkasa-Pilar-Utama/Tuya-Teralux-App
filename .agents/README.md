# AI Rules Hub

This directory is the single source of truth for AI assistant behavior in this repository.

Any AI tool integration file (Claude Code, Codex, Antigravity, GitHub Copilot) must resolve to the rules documented here.

## Start Here
Read these files in order:

1. `project-context.md`
2. `coding-rules.md`
3. `workflow-rules.md`

## Task-Specific Guides
- API contract requests: `api-contract-guide.md`

## Rule Priority
When rules conflict, use this priority:

1. Direct user request for the current task
2. Security and safety constraints
3. Repository coding and workflow rules in this folder
4. Tool-specific defaults

## Scope
These rules apply to:

- New code
- Refactoring
- Tests
- Documentation
- AI-generated commit or review suggestions
