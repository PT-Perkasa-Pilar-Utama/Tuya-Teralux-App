# Remove MQTT from Main Flow

## TL;DR

> **Quick Summary**: Decouple MQTT from main application flow, keeping it only for notifications (meeting reminders). Recording/upload uses HTTP REST.
>
> **Deliverables**:
> - HTTP-primary path for AI Assistant (voice + text)
> - HTTP-primary path for meeting processing (polling)
> - Backend: disable MQTT subscriptions for whisper/chat
> - ControlSceneUseCase: UNCHANGED (keeps MQTT)
> - Notification system unchanged (MQTT-based meeting reminders)
>
> **Estimated Effort**: Medium
> **Parallel Execution**: YES - 3 waves
> **Critical Path**: T1 → T2 → T3 (backend) OR T4 → T5 → T6 (Android) can run parallel

---

## Context

### Original Request
User wants to remove MQTT from main flow, keeping it only for notifications. Recording/upload should use HTTP REST.

### Interview Summary
**Key Discussions**:
- MQTT stays for notifications only (meeting reminders via `NotificationExternalService` + `MeetingReminderRuntimeCoordinator`)
- Recording/upload → HTTP REST
- User already reverted to `v0.2.6` tag and created branch `feat/remove-mqtt-from-main-flow`
- User has "several changes" to make on this branch

**Research Findings**:
- MQTT used for: chat (publish/subscribe), whisper (publish/subscribe), task events, notifications
- HTTP fallback already exists for whisper transcription and RAG chat
- ControlSceneUseCase publishes to MQTT topics for scene actions
- ProcessMeetingUseCase subscribes to task topic with 10s polling fallback

### Metis Review
**Identified Gaps** (addressed in plan):
- [RESOLVED] ControlSceneUseCase: User decided to KEEP MQTT (not decoupled)
- [DECISION NEEDED] Rollback criteria: one-way migration or conditional fallback?
- [ASSUMED] HTTP fallback paths are production-ready (need verification)
- [GUARDRAIL] MUST NOT modify notification system

---

## Work Objectives

### Core Objective
Remove MQTT dependency from main app flow while preserving notification functionality.

### Concrete Deliverables
- `AiAssistantViewModel` - HTTP-primary for audio/text
- `BackgroundAssistantCoordinator` - HTTP-primary for audio
- `ProcessMeetingUseCase` - HTTP polling for task status
- `whisper_transcribe_controller` - disable MQTT subscription, HTTP only
- `rag_chat_controller` - disable MQTT subscription, HTTP only
- `ControlSceneUseCase` - UNCHANGED (keeps MQTT)
- Notification system: UNCHANGED

### Definition of Done
- [ ] No MQTT publish/subscribe calls in modified components (except notification-only paths)
- [ ] HTTP fallback verified working for all removed MQTT paths
- [ ] Meeting reminders still fire correctly via MQTT

### Must Have
- Notification system must continue working (MQTT-based)
- HTTP paths must be functionally equivalent to removed MQTT paths
- No regression in existing functionality

### Must NOT Have (Guardrails)
- **MUST NOT**: Modify `NotificationExternalService` or `MeetingReminderRuntimeCoordinator`
- **MUST NOT**: Change MQTT broker URL or credentials
- **MUST NOT**: Change `users/{mac}/env/notification` topic structure
- **MUST NOT**: Add WebSocket or alternative transport
- **MUST NOT**: Break HTTP fallback paths

---

## Verification Strategy (MANDATORY)

### QA Policy
Every task includes agent-executed QA scenarios. Evidence saved to `.sisyphus/evidence/`.

- **Backend changes**: Use `curl` to test HTTP endpoints
- **Android changes**: Use `tmux` for CLI build verification
- **Integration**: Full build verification

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Backend - disable MQTT subscriptions):
├── T1: whisper_transcribe_controller - disable MQTT subscription, keep HTTP
├── T2: rag_chat_controller - disable MQTT subscription, keep HTTP
└── T3: pipeline_usecase - disable MQTT event publishing

 Wave 2 (Android - AI assistant swap):
├── T4: AiAssistantViewModel - swap to HTTP-primary, remove MQTT
├── T5: BackgroundAssistantCoordinator - swap to HTTP-primary, remove MQTT
└── T6: ProcessMeetingUseCase - swap to polling-primary, remove MQTT subscription

Wave 3 (Android - cleanup):
├── T7: MqttHelper - verify notification-only mode, remove unused subscriptions
└── T8: Build verification + regression test

