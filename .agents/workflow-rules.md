# Workflow Rules

## Before Editing

- Read relevant files (backend or android) first.
- Confirm assumptions from existing code patterns.
- Reuse established project conventions.
- Be aware of the sub-project context: `backend/`, `sensio_app/`, or `sensio_notification/`.
- Check the root `Makefile` for available automation.

## During Implementation

- Prefer incremental changes over broad rewrites.
- Start with the exact problem the user requested.
- Maintain consistency between backend Go logic and Android Kotlin implementation.
- Maintain consistency between backend Go logic and Android Kotlin implementation.
- Use `make dev` for a hot-reloading development environment.
- **Requirement**: Always run linter and build after edits and fix any warnings/errors immediately.

## Validation

- For project-wide checks: `make test` and `make vet` from the root directory.
- For backend specific: `cd backend && make test`.
- For Android specific: `cd sensio_app && ./gradlew build`.
- Always check that your changes don't break the MQTT communication layer.

## Output Quality

- Summarize what changed and why.
- Reference touched files clearly.
- Highlight risks, tradeoffs, and follow-up steps.
