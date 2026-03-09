# Project Context

## Repository

- Name: Tuya-Teralux-App
- Type: Monorepo
- Core technologies:
  - Go backend service
  - Android (Kotlin/Gradle) app
  - Kotlin Multiplatform notification app
  - Docker-based local/runtime services

## High-Level Structure

- `backend/`: Main Go API/backend service.
  - Domain code: `backend/domain/...`
  - Local automation: `backend/Makefile`
  - Remote automation scripts: `backend/scripts/*.sh`
- `sensio_app/`: Android application module (Gradle + ktlint).
- `sensio_notification/`: Notification application (KMP/Compose + ktlint).
- `scripts/remote/`: Shared remote sync/deploy helpers (`common_arch.sh`).
- **Host Mapping**: `arch` == Nitro 5 host.

## Component Boundaries

- Keep backend API/domain logic in `backend/domain/...`.
- Keep mobile app logic confined to `sensio_app/...`.
- Keep notification app logic confined to `sensio_notification/...`.
- Reuse shared remote helper behavior from `scripts/remote/common_arch.sh`; avoid duplicating transport/sync logic in module scripts.

## Common Validation Commands

Run commands from each module root:

- `backend/`
  - `make lint` (always run golangci-lint)
  - `make lint-strict` (always run golangci-lint)
  - `make vet`
  - `make build`
  - `make test`
- `sensio_app/`
  - `make lint`
  - `make build`
- `sensio_notification/`
  - `make lint`
  - `make build`
- **ThinkPad Note**: If a ThinkPad is detected, local validation is disallowed.
- **Source of Truth**: The ThinkPad working tree is the authoritative source; all remote validation on `arch` must use a snapshot synced from ThinkPad using `preflight_check` and `sync_source_delta`.
  - **Delta-Only Policy**: Sync propagates tracked changes and deletions only.
- **Host Mapping**: `arch` == Nitro 5 remote host.

## Planning Flow

- Always read `.agents` files in this sequence:
  1. `project-context.md`
  2. `coding-rules.md`
  3. `workflow-rules.md`
- AI must explicitly confirm this read sequence in every implementation plan.

Deliver minimal, safe, and reversible changes while preserving module boundaries, validation discipline, and secure configuration handling (`.env`, keystore, credentials).