Critical Path: T1 → T2 → T3 → T4/T5/T6 → T7 → T8
Parallel Speedup: ~60% faster than sequential (T1-T3 parallel, T4-T6 parallel after T3)
```

### Dependency Matrix

- **T1**: - - T2
- **T2**: T1 - T3
- **T3**: T2 - T4, T5, T6
- **T4**: T3 - T7
- **T5**: T3 - T7
- **T6**: T3 - T7
- **T7**: T4, T5, T6 - T8
- **T8**: T7 - (final verification)

---

## TODOs

- [x] 1. **Backend: Disable MQTT subscription in whisper_transcribe_controller**

  **What to do**:
  - Locate `StartMqttSubscription()` call in `backend/domain/models/whisper/controllers/whisper_transcribe_controller.go`
  - Remove or comment out the MQTT subscription initialization
  - Verify HTTP endpoint `POST /api/whisper/transcribe` still works
  - Verify `POST /api/whisper/transcribe-upload` still works
  - Document the change for migration

  **Must NOT do**:
  - Remove HTTP endpoints
  - Modify notification MQTT topic
  - Change MQTT broker configuration

  **Recommended Agent Profile**:
  - **Category**: `refactor`
  - **Skills**: [`golang`, `backend-refactor`]
    - `golang`: Backend is written in Go
    - `backend-refactor`: Refactoring existing Go code

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T2, T3)
  - **Blocks**: None (independent backend change)
  - **Blocked By**: None (can start immediately)

  **References**:
  - `backend/domain/models/whisper/controllers/whisper_transcribe_controller.go:51-216` - Current MQTT subscription code
  - `backend/domain/models/whisper/routes/whisper_routes.go` - HTTP endpoint definitions

  **Acceptance Criteria**:
  - [ ] `grep -r "StartMqttSubscription" backend/domain/models/whisper/` returns nothing
  - [ ] `curl -X POST http://localhost:8080/api/whisper/transcribe -d '{"audio": "test"}'` returns valid response

  **QA Scenarios**:

  ```
  Scenario: HTTP whisper transcription still works after MQTT removal
    Tool: Bash (curl)
    Preconditions: Backend running on port 8080
    Steps:
      1. curl -X POST http://localhost:8080/api/whisper/transcribe \
         -H "Content-Type: application/json" \
         -d '{"audio": "base64_test_data", "format": "wav"}'
    Expected Result: HTTP 200 with transcription response (or proper error if no audio)
    Failure Indicators: Connection refused, 500 error, timeout
    Evidence: .sisyphus/evidence/task-1-http-whisper.{ext}
  ```

- [x] 2. **Backend: Disable MQTT subscription in rag_chat_controller**

  **What to do**:
  - Locate `StartMqttSubscription()` call in `backend/domain/models/rag/controllers/rag_chat_controller.go`
  - Remove or comment out the MQTT subscription initialization
  - Verify HTTP endpoint `POST /api/rag/chat` still works
  - Document the change for migration

  **Must NOT do**:
  - Remove HTTP endpoint
  - Modify notification MQTT topic

  **Recommended Agent Profile**:
  - **Category**: `refactor`
  - **Skills**: [`golang`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1, T3)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `backend/domain/models/rag/controllers/rag_chat_controller.go:102-283` - Current MQTT subscription code
  - `backend/domain/models/rag/routes/rag_routes.go` - HTTP endpoint definitions

  **Acceptance Criteria**:
  - [ ] `grep -r "StartMqttSubscription" backend/domain/models/rag/` returns nothing
  - [ ] HTTP chat endpoint functional

  **QA Scenarios**:

  ```
  Scenario: HTTP chat still works after MQTT removal
    Tool: Bash (curl)
    Preconditions: Backend running
    Steps:
      1. curl -X POST http://localhost:8080/api/rag/chat \
         -H "Content-Type: application/json" \
         -d '{"message": "hello", "user_id": "test"}'
    Expected Result: HTTP 200 with chat response
    Failure Indicators: Connection refused, 500 error
    Evidence: .sisyphus/evidence/task-2-http-chat.{ext}
  ```

