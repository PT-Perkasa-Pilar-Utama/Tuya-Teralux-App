# Contributing via AI Agents

This project supports AI-assisted contribution with a single source of rules under `.agents/`.

## Canonical Rules

Always read and apply all instructions under the `.agents/` directory. Bridge files must not define parallel or conflicting rules.

## Agent Entry Files

- Codex: `AGENTS.md`
- Claude Code: `CLAUDE.md`
- Antigravity: `ANTIGRAVITY.md`
- GitHub Copilot: `.github/copilot-instructions.md`
- JetBrains AI Assistant: `.aiassistant/rules/01-agents-rules.md`
- Junie: `.junie/guidelines.md`

All files above should point back to `.agents/` and must not define conflicting rules.

## Contribution Workflow (AI-Assisted)

1. Sync with current repository context and read the canonical rules.
2. Keep changes focused and minimal.
3. Follow existing architecture boundaries (`domains/`, `infrastructure/`, `internal/`).
4. Update documentation when behavior or contracts change.
5. Run required validation and ensure it passes (`lint` and `build` are mandatory minimum gates; `test` when applicable).
6. In PR notes, mention:
   - which rules were applied,
   - what files changed,
   - what validation was run and passed (`lint`, `build`, and `test` when applicable).

## Updating AI Rules

If you need to change AI behavior:

1. Update files under `.agents/` first.
2. Ensure bridge files point only to the `.agents/` directory.
3. Do not create parallel rule sets in unrelated files.

## Mandatory Execution Policy

Every AI agent must:

1. Create a plan and get approval.
2. Pass `lint` and `build` gates before submitting.
3. Fix all discovered errors/warnings in the scope of changes.
4. Never bypass validation hooks.

## Quick Checklist

- Rules were read in the required order.
- Task-specific guide used when relevant.
- No conflicting instruction file was introduced.
- Validation status is clearly reported with `lint` and `build` pass results.
