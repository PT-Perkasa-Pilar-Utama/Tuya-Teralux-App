# ENDPOINT: POST /api/pipeline/job

## Description

Unified AI pipeline that orchestrates multiple stages in a single job. This is the preferred endpoint for the mobile app as it reduces network overhead and ensures atomic processing of the meeting intelligence flow.

### Processing Flow

1. **Transcription**: Converts audio to text using configured provider.
2. **Refinement (Optional)**: Fixes grammar and spelling of the transcript.
3. **Translation (Optional)**: Translates the refined text to the `target_language`.
4. **Summary (Optional)**: Generates meeting minutes from the final text.

## Authentication

- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body

- **Content-Type**: `multipart/form-data`
- **Headers**:
  - `Idempotency-Key` (string, optional): Unique key to deduplicate requests.
- **Parameters**:
  - `audio` (file, required): Audio file.
  - `language` (string, required): Source language (e.g., "id", "en").
  - `target_language` (string, optional): Target language for translation/summary.
  - `refine` (boolean, optional): Default `true`. Grammar/spelling polish.
  - `summarize` (boolean, optional): Default `false`. Generate professional MoM.
  - `context` (string, optional): Meeting context for summary.
  - `style` (string, optional): Summary style (e.g., "minutes").
  - `date`, `location`, `participants`: Metadata for the summary report.
  - `diarize` (boolean, optional): Speaker identification.

## Example Response

```json
{
  "status": true,
  "message": "Pipeline job submitted successfully",
  "data": {
    "task_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

## Status Tracking

- **Endpoint**: `GET /api/pipeline/status/:task_id`
- **Payload**: Includes an `overall_status` and a map of `stages` with individual results and durations.

### Stage Keys

- `transcription`: Initial audio-to-text.
- `translation`: Translated text (if `target_language` provided).
- `summary`: Professional MoM and PDF link (if `summarize` is true).

### Successful Summary Result Structure

When the `summary` stage is `completed`, its `result` contains:

- `summary` (string): The actual AI-generated summary text.
- `pdf_url` (string): Relative URL to the generated PDF report.
- `agenda_context` (string): The context used to generate the summary.

## Example Response (Job Completed)

```json
{
  "status": true,
  "message": "Pipeline status retrieved",
  "data": {
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "overall_status": "completed",
    "stages": {
      "transcription": {
        "status": "completed",
        "result": "Meeting transcript content...",
        "duration_ms": 1200
      },
      "summary": {
        "status": "completed",
        "result": {
          "summary": "Meeting summary text...",
          "pdf_url": "uploads/pdf/report.pdf",
          "agenda_context": "Project X Sync"
        },
        "duration_ms": 3500
      }
    }
  }
}
```

## Example Request

```bash
curl -X POST http://localhost:8080/api/pipeline/job \
  -H "Authorization: Bearer <token>" \
  -H "Idempotency-Key: mobile-meeting-123" \
  -F "audio=@meeting.mp3" \
  -F "language=id" \
  -F "target_language=en" \
  -F "summarize=true"
```
