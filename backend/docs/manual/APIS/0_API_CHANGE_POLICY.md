# API Change Policy & Scope Lock

As part of the AI Pipeline Refactor and ongoing development, the following rules apply strictly to all backend API endpoints.

## Affected Endpoints (Pipeline Refactor)

- `/api/speech/transcribe`
- `/api/rag/translate`
- `/api/rag/summary`
- `/api/speech/transcribe/status`
- `/api/rag/status`

## API Contract Change Rule

**NO PR SHALL BE MERGED** if the request or response contract changes without accompanying updates to:

1. Controller Swagger annotations (`backend/domain/.../controllers/...`)
2. Manual API docs under `backend/docs/manual/APIS/...`

### Required Doc Updates on Contract Change:

- Update request/response examples.
- Update status lifecycle examples.
- Update error cases.
- Add migration notes for client teams (Kotlin, iOS) comparing old vs. new payload fields.

### Compatibility

- Changes must be backward-compatible whenever possible.
- If a breaking change is unavoidable, it must either:
  1. Use versioned `v2` endpoints.
  2. Be accompanied by a published versioning/deprecation timeline in the documentation.
