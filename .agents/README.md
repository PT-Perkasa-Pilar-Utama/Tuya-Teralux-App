# AI Rules Hub

`.agents/` is the canonical rule source for all AI assistants in this repository.

Bridge files must only point to `.agents/` and must not define conflicting behavior:
- `AGENTS.md`
- `CLAUDE.md`
- `ANTIGRAVITY.md`
- `.github/copilot-instructions.md`
- other assistant bridge files in this repo

## Read Order

Always load in this order:
1. `project-context.md`
2. `coding-rules.md`
3. `workflow-rules.md`

## Task-Specific Guides

- API contract work: `api-contract-guide.md`

## Rule Priority

If rules conflict, use this priority:
1. Direct user request for the active task
2. Security and safety constraints
3. Repository rules in `.agents/`
4. Tool defaults

## Scope

These rules apply to code, scripts, tests, docs, and AI-generated review suggestions across:
- `backend/`
- `sensio_app/`
- `sensio_notification/`
