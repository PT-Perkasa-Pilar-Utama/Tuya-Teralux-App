package dtos

// ActionItem represents a structured action item extracted from meeting summary
type ActionItem struct {
	ID       int    `json:"id"`
	Task     string `json:"task"`
	PIC      string `json:"pic,omitempty"` // Person in charge (may be empty if not specified)
	Deadline string `json:"deadline,omitempty"`
	Status   string `json:"status,omitempty"`
}

// Decision represents a structured decision made during the meeting
type Decision struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Rationale   string `json:"rationale,omitempty"`
}

// OpenIssue represents an unresolved topic or question
type OpenIssue struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Owner       string `json:"owner,omitempty"` // May be empty if not assigned
}

// Risk represents an identified risk with mitigation strategy
type Risk struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Impact      string `json:"impact,omitempty"` // Low/Medium/High or 1-10 score
	Mitigation  string `json:"mitigation,omitempty"`
}

// SpeakerCoverage provides statistics about speaker attribution in the transcript
type SpeakerCoverage struct {
	TotalSpeakers     int            `json:"total_speakers,omitempty"`
	SpeakersWithNames int            `json:"speakers_with_names,omitempty"` // Speakers with identified names vs "Speaker 1"
	UtteranceCount    int            `json:"utterance_count,omitempty"`
	SpeakerBreakdown  map[string]int `json:"speaker_breakdown,omitempty"` // Speaker label -> utterance count
}

// CoverageStats provides statistics about transcript coverage in hierarchical summarization
type CoverageStats struct {
	TotalWindows     int     `json:"total_windows"`
	ProcessedWindows int     `json:"processed_windows"`
	EmptyWindows     int     `json:"empty_windows"`  // Windows with no extractable content
	CoverageRatio    float64 `json:"coverage_ratio"` // processed_windows / total_windows
	SourceChars      int     `json:"source_chars"`
	SummaryChars     int     `json:"summary_chars"`
	CompressionRatio float64 `json:"compression_ratio"`
}

// StructuredSummary encapsulates structured meeting summary artifacts
type StructuredSummary struct {
	SummaryVersion         string           `json:"summary_version,omitempty"` // e.g., "2.0-structured"
	SummaryMode            string           `json:"summary_mode,omitempty"`    // "single_pass" or "hierarchical_structured"
	ActionItems            []ActionItem     `json:"action_items,omitempty"`
	Decisions              []Decision       `json:"decisions,omitempty"`
	OpenIssues             []OpenIssue      `json:"open_issues,omitempty"`
	Risks                  []Risk           `json:"risks,omitempty"`
	CoverageStats          *CoverageStats   `json:"coverage_stats,omitempty"`
	SpeakerCoverage        *SpeakerCoverage `json:"speaker_coverage,omitempty"`
	SourceLanguage         string           `json:"source_language,omitempty"`
	TranslatedFromLanguage string           `json:"translated_from_language,omitempty"` // If translation was applied
}

type RAGRequestDTO struct {
	Text       string `json:"text" binding:"required" example:"Ini adalah transkrip panjang dari rapat teknis..."`
	Language   string `json:"language,omitempty" example:"id"`
	MacAddress string `json:"mac_address,omitempty" example:"AA:BB:CC:DD:EE:FF"`
}

type RAGSummaryRequestDTO struct {
	Text         string   `json:"text" binding:"required" example:"This is a long transcript of a technical meeting..."`
	Language     string   `json:"language,omitempty" example:"id"` // "id" or "en"
	Context      string   `json:"context,omitempty" example:"technical meeting"`
	Style        string   `json:"style,omitempty" example:"minutes"` // e.g., "minutes", "executive"
	Location     string   `json:"location,omitempty" example:"Meeting Room A"`
	Date         string   `json:"date"`
	Participants []string `json:"participants,omitempty" example:"[\"Alice\", \"Bob\"]"`
	MacAddress   string   `json:"mac_address,omitempty" example:"AA:BB:CC:DD:EE:FF"`
}

