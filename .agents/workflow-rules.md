# Workflow Rules

## Before Editing

- Read relevant files first and confirm existing patterns.
- Identify affected module: `backend`, `sensio_app`, or `sensio_notification`.
- Prefer existing automation (`Makefile`, `scripts/`) over introducing new workflows.

## Mandatory Sequence

All tasks should follow:
1. Research and plan
2. Plan approval
3. Incremental implementation
4. Validation (`lint`, `build`, and `test` when applicable)
5. Final report with commands and pass/fail status

## Validation Gates by Module

- `backend/`
  - Local dev minimum: `make lint` + `make build`
  - Strict gate (required for remote/CI/release-like flow): `make lint-strict` + `make vet` + `make build`
  - Functional checks when behavior changes: `make test`
- `sensio_app/`
  - `make lint`
  - `make build`
- `sensio_notification/`
  - `make lint`
  - `make build`

## ThinkPad Lint Guard Policy (Backend)

- `backend` target `make lint` may skip `golangci-lint` when local machine is detected as ThinkPad via DMI fields (`sys_vendor`, `product_name`, `product_version`).
- This skip is local-dev convenience only for ThinkPad laptops.
- Rationale: some ThinkPad dev machines have limited RAM and `golangci-lint` can get force-closed under memory pressure.
- Remote scripts and strict pipelines must use `make lint-strict` so lint is always enforced.
- Do not downgrade strict gates by replacing `lint-strict` with `lint` in remote/deploy scripts.

## Quality and Safety Rules

- Never bypass checks with hook-skip flags.
- Do not ignore newly introduced validation failures.
- Keep backward compatibility unless explicitly requested.
- Document workflow or command changes in `.agents` when they affect contributors.
