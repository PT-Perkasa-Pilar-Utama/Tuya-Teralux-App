# API Contract Guide

Use this guide whenever the user asks to create, update, or review API contract docs.

## Primary Reference
- Existing pattern source: `docs/api/users/*.md`
- Keep wording, section order, and response envelope aligned with existing docs.

## File Naming Convention
- Path: `docs/api/<domain>/`
- Filename pattern: `<nn>_<method>_<endpoint_path_with_underscores>.md`
- Rules:
  - `<nn>` uses 2 digits (`01`, `02`, ...).
  - `<method>` lowercase HTTP method (`get`, `post`, `patch`, `put`, `delete`).
  - Endpoint path uses `_` and removes `/api/` prefix slashes.
  - Path params use plain token names (example `{user_id}` -> `user_id`).
- Example:
  - Endpoint: `PATCH /api/users/{user_id}/details`
  - File: `08_patch_api_users_user_id_details.md`

## Required Document Structure (Exact Order)
1. `# ENDPOINT: <METHOD> <path>`
2. `## Description`
3. `## Authentication`
4. `## Session Metadata Headers` (only if endpoint uses these headers)
5. `## Test Scenarios`
   - `### 1. Error: Internal Server Error`
   - `### 2. Error: Header`
   - `### 3. Error: Params`
   - `### 4. Error: Body`
   - `### 5. Success`

## Scenario Content Rules
- Scenario 1:
  - Always include `- **Method**:`
  - Include `- **Pre-conditions**:`
  - Include request sample when body is relevant.
  - Include `- **Expected Response**:` with JSON.
  - Include status line `*(Status: 500 Internal Server Error)*`.
  - Include `- **Response Header**: \`X-Trace-Id: ...\``.
- Scenario 2:
  - Header/auth failure, typically `401 Unauthorized`.
- Scenario 3:
  - Parameter validation (query/path) and `400`/`404` cases if relevant.
- Scenario 4:
  - Body validation cases.
  - For endpoints without body, use `Request body is not allowed` (`400`).
  - If multiple validation branches exist, use `Case 4.1`, `4.2`, etc.
- Scenario 5:
  - Success request sample (if relevant).
  - Success response sample.
  - Status must match method semantics (`200`, `201`, etc).
  - Include response headers when relevant (for example `Set-Cookie`).

## Response Envelope Conventions
- Error envelope:
```json
{
  "status": false,
  "message": "..."
}
```
- Validation error envelope:
```json
{
  "status": false,
  "message": "Validation failed",
  "errors": {
    "field": "reason"
  }
}
```
- Success envelope:
```json
{
  "status": true,
  "message": "...",
  "data": {
    "<resource_key>": {}
  }
}
```
- Use resource keys consistent with current docs (`users`, `user_details`).

## Wording Conventions
- Reuse existing standard messages when applicable:
  - `Internal server error`
  - `Unauthorized`
  - `Invalid query parameters`
  - `Invalid path parameters`
  - `Invalid request body`
  - `Validation failed`
  - `Data not found`
  - `Request body is not allowed`

## Formatting Conventions
- Use Markdown headings exactly as shown above.
- Use bullet labels in bold: `- **Method**:`, `- **Case**:`, etc.
- Use fenced `json` blocks for JSON payloads.
- Use inline backticks for headers, URLs, field names, and status labels.
- Keep status line format exactly: `*(Status: <code> <reason>)*`.

## Quality Checklist Before Finalizing
- Endpoint title, method, and path are correct.
- Section order matches this guide.
- All 5 scenario groups exist.
- Auth details match endpoint behavior (Header/Cookie/Bearer/API key).
- Error and success payloads follow existing envelope style.
- Status codes are consistent with described behavior.
- Filename numbering and naming pattern are consistent with folder sequence.
