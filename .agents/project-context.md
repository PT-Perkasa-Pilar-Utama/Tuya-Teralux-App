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

## Component Boundaries

- Keep backend API/domain logic in `backend/domain/...`.
- Keep mobile app logic confined to `sensio_app/...`.
- Keep notification app logic confined to `sensio_notification/...`.
- Reuse shared remote helper behavior from `scripts/remote/common_arch.sh`; avoid duplicating transport/sync logic in module scripts.

## Common Validation Commands

Run commands from each module root:

- `backend/`
  - `make lint` (local dev lint; ThinkPad guard may skip)
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

## Engineering Goal

Deliver minimal, safe, and reversible changes while preserving module boundaries, validation discipline, and secure configuration handling (`.env`, keystore, credentials).
