# MQTT Transcription Ephemerality Policy

## Overview

This document defines the data persistence policy for MQTT-origin transcription requests in the Teralux AI assistant system.

## Policy Statement

**MQTT wake-word/assistant voice requests are transient by design and do NOT create database recording rows.**

## Rationale

### Design Decisions

1. **High-Frequency Interactions**: Wake-word triggered assistant interactions are conversational and high-frequency. Persisting every casual query would create unnecessary database write load.

2. **Conversational Nature**: Most wake-word interactions are casual queries (weather, timers, quick questions) that don't require long-term audit trails.

3. **Storage Efficiency**: Ephemeral transcription reduces storage costs and database growth rate significantly.

4. **Privacy Considerations**: Not persisting casual voice interactions aligns with privacy-by-design principles for always-listening devices.

## Implications

### For Frontend Developers

- **Do NOT assume `recording_id` exists** for MQTT-origin transcription requests
- MQTT path returns `recording_id=""` or omits it entirely from responses
- Historical transcription lookup is **not available** for MQTT requests
- Design UI flows that don't depend on persistent recording metadata

### For Backend Developers

- MQTT transcription flow bypasses the recording persistence use cases
- No `Recording` entity is created for MQTT requests
- Transcription status is cached (TTL-based) but not persisted to database
- ASR quality gate rejections (silent audio, hallucination) are also ephemeral

### For API Consumers

- HTTP upload endpoints (`POST /api/whisper/upload`) → **Persistent** (creates DB recording)
- MQTT whisper topics (`users/{mac}/{env}/whisper`) → **Ephemeral** (no DB recording)
- Choose the appropriate transport based on your persistence requirements

## When to Use HTTP Upload Instead

Use HTTP upload with DB persistence when:

1. **Meeting Recordings**: Transcriptions require audit trails or historical lookup
2. **Compliance**: Legal/regulatory requirements mandate data retention
3. **Searchability**: Transcriptions must be searchable in historical records
4. **Evidence**: Transcriptions may be used as evidence or official records
5. **User Export**: Users expect to export/download their transcription history

## Technical Implementation

### MQTT Flow (Ephemeral)

```
Wake Word → Record Audio → MQTT Publish → Transcribe → Chat Chain → Response
                ↓
          (No DB Recording Created)
                ↓
          Status cached with TTL only
```

### HTTP Upload Flow (Persistent)

```
Record Audio → HTTP Upload → Save Recording (DB) → Transcribe → Response
                    ↓
            Recording Entity Created
                    ↓
            Persistent in database
```

## Backend Code Reference

The ephemerality is enforced in the transcription use case:

- MQTT requests set `metadata.Source = "mqtt"`
- Recording persistence is skipped for MQTT source
- Only HTTP upload paths create `Recording` entities

See: `backend/domain/models/whisper/usecases/transcribe_usecase.go`

## Exception Handling

If you need to persist MQTT-origin transcriptions for specific use cases:

1. **Do NOT modify the default MQTT flow** (breaks contract with existing clients)
2. Add an explicit `persist` flag to the MQTT payload if needed
3. Or use HTTP upload endpoint for persistence-required scenarios
4. Document any exceptions clearly in API contracts

## Migration Notes

### For Existing Integrations

If your integration currently assumes MQTT transcription creates DB recordings:

1. **Check for `recording_id`** in responses (it will be empty for MQTT)
2. **Update error handling** to account for missing recording metadata
3. **Switch to HTTP upload** if you need persistence
4. **Update documentation** to reflect the ephemerality policy

### Future Changes

Any change to this policy (e.g., adding optional persistence to MQTT) requires:

1. API versioning (backward-compatible change)
2. Clear documentation updates
3. Client library updates
4. Migration guide for existing integrations

## Related Documentation

- `WHISPER/MEETING_PIPELINE_ARCHITECTURE.md` - Persistent meeting transcription flow
- `WHISPER/VOICE_CONTROL_PIPELINE_ARCHITECTURE.md` - Voice control flow
- `MQTTS/` - MQTT topic structure and messaging patterns

## Questions?

If you're unsure whether your use case should use MQTT (ephemeral) or HTTP (persistent):

- **Casual assistant queries** → MQTT (ephemeral) ✓
- **Meeting notes/recordings** → HTTP upload (persistent) ✓
- **Voice commands for automation** → MQTT (ephemeral) ✓
- **Legal/compliance recordings** → HTTP upload (persistent) ✓

When in doubt, default to HTTP upload for persistence requirements.
