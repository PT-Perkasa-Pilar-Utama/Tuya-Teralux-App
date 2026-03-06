# Coding Rules

## General

- Keep changes minimal, focused, and reversible.
- Do not alter behavior unless requested.
- Preserve backward compatibility unless a breaking change is explicitly required.
- Never hardcode secrets, keys, tokens, passwords, or private endpoints.
- Use environment/config files for sensitive values.
- Add or update tests when behavior changes.
- Fix lint/build failures caused or surfaced by your changes before finishing.

## Go Rules (`backend/`)

- Follow idiomatic Go formatting (`gofmt`) and existing project structure.
- Prioritize clear package boundaries under `backend/domain/...`.
- Return contextual errors; avoid silent failure paths.
- Avoid global mutable state unless already established and justified.
- Keep controller/usecase/service/repository responsibilities separated as currently organized.

## Kotlin / Gradle Rules (`sensio_app/`, `sensio_notification/`)

- Follow existing Gradle and module conventions; avoid ad-hoc task wiring.
- Keep UI/presentation logic separated from service/integration logic where project structure already enforces it.
- Use `ktlint` conventions; do not add style bypasses unless explicitly approved.
- Avoid embedding secrets in `build.gradle*`, source files, or checked-in property files.

## Automation and Scripts

- Prefer existing Makefile targets and `scripts/` helpers over one-off custom commands.
- For remote workflows, keep `backend/scripts/*` aligned with shared helper conventions in `scripts/remote/common_arch.sh`.
- Any workflow policy change in scripts must be mirrored in `.agents` documentation.

## Documentation

- Update docs when commands, validation flow, or developer workflow changes.
- Keep instructions concrete and repository-specific.
