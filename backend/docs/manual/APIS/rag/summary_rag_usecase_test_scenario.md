# ENDPOINT: POST /api/rag/summary

## Description
Transforms a raw transcription into professional meeting minutes using an LLM. It supports multi-language output, context steering, and style selection.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body
- **Content-Type**: `application/json`
- **Required Fields**:
  - `text` (string): The raw transcription text.
- **Optional Fields**:
  - `language` (string): Target language ("id" or "en"). Default: "id".
  - `context` (string): Context of the meeting (e.g., "Technical", "Marketing").
  - `style` (string): Professional style (e.g., "formal", "bulleted").

## Test Scenarios

### 1. Generate Summary (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Request Body**:
```json
{
  "text": "budi bilang kita harus deploy besok jam 9. andi setuju.",
  "language": "id",
  "context": "Deployment Plan"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Summary generated successfully",
  "data": "# Notulen Rapat: Deployment Plan\n\n## 1. Agenda\nDiskusi jadwal deployment.\n\n## 2. Key Discussion Points\n- Rencana deployment besok pagi.\n\n## 3. Decisions Made\n- Deployment ditetapkan besok jam 09:00.\n\n## 4. Action Items\n- [ ] **Andi/Budi**: Eksekusi deployment."
}
```
*(Status: 200 OK)*

### 2. Generate English Summary from Indonesian Input
- **Request Body**:
```json
{
  "text": "rapat hari ini membahas fitur baru.",
  "language": "en"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Summary generated successfully",
  "data": "# Meeting Minutes\n\n## 1. Agenda\nDiscussion about new features..."
}
```

### 3. Validation: Whitespace Only
- **Method**: `POST`
- **Request Body**:
```json
{
  "text": "   "
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Invalid request",
  "details": "Key: 'RAGSummaryRequestDTO.Text' Error:Field validation for 'Text' failed on the 'required' tag"
}
```
*(Status: 400 Bad Request)*

### 4. Validation: Invalid Language Code
- **Method**: `POST`
- **Request Body**:
```json
{
  "text": "Valid text here",
  "language": "alien"
}
```
- **Expected Behavior**: The system internally defaults to "id" (Indonesian) for the prompt generation.
- **Expected Response**: `200 OK` with summary in Indonesian.

### 5. Technical Note: Maximum Length
- **Recommendation**: For transcriptions exceeding 100,000 characters (~15,000 words), it is recommended to chunk the text manually to ensure LLM prompt limits are not exceeded, though the system attempts to process long texts via internal optimization.