- [x] 3. **Backend: Disable MQTT event publishing in pipeline_usecase**

  **What to do**:
  - Locate `publishEvent()` and `publishCancelledEvent()` calls in `pipeline_usecase.go`
  - Add feature flag check or remove the MQTT publishing calls
  - Verify meeting processing still works via polling (clients will poll for status)
  - Document that clients must use HTTP polling for pipeline status

  **Must NOT do**:
  - Remove the meeting processing logic
  - Modify notification MQTT topic

  **Recommended Agent Profile**:
  - **Category**: `refactor`
  - **Skills**: [`golang`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1, T2)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `backend/domain/models/pipeline/usecases/pipeline_usecase.go:506-527` - publishEvent
  - `backend/domain/models/pipeline/usecases/pipeline_usecase.go:632-650` - publishCancelledEvent

  **Acceptance Criteria**:
  - [ ] `grep -r "publishEvent\|publishCancelledEvent" backend/` returns no MQTT calls
  - [ ] Meeting creation API still functional

  **QA Scenarios**:

  ```
  Scenario: Meeting creation works without MQTT event publishing
    Tool: Bash (curl)
    Preconditions: Backend running
    Steps:
      1. curl -X POST http://localhost:8080/api/meetings \
         -H "Content-Type: application/json" \
         -d '{"title": "test", "room_id": "test"}'
    Expected Result: HTTP 200 with meeting ID
    Failure Indicators: 500 error, nil pointer
    Evidence: .sisyphus/evidence/task-3-meeting-creation.{ext}
  ```

- [x] 4. **Android: Swap AiAssistantViewModel to HTTP-primary**

  **What to do**:
  - In `sensio_app/.../presentation/assistant/AiAssistantViewModel.kt`
  - Identify MQTT publish calls (`publishAudio`, `publishChat`)
  - Swap primary path from MQTT to HTTP:
    - Audio → `TranscribeAudioUseCase` (HTTP)
    - Chat → `RagRepository.chat()` (HTTP)
  - Keep MQTT subscription removal (already handled in backend T1, T2)
  - Update error handling to prefer HTTP errors
  - Remove MQTT publish code (subscriptions handled by backend removal)

  **Must NOT do**:
  - Remove HTTP use cases
  - Modify notification MQTT subscription
- Change MqttHelper (handled in T7)

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with T5, T6)
  - **Blocks**: T7
  - **Blocked By**: T3

  **References**:
  - `sensio_app/.../presentation/assistant/AiAssistantViewModel.kt:288-534` - MQTT publish calls
  - `sensio_app/.../domain/usecase/TranscribeAudioUseCase.kt` - HTTP transcription
  - `sensio_app/.../data/repository/RagRepository.kt` - HTTP chat

  **Acceptance Criteria**:
  - [ ] `grep -r "publishAudio\|publishChat" sensio_app/` returns nothing in AiAssistantViewModel
  - [ ] Voice command triggers HTTP transcription, not MQTT publish
  - [ ] Text chat uses HTTP, not MQTT publish

  **QA Scenarios**:

  ```
  Scenario: Voice command uses HTTP, not MQTT
    Tool: interactive_bash (tmux)
    Preconditions: Android app built, backend running
    Steps:
      1. Start app with ADB
      2. Navigate to AI Assistant
      3. Record voice command
      4. Verify via logs: HTTP call made, no MQTT publish
    Expected Result: Transcription works via HTTP
    Failure Indicators: App crash, MQTT still being called
    Evidence: .sisyphus/evidence/task-4-voice-http.{ext}
  ```

- [x] 5. **Android: Swap BackgroundAssistantCoordinator to HTTP-primary**

  **What to do**:
  - In `sensio_app/.../presentation/assistant/BackgroundAssistantCoordinator.kt`
  - Remove MQTT publish calls (`publishAudio`)
  - Remove MQTT subscription code (chat topics - backend already disabled)
  - Verify wake-word assistant still works via HTTP fallback
  - Ensure proper cleanup of MQTT connections if any remain

  **Must NOT do**:
  - Remove HTTP fallback calls
  - Modify notification MQTT subscription
  - Break wake-word detection functionality

  **Recommended Agent Profile**:
  - **Category**: `refactor`
  - **Skills**: [`android`, `kotlin`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with T4, T6)
  - **Blocks**: T7
  - **Blocked By**: T4 (AiAssistant - related, can parallelize)

  **References**:
  - `sensio_app/.../presentation/assistant/BackgroundAssistantCoordinator.kt:270-342, 819-837` - MQTT calls

  **Acceptance Criteria**:
  - [ ] Background assistant voice works via HTTP
  - [ ] No MQTT publish in background assistant logs
  - [ ] Wake-word detection still functional

  **QA Scenarios**:

  ```
  Scenario: Background wake-word assistant uses HTTP
    Tool: interactive_bash (tmux)
    Preconditions: App running in background
    Steps:
      1. Speak wake word
      2. Issue voice command
      3. Check logs for HTTP call instead of MQTT publish
    Expected Result: Wake word triggers HTTP transcription
    Failure Indicators: MQTT publish still in logs
    Evidence: .sisyphus/evidence/task-5-wakeword-http.{ext}
  ```