type RAGStatusDTO struct {
	Status          string            `json:"status" example:"completed"`
	Result          string            `json:"result,omitempty" example:"The meeting discussed..."`
	Summary         string            `json:"summary,omitempty"` // Alias for Result in summary tasks
	PDFUrl          string            `json:"pdf_url,omitempty"`
	AgendaContext   string            `json:"agenda_context,omitempty"`
	MeetingContext  string            `json:"meeting_context,omitempty"`
	Language        string            `json:"language,omitempty"`
	Error           string            `json:"error,omitempty" example:"gemini api returned status 503"`
	Trigger         string            `json:"trigger,omitempty" example:"/api/rag/summary"`
	MacAddress      string            `json:"mac_address,omitempty"`
	HTTPStatusCode  int               `json:"-"`
	StartedAt       string            `json:"started_at,omitempty" example:"2026-02-21T11:00:00Z"`
	DurationSeconds float64           `json:"duration_seconds,omitempty" example:"2.5"`
	Endpoint        string            `json:"endpoint,omitempty"`
	Method          string            `json:"method,omitempty"`
	Body            interface{}       `json:"body,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	ExecutionResult interface{}       `json:"execution_result,omitempty"` // holds the response from the fetched endpoint
	ExpiresAt       string            `json:"expires_at,omitempty"`
	ExpiresInSeconds int64             `json:"expires_in_seconds,omitempty"`

	// Optional structured summary artifacts (backward compatible - empty when not available)
	SummaryVersion         string           `json:"summary_version,omitempty"`
	SummaryMode            string           `json:"summary_mode,omitempty"` // "single_pass" or "hierarchical_structured"
	ActionItems            []ActionItem     `json:"action_items,omitempty"`
	Decisions              []Decision       `json:"decisions,omitempty"`
	OpenIssues             []OpenIssue      `json:"open_issues,omitempty"`
	Risks                  []Risk           `json:"risks,omitempty"`
	CoverageStats          *CoverageStats   `json:"coverage_stats,omitempty"`
	SpeakerCoverage        *SpeakerCoverage `json:"speaker_coverage,omitempty"`
	SourceLanguage         string           `json:"source_language,omitempty"`
	TranslatedFromLanguage string           `json:"translated_from_language,omitempty"`
}

// SetExpiry implements tasks.StatusWithExpiry interface.
// This allows the generic status usecase to automatically populate TTL info.
func (s *RAGStatusDTO) SetExpiry(expiresAt string, expiresInSeconds int64) {
	s.ExpiresAt = expiresAt
	s.ExpiresInSeconds = expiresInSeconds
}

// RAGProcessResponseDTO is the payload returned by POST /api/rag
// It contains the generated task id and optionally a `status` DTO when available.
type RAGProcessResponseDTO struct {
	TaskID     string        `json:"task_id"`
	TaskStatus *RAGStatusDTO `json:"task_status,omitempty"`
}

type RAGSummaryResponseDTO struct {
	Summary       string `json:"summary"`
	PDFUrl        string `json:"pdf_url,omitempty"`
	AgendaContext string `json:"agenda_context,omitempty"`

	// Optional structured summary artifacts (backward compatible - empty when not available)
	SummaryVersion         string           `json:"summary_version,omitempty"`
	SummaryMode            string           `json:"summary_mode,omitempty"`
	ActionItems            []ActionItem     `json:"action_items,omitempty"`
	Decisions              []Decision       `json:"decisions,omitempty"`
	OpenIssues             []OpenIssue      `json:"open_issues,omitempty"`
	Risks                  []Risk           `json:"risks,omitempty"`
	CoverageStats          *CoverageStats   `json:"coverage_stats,omitempty"`
	SpeakerCoverage        *SpeakerCoverage `json:"speaker_coverage,omitempty"`
	SourceLanguage         string           `json:"source_language,omitempty"`
	TranslatedFromLanguage string           `json:"translated_from_language,omitempty"`
}

type RAGChatRequestDTO struct {
	// RequestID is a unique identifier for cross-channel idempotency.
	// When the same request_id is sent via both HTTP and MQTT channels,
	// only ONE Tuya command execution will occur. The second channel will
	// return either a "processing" acknowledgment (if first request is still
	// in progress) or the cached response (if first request completed).
	// Client should generate a UUID per user interaction and use it for both
	// HTTP and MQTT dispatch to prevent duplicate device control.
	RequestID  string `json:"request_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Prompt     string `json:"prompt" binding:"required" example:"Nyalakan AC"`
	Language   string `json:"language,omitempty" example:"id"`
	TerminalID string `json:"terminal_id" binding:"required" example:"tx-1"`
	UID        string `json:"uid,omitempty" example:"sg1765..."` // Must be Tuya UID (never MAC/terminal identity)
}

