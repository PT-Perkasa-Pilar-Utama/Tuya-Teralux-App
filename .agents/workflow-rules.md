# Workflow Rules

## Before Editing

- Read relevant files first and confirm existing patterns.
- Identify affected module: `backend`, `sensio_app`, or `sensio_notification`.
- Prefer existing automation (`Makefile`, `scripts/`) over introducing new workflows.

## Mandatory Sequence

All tasks should follow:

1. **Plan Preparation**: Read `.agents` in canonical order (`project-context.md`, `coding-rules.md`, `workflow-rules.md`).
2. **AI Plan Approval**:
   - **Plan language = English only** (regardless of conversation language).
   - **Plan must explicitly state** that the `.agents` read step was done.
   - **Plan must include explanation/rationale** for each major step.
   - **Plan must explicitly include** a post-change validation step: run `lint` and `build` in affected module(s), with rationale that these are mandatory quality gates before completion.
3. Incremental implementation.
4. Validation:
   - Always run `lint` and `build` in affected module(s) after changes.
   - Run `test` when behavior changes.
5. Final report with commands and pass/fail status.

## Validation Gates by Module

- `backend/`
  - Mandatory local baseline after changes: `make lint` + `make build`
  - Strict gate (required for CI/release-like flow): `make lint-strict` + `make test` + `make build`
  - Functional checks when behavior changes: `make test`
- `sensio_app/`
  - Mandatory local baseline after changes: `make lint` + `make build`
- `sensio_notification/`
  - Mandatory local baseline after changes: `make lint` + `make build`

## Quality and Safety Rules

- Never bypass checks with hook-skip flags.
- Do not ignore newly introduced validation failures.
- Keep backward compatibility unless explicitly requested.
- Document workflow or command changes in `.agents` when they affect contributors.
