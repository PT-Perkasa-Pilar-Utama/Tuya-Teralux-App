# Meeting Pipeline Architecture

This document outlines the technical implementation used to harden the Sensio Meeting Pipeline. The architecture is designed to handle up to ~2GB audio streams natively without out of memory (OOM) errors or gateway timeouts, while propagating live updates safely via MQTT.

## Backend Implementation

### Accurate Overlapped Segmentation

The pipeline utilizes an `audio_segmenter` that employs `ffmpeg` with `-ss` and `-t` flags for precise slicing. This implements a robust windowing formula (`start = i * segmentSec - overlap`), ensuring stable boundaries for the deduplication engine.

### Transcription Orchestrator

The main processing engine handles concurrent segments using a configurable semaphore. It implements per-segment retries (up to 2 attempts) to handle ephemeral network failures, alongside a word-level deduplication merge strategy to handle overlap seams perfectly.

### Provider Hardening & Memory Streaming

- **OpenAI, Groq, Orion**: Services natively stream bytes using `io.Pipe`, coupled with proper error propagation (`CloseWithError`) to prevent dangling requests.
- **Gemini**: Due to provider limitations, Gemini employs inline base64 payloads governed by a strict 20MB file-size guard. Large files are safely processed via the segmented transcription path, avoiding memory limits constraints.

### MQTT Pipeline Trace Updates

A comprehensive task event specification (`TaskEventV1`) is broadcasted in real-time as tasks traverse the workflow stages. This guarantees the client is always synchronized with the backend processing state (`accepted`, `started`, `stage_update`, `completed`, `failed`).

## Android Upload Optimizations

### Concurrency & Resiliency

The Android platform chunks massive media files asynchronously in parallel (`maxConcurrency = 3`). The architecture features robust fault tolerance through an exponential backoff loop, capping at 3 retries for isolated chunk upload failures.

### Local Session Persistence (True Resume Capability)

To support resuming uploads across application restarts, the Android client utilizes `SharedPreferences` to persist the active `sessionId` linked to the absolute file path.

- The system seamlessly verifies this cache prior to instantiation, downloading `missing_ranges` to continue an interrupted transfer.
- It falls back to generating a new session if the cached ID is expired.
- The cache eviction policy operates explicitly upon a definitively successful pipeline submit.

## Android Realtime Event Synchronization

### MQTT Multiplexing Flow

The primary Android Progress UI dynamically mirrors the backend state reactively from MQTT socket broadcasts on `users/{mac}/{env}/task`.

### Fault-Tolerant Polling Recovery

To guarantee status resilience in unstable mobile networks, an HTTP poll ping gracefully runs in the background. It dynamically activates only when the active MQTT socket loses connection, or when the system starves, failing to receive an inbound progression beat for more than 10 seconds.

## Unified Meeting Insight Flow (File-First)

### Local App-Private Storage (`MeetingAudioFileStore`)

The Meeting Transcriber Screen strictly enforces a "File-First" policy. Regardless of whether the user captures audio natively via the microphone or selects a file via the OS picker, the audio stream is instantly saved as a permanent, uniquely-named file inside `context.filesDir/meetings/`.

### No-Delete Guarantee

Audio files generated in this guarded directory are never automatically deleted by the internal application flow. This architectural decision promises zero data loss during severe network disruption and perfectly sets up the "True Resume Capability" since every recording inherently holds a stable `absolutePath` needed for stable upload session resumption. _(Note: Files are still subject to lifecycle termination if the user explicitly clears app data or uninstalls the application)._
