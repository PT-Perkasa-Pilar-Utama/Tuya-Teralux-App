# Coding Rules

## General

- Keep code changes minimal, focused, and reversible.
- Do not change behavior unless explicitly requested.
- Prioritize security: never hardcode passwords, keys, tokens, or other credentials in the source code.
- Always use environment variables for sensitive data.
- Add or update tests for behavior changes.
- **IMPORTANT**: After editing any file, always run the linter and build to ensure there are no warnings or errors.

## Go Conventions (`backend/`)

- Follow standard Go style and `gofmt`.
- Return errors with useful context using `fmt.Errorf` or similar.
- Keep `services/` modular.
- Avoid global mutable state.
- Ensure Swagger documentation is updated if API endpoints change.

## Kotlin / Android Conventions (`sensio_app/`)

- Follow modern Kotlin best practices (ViewBinding, ViewModel, Coroutines).
- Use descriptive names for UI elements and resources.
- Ensure proper lifecycle management in Activities and Fragments.
- Maintain consistency with existing material design patterns.

## Architecture & Communication

- Maintain strict boundaries between the backend and Android client.
- Ensure MQTT payloads remain consistent across both ends.
- Backend services (e.g., `EMQX-Auth-Service`) should be treated as internal infrastructure.

## Data and Migrations

- Backend schema changes must include migration files handled by `make migrate-up` (or equivalent in `backend/Makefile`).
- Database interactions should be decoupled from service logic.

## Documentation

- Update `README.md` or specialized guides when behavior changes.
- Ensure `Makefile` help is kept up to date.
