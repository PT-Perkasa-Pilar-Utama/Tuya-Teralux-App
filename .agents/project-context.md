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
  - Automation: `backend/Makefile`, `backend/scripts/*.sh`
- `sensio_app/`: Android application module (Gradle + ktlint).
- `sensio_notification/`: Notification application (KMP/Compose + ktlint).

## Component Boundaries

- Keep backend API/domain logic in `backend/domain/...`.
- Keep mobile app logic confined to `sensio_app/...`.
- Keep notification app logic confined to `sensio_notification/...`.

## Common Validation Commands

Run commands from each module root:

- `backend/`
  - `make lint` (always run go vet and go build)
  - `make lint-strict` (always run go vet and go build)
  - `make vet`
  - `make build`
  - `make test`
- `sensio_app/`
  - `make lint`
  - `make build`
- `sensio_notification/`
  - `make lint`
  - `make build`
- **Mandatory Baseline**: After every code/config change, run local `lint` and `build` in affected module(s).
- **Behavior-Change Gate**: Run tests when behavior changes.

## Planning Flow

- Always read `.agents` files in this sequence:
  1. `project-context.md`
  2. `coding-rules.md`
  3. `workflow-rules.md`
- AI must explicitly confirm this read sequence in every implementation plan.
- Plans must explicitly include post-change `lint` + `build` and explain that they are mandatory quality gates.

Deliver minimal, safe, and reversible changes while preserving module boundaries, validation discipline, and secure configuration handling (`.env`, keystore, credentials).
