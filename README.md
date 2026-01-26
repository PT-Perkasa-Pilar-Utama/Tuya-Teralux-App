# Teralux Smart Home System

Welcome to the **Teralux** project source code. this repository hosts the codebase for the Teralux smart home ecosystem, designed to provide seamless control and management of smart devices.

## 📂 Repository Structure

The project is divided into two main components:

*   **`android/`**
    *   Contains the native Android mobile application source code.
    *   Built using **Kotlin** and **Jetpack Compose**.
    *   Acts as the client interface for users to interact with smart home environment.

*   **`backend/`**
    *   Contains the server-side logic and API implementation.
    *   Built using **Go (Golang)**.
    *   Manages device connectivity, state synchronization, and integration with third-party platforms like Tuya.

### Installation & User Guide
For detailed instructions on how to configure, build, and run both the Backend and Android applications (using `Makefile`), please consult the dedicated documentation:

👉 **[Read the Installation Guide](backend/docs/manual/installation_guide.md)**

## 🛠️ Key Technologies

*   **Mobile**: Android, Kotlin, Jetpack Compose
*   **Backend**: Go, Docker
*   **Database**: SQLite
*   **IoT Integration**: Tuya Smart Cloud

## 🗂️ Recent change

- ✅ **`domain/rag`** and **`domain/speech`** have been migrated from `stt-service` into `backend/domain`.
- ⚠️ `stt-service` is no longer required for these features; the code is now maintained under `backend`.