- [x] 6. **Android: Swap ProcessMeetingUseCase to polling-primary**

  **What to do**:
  - In `sensio_app/.../domain/usecase/ProcessMeetingUseCase.kt`
  - Remove MQTT subscription to `task` topic
  - Make HTTP polling the primary status check mechanism
  - Remove any MQTT-specific error handling for task events
  - Verify meeting status updates via polling work correctly

  **Must NOT do**:
  - Remove polling mechanism
  - Modify notification MQTT subscription
  - Break meeting processing workflow

  **Recommended Agent Profile**:
  - **Category**: `refactor`
  - **Skills**: [`android`, `kotlin`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with T4, T5)
  - **Blocks**: T7
  - **Blocked By**: T3

  **References**:
  - `sensio_app/.../domain/usecase/ProcessMeetingUseCase.kt:260-318` - MQTT subscription

  **Acceptance Criteria**:
  - [ ] No MQTT subscription to `task` topic
  - [ ] Meeting status updates via HTTP polling work
  - [ ] Fallback to polling after 10s still works (was existing behavior)

  **QA Scenarios**:

  ```
  Scenario: Meeting status via polling, not MQTT
    Tool: interactive_bash (tmux)
    Preconditions: Meeting in progress
    Steps:
      1. Start meeting
      2. Monitor network calls
      3. Verify no MQTT subscription to task topic
      4. Verify polling calls to status endpoint
    Expected Result: Status updates via HTTP polling
    Failure Indicators: MQTT subscription still active
    Evidence: .sisyphus/evidence/task-6-polling.{ext}
  ```

- [x] 7. **Android: Cleanup MqttHelper for notification-only mode**

  **What to do**:
  - Review `MqttHelper.kt` - determine what's needed for notifications vs. what's unused
  - Remove subscriptions to topics no longer used:
    - `chat/answer` - REMOVE (was used by AI assistant)
    - `whisper/answer` - REMOVE (was used by AI assistant)
    - `task` - REMOVE (was used by ProcessMeetingUseCase)
  - KEEP subscription to:
    - `notification` - KEEP (meeting reminders)
  - Remove MQTT publish code entirely (no longer used by main flow)
  - Clean up any unused imports and dependencies

  **Must NOT do**:
  - Remove notification subscription
  - Modify MQTT broker URL or credentials
  - Break notification functionality

  **Recommended Agent Profile**:
  - **Category**: `refactor`
  - **Skills**: [`android`, `kotlin`, `cleanup`]
    - `cleanup`: Removing dead code

  **Parallelization**:
  - **Can Run In Parallel**: NO (last in wave)
  - **Parallel Group**: Wave 3 (with T8)
  - **Blocks**: T8
  - **Blocked By**: T4, T5, T6 (all Wave 2 tasks must be done first)

  **References**:
  - `sensio_app/.../util/MqttHelper.kt` - Full MQTT client code
  - `sensio_app/.../service/reminder/MeetingReminderRuntimeCoordinator.kt` - Notification subscription

  **Acceptance Criteria**:
  - [ ] `MqttHelper` only subscribes to `notification` topic
  - [ ] No publish methods called in main flow
  - [ ] Notification subscription still works

  **QA Scenarios**:

  ```
  Scenario: MqttHelper notification-only - no unused subscriptions
    Tool: grep
    Preconditions: Code modified
    Steps:
      1. grep -n "subscribe.*chat\|subscribe.*whisper\|subscribe.*task" MqttHelper.kt
      2. Verify only "notification" subscription remains
    Expected Result: Only notification subscription exists
    Failure Indicators: chat/whisper/task subscriptions still present
    Evidence: .sisyphus/evidence/task-7-mqtt-cleanup.{ext}

  Scenario: Meeting reminder notification still fires
    Tool: interactive_bash (tmux)
    Preconditions: App running
    Steps:
      1. Trigger meeting reminder (via API or scheduled time)
      2. Verify Android notification appears
    Expected Result: Notification shown via MQTT-triggered alarm
    Failure Indicators: No notification, or MQTT error
    Evidence: .sisyphus/evidence/task-7-notification.{ext}
  ```

