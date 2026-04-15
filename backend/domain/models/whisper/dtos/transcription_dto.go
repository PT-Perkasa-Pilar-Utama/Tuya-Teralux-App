package dtos

// Utterance represents a single speaker turn with timing information
// WARNING: start_ms and end_ms are ESTIMATES based on text length heuristics
// when not explicitly provided by the transcription provider. They are NOT
// audio-aligned timestamps and should NOT be treated as precise evidence.
// Only trust timing data when the provider explicitly returns it.
type Utterance struct {
	SpeakerLabel string  `json:"speaker_label,omitempty"` // e.g., "Speaker 1", "John Doe"
	StartMs      int64   `json:"start_ms"`                // Start time in milliseconds (ESTIMATE if provider doesn't supply)
	EndMs        int64   `json:"end_ms"`                  // End time in milliseconds (ESTIMATE if provider doesn't supply)
	Text         string  `json:"text"`                    // Transcribed text for this utterance
	Confidence   float64 `json:"confidence,omitempty"`    // Confidence score (0.0-1.0) if available
}

// TranscriptSegment represents a chunk of transcript used in segmented/long-audio transcription
// WARNING: start_ms and end_ms are ESTIMATES based on cumulative text length heuristics
// when not explicitly provided by the transcription provider. They are NOT audio-aligned.
type TranscriptSegment struct {
	Index      int         `json:"index"`                // Segment index (0, 1, 2, ...)
	StartMs    int64       `json:"start_ms"`             // Segment start time in milliseconds (ESTIMATE)
	EndMs      int64       `json:"end_ms"`               // Segment end time in milliseconds (ESTIMATE)
	Text       string      `json:"text"`                 // Transcribed text for this segment
	Utterances []Utterance `json:"utterances,omitempty"` // Utterances within this segment, if available
}

// TranscriptFormat indicates the structure of the transcript
type TranscriptFormat string

const (
	TranscriptFormatPlainText     TranscriptFormat = "plain_text"
	TranscriptFormatUtteranceList TranscriptFormat = "utterance_list"
)

// ConfidenceSummary holds provider metadata about transcription confidence
type ConfidenceSummary struct {
	AverageConfidence float64 `json:"average_confidence,omitempty"`
	MinConfidence     float64 `json:"min_confidence,omitempty"`
	MaxConfidence     float64 `json:"max_confidence,omitempty"`
	SegmentsCount     int     `json:"segments_count,omitempty"`
	UtterancesCount   int     `json:"utterances_count,omitempty"`
}

// WhisperResult represents the result of a transcription from any provider
type WhisperResult struct {
	Transcription    string `json:"transcription"` // Plain text transcription (backward compatible)
	DetectedLanguage string `json:"detected_language"`
	Source           string `json:"source"`   // Which service was used: "Orion", "Local", "Gemini"
	Diarized         bool   `json:"diarized"` // Whether diarization was performed

	// Optional structured artifacts (backward compatible - empty when not available)
	Utterances        []Utterance         `json:"utterances,omitempty"`         // Ordered speaker turns
	Segments          []TranscriptSegment `json:"segments,omitempty"`           // Ordered transcript chunks
	TranscriptFormat  TranscriptFormat    `json:"transcript_format,omitempty"`  // Structure type
	ConfidenceSummary *ConfidenceSummary  `json:"confidence_summary,omitempty"` // Provider metadata
}

type WhisperMqttRequestDTO struct {
	RequestID  string `json:"request_id,omitempty"`
	Audio      string `json:"audio" binding:"required"` // Base64 encoded audio
	Language   string `json:"language,omitempty"`
	TerminalID string `json:"terminal_id" binding:"required"`
	UID        string `json:"uid,omitempty"`
	Diarize    bool   `json:"diarize,omitempty"`
}

type TranscriptionRequestDTO struct {
	Language string `form:"language" json:"language"`
	Diarize  bool   `form:"diarize" json:"diarize"`
}

// OutsystemsTranscriptionResultDTO represents the structured response from Outsystems
type OutsystemsTranscriptionResultDTO struct {
	Filename         string `json:"filename"`
	Transcription    string `json:"transcription" example:"Halo dunia"`
	DetectedLanguage string `json:"detected_language" example:"id"`
}

type TranscriptionTaskResponseDTO struct {
	TaskID      string `json:"task_id" example:"abc-123"`
	TaskStatus  string `json:"task_status" example:"pending"`
	RecordingID string `json:"recording_id,omitempty" example:"uuid-v4"`
	RequestID   string `json:"request_id,omitempty" example:"req-uuid"`
	Source      string `json:"source,omitempty" example:"MQTT_ACK"`
}

type MqttPublishRequest struct {
	Message string `json:"message" binding:"required"`
}

// Async Transcription DTOs (Consolidated)
type AsyncTranscriptionResultDTO struct {
	Transcription    string `json:"transcription" example:"Halo dunia"`
	RefinedText      string `json:"refined_text,omitempty" example:"Hello world"`
	DetectedLanguage string `json:"detected_language,omitempty" example:"id"`

	// Optional structured artifacts (backward compatible - empty when not available)
	Utterances           []Utterance         `json:"utterances,omitempty"`
	Segments             []TranscriptSegment `json:"segments,omitempty"`
	TranscriptFormat     TranscriptFormat    `json:"transcript_format,omitempty"`
	ConfidenceSummary    *ConfidenceSummary  `json:"confidence_summary,omitempty"`
	NormalizationApplied bool                `json:"normalization_applied,omitempty"` // Whether safe normalization was applied
}

type AsyncTranscriptionStatusDTO struct {
	Status          string                       `json:"status" example:"completed"`
	Result          *AsyncTranscriptionResultDTO `json:"result,omitempty"`
	Error           string                       `json:"error,omitempty" example:"service unavailable"`
	Trigger         string                       `json:"trigger,omitempty" example:"/api/whisper/models/gemini"`
	MacAddress      string                       `json:"mac_address,omitempty"`
	TerminalID      string                       `json:"terminal_id,omitempty"`
	HTTPStatusCode  int                          `json:"-"`
	StartedAt       string                       `json:"started_at,omitempty" example:"2026-02-21T11:00:00Z"`
	DurationSeconds float64                      `json:"duration_seconds,omitempty" example:"1.5"`
	ExpiresAt       string                       `json:"expires_at,omitempty"`
	ExpiresInSecond int64                        `json:"expires_in_seconds,omitempty"`
}

// SetExpiry implements tasks.StatusWithExpiry interface
func (s *AsyncTranscriptionStatusDTO) SetExpiry(expiresAt string, expiresInSeconds int64) {
	s.ExpiresAt = expiresAt
	s.ExpiresInSecond = expiresInSeconds
}

type AsyncTranscriptionProcessStatusResponseDTO struct {
	TaskID     string                       `json:"task_id"`
	TaskStatus *AsyncTranscriptionStatusDTO `json:"task_status,omitempty"`
}

type AsyncTranscriptionProcessResponseDTO = TranscriptionTaskResponseDTO

// StatusUpdateMessage is used for real-time notifications via MQTT
type StatusUpdateMessage struct {
	Type   string      `json:"type"`
	TaskID string      `json:"task_id"`
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}
