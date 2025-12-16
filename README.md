# Teralux Smart Home System

Welcome to the **Teralux** project source code. this repository hosts the codebase for the Teralux smart home ecosystem, designed to provide seamless control and management of smart devices.

## 📂 Repository Structure

The project is divided into two main components:

*   **`android/`**
    *   Contains the native Android mobile application source code.
    *   Built using **Kotlin** and **Jetpack Compose**.
    *   Acts as the client interface for users to interact with their smart home environment.

*   **`backend/`**
    *   Contains the server-side logic and API implementation.
    *   Built using **Go (Golang)**.
    *   Manages device connectivity, state synchronization, and integration with third-party platforms like Tuya.

## 🚀 Getting Started

### Backend Setup
For detailed instructions on how to configure, build, and run the backend server (including Docker and Make commands), please consult the dedicated documentation:

👉 **[Read the Backend README](backend/README.md)**

### Android Setup
To work on the mobile application:
1.  Open the `android` directory in **Android Studio**.
2.  Sync the Gradle project.
3.  Ensure your `local.properties` is configured.
    *   **Important**: You must define `API_KEY` in this file to match the backend's configuration.
    *   Example: `API_KEY="your_backend_api_key_here"`
4.  Run the application on an emulator or physical device.

## 🛠️ Key Technologies

*   **Mobile**: Android, Kotlin, Jetpack Compose
*   **Backend**: Go, Docker
*   **Database**: SQLite
*   **IoT Integration**: Tuya Smart Cloud
