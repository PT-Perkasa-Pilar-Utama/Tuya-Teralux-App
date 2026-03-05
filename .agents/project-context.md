# Project Context - Sensio App

## Repository

- Name: Tuya-Teralux-App (Sensio App)
- Type: Monorepo IoT Platform
- Core Technologies: Go (Backend), Kotlin (Android), Docker, MQTT, Tuya SDK
- Purpose: A comprehensive IoT management platform (Sensio) for controlling Tuya-compatible and other IoT devices.

## High-Level Structure

- `backend/`: Main backend service written in Go.
  - `backend/main.go`: Entry point.
  - `backend/services/`: Core logic and service integrations (e.g., EMQX, LLM, Whisper).
  - `backend/domain/`: Domain models and business logic.
  - `backend/Makefile`: Backend-specific automation.
- `sensio_app/`: Main Android client application written in Kotlin/Java.
  - `sensio_app/app/`: Android app source code.
  - `sensio_app/Makefile`: Android build automation.
- `sensio_notification/`: Notification service/component.
- `Makefile`: Root Makefile for project-wide automation (setup, dev, tests).

## Development Commands

- Root: `make setup`, `make dev`, `make test`, `make vet`.
- Backend: `cd backend && make dev`, `go mod tidy`.
- Android: `cd sensio_app && ./gradlew build`.

## Engineering Goal

Preserve the separation between the Go backend and the Kotlin Android app. Maintain security hygiene, especially around Tuya credentials and MQTT authentication. Favor the root `Makefile` for multi-component tasks.
