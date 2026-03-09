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
3. Incremental implementation.
4. Validation (`lint`, `build`, and `test` when applicable).
5. Final report with commands and pass/fail status.

## Validation Gates by Module

- `backend/`
  - Local dev minimum **(non-ThinkPad only)**: `make lint` + `make build`
  - Strict gate (required for remote/CI/release-like flow): `make lint-strict` + `make vet` + `make build`
  - Functional checks when behavior changes: `make test`
- `sensio_app/`
  - Local dev minimum **(non-ThinkPad only)**: `make lint` + `make build`
- `sensio_notification/`
  - Local dev minimum **(non-ThinkPad only)**: `make lint` + `make build`

## ThinkPad-to-Nitro5 Enforcement Policy (All Modules)

- This policy applies to `backend/`, `sensio_app/`, and `sensio_notification/`.
- **Detection**: The local machine is a ThinkPad if `/sys/devices/virtual/dmi/id/` fields (`sys_vendor`, `product_name`, `product_version`) match `thinkpad` (case-insensitive).
- **Enforcement**: On a ThinkPad, local `lint` and `build` execution is **disallowed** for all modules.
- **Mandatory Pre-step**: AI **must** sync ThinkPad -> `arch` (Nitro 5) before remote validation.
  - **Delta-Only Sync**: Sync only changed tracked files + tracked deletions using `preflight_check` and `sync_source_delta`.
  - **No Full Raw Copy**: Avoid full repository copy by default; untracked files are skipped.
- **Remote Execution**: Only after a successful sync may `lint`, `build`, and `test` run on the remote host (`arch`).
- **Non-compliant Flow**: Running `lint` or `build` on `arch` without a fresh sync from ThinkPad is strictly non-compliant.
- **Rationale**: Ensures the remote environment exactly reflects the current ThinkPad working tree before validation while optimizing for speed via delta sync.
- **Exceptions**: Non-ThinkPad machines may follow module-local validation commands.

## Quality and Safety Rules

- Never bypass checks with hook-skip flags.
- Do not ignore newly introduced validation failures.
- Keep backward compatibility unless explicitly requested.
- Document workflow or command changes in `.agents` when they affect contributors.