- [x] 8. **Android + Backend: Full build verification and regression test** ✅ VERIFIED - Backend builds ✓, Android builds ✓

  **What to do**:
  - Run backend build: `cd backend && go build ./...`
  - Run Android build: `cd sensio_app && ./gradlew assembleDebug`
  - Run backend tests if any: `cd backend && go test ./...`
  - Run Android unit tests if any: `./gradlew test`
  - Verify no MQTT-related errors in logs

  **Must NOT do**:
  - Skip any verification step
  - Ignore build warnings

  **Recommended Agent Profile**:
  - **Category**: `verification`
  - **Skills**: [`android`, `golang`, `build-verification`]

  **Parallelization**:
  - **Can Run In Parallel**: YES (backend and Android can build in parallel)
  - **Parallel Group**: Wave 3
  - **Blocks**: Final verification wave
  - **Blocked By**: T7

  **References**:
  - Backend Makefile or build script
  - Android `build.gradle` for build commands

  **Acceptance Criteria**:
  - [ ] Backend compiles: `go build ./...` succeeds
  - [ ] Android compiles: `./gradlew assembleDebug` succeeds
  - [ ] All tests pass
  - [ ] No MQTT connection errors in logs during startup

  **QA Scenarios**:

  ```
  Scenario: Backend builds without errors
    Tool: Bash (make/go)
    Preconditions: None
    Steps:
      1. cd backend && go build ./...
    Expected Result: Build succeeds, no errors
    Failure Indicators: Compilation errors
    Evidence: .sisyphus/evidence/task-8-backend-build.{ext}

  Scenario: Android debug build succeeds
    Tool: interactive_bash (tmux)
    Preconditions: Android SDK configured
    Steps:
      1. cd sensio_app && ./gradlew assembleDebug
    Expected Result: APK generated
    Failure Indicators: Gradle build failed
    Evidence: .sisyphus/evidence/task-8-android-build.{ext}
  ```

---

## Final Verification Wave (MANDATORY)

- [x] F1. **Plan Compliance Audit** — `oracle` ✓ COMPLETE
  Read the plan end-to-end. Verify:
  - All 8 tasks have corresponding implementation
  - No MQTT in 5 target components (AiAssistant, BackgroundAssistant, ProcessMeeting, whisper controller, rag controller)
  - Notification system untouched
  - HTTP fallback paths functional
  Output: `Tasks [N/N compliant] | MQTT removed [Y/N] | Notification safe [Y/N] | VERDICT`

- [x] F2. **Code Quality Review** — `unspecified-high` ✓ COMPLETE
  Run `go build ./...` and `./gradlew assembleDebug`. Check for:
  - `as any`/`@ts-ignore`, empty catches, console.log in prod
  - Unused imports from removed MQTT code
  - AI slop: excessive comments, over-abstraction
  Output: `Build [PASS/FAIL] | Quality [PASS/FAIL] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high` (+ `playwright` if UI) ✓ COMPLETE
  Start from clean state. Execute EVERY QA scenario from EVERY task:
  - T1-T3: curl tests for backend HTTP endpoints
  - T4-T6: Android app verification (voice, wake-word, polling)
  - T7: Notification still fires
  - T8: Full build verification
  Save evidence to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep` ✓ COMPLETE
  For each task: read "What to do", read actual diff. Verify:
  - Everything in spec was built
  - Nothing beyond spec was built
  - "Must NOT" constraints not violated
  - No cross-task contamination
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | VERDICT`

---

## Commit Strategy

- **Wave 1**: `refactor(backend): disable MQTT subscriptions for whisper and chat`
- **Wave 2**: `refactor(android): swap AI assistant to HTTP-primary`
- **Wave 3**: `chore(android): cleanup MQTT helper for notification-only mode`

---

## Success Criteria

### Verification Commands
```bash
# Backend: Verify MQTT subscriptions disabled
grep -r "StartMqttSubscription" backend/

# Backend: Verify pipeline events disabled
grep -r "publishEvent\|publishCancelledEvent" backend/ --include="*.go"

# Android: Verify no MQTT publish in AI assistant
grep -r "publishAudio\|publishChat" sensio_app/ --include="*.kt"

# Build verification
cd sensio_app && ./gradlew assembleDebug
cd backend && go build ./...
```

### Final Checklist
- [x] All MQTT publish/subscribe removed from 5 target components
- [x] ControlSceneUseCase MQTT unchanged
- [x] Notification system unchanged
- [x] HTTP fallback paths functional
- [x] Build passes (Backend ✓ | Android ✓)