type RAGChatResponseDTO struct {
	Response       string       `json:"response,omitempty"`
	IsControl      bool         `json:"is_control,omitempty"`
	IsBlocked      bool         `json:"is_blocked"`
	Redirect       *RedirectDTO `json:"redirect,omitempty"`
	HTTPStatusCode int          `json:"-"`                     // HTTP status code to return (not exposed in JSON)
	RequestID      string       `json:"request_id,omitempty"`  // Tracking ID (echoes request_id from request)
	Source         string       `json:"source,omitempty"`      // Response source: "HTTP_HANDLER", "MQTT_SUBSCRIBER", "IDEMPOTENCY_CACHED", "IDEMPOTENCY_IN_PROGRESS", "MQTT_SYNC_DROP"
	InstanceID     string       `json:"instance_id,omitempty"` // Server start time

	// Idempotency Source Contract:
	// - "IDEMPOTENCY_CACHED": Duplicate request with same request_id, returning cached completed response
	// - "IDEMPOTENCY_IN_PROGRESS": Duplicate request with same request_id, first request still processing
	// - "MQTT_SYNC_DROP": Text query dropped because Whisper transcription is active for same terminal
	// - "HTTP_HANDLER" / "MQTT_SUBSCRIBER": Fresh request processed via respective channel
	// Frontend should use these markers to suppress duplicate UI updates and handle silent-complete flows.
}

type RedirectDTO struct {
	Endpoint string      `json:"endpoint"`
	Method   string      `json:"method"`
	Body     interface{} `json:"body,omitempty"`
}

type RAGControlRequestDTO struct {
	Prompt     string `json:"prompt" binding:"required" example:"Nyalakan AC"`
	TerminalID string `json:"terminal_id" binding:"required" example:"tx-1"`
}

type ControlResultDTO struct {
	Message        string `json:"message"`
	DeviceID       string `json:"device_id,omitempty"`
	Command        string `json:"command,omitempty"` // e.g., "turn_on", "turn_off", "set_temp_24"
	HTTPStatusCode int    `json:"-"`                 // HTTP status code to return (not exposed in JSON)
}

// RAGRawPromptRequestDTO represents a raw prompt request to a specific model.
type RAGRawPromptRequestDTO struct {
	Prompt string `json:"prompt" binding:"required" example:"Hello, how are you?"`
}

// RAGRawPromptResponseDTO represents the direct string response from an LLM model, formatted to match the Speech tracking style.
type RAGRawPromptResponseDTO struct {
	Status          string  `json:"status" example:"completed"`
	Error           string  `json:"error,omitempty"`
	Trigger         string  `json:"trigger,omitempty"`
	HTTPStatusCode  int     `json:"-"`
	StartedAt       string  `json:"started_at,omitempty"`
	DurationSeconds float64 `json:"duration_seconds,omitempty"`
	Result          string  `json:"result,omitempty"`
}
