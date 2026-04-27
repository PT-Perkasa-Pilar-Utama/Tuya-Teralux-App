package dtos

type PipelineStageStatus struct {
	Status          string      `json:"status" example:"pending"` // pending, processing, completed, failed, skipped, cancelled
	Result          interface{} `json:"result,omitempty"`
	Error           string      `json:"error,omitempty"`
	StartedAt       string      `json:"started_at,omitempty"`
	DurationSeconds float64     `json:"duration_seconds,omitempty"`
}

type PipelineStatusDTO struct {
	TaskID           string                         `json:"task_id"`
	OverallStatus    string                         `json:"overall_status" example:"processing"` // pending, processing, completed, failed, cancelled
	Stages           map[string]PipelineStageStatus `json:"stages"`
	StartedAt        string                         `json:"started_at"`
	DurationSeconds  float64                        `json:"duration_seconds"`
	ExpiresAt        string                         `json:"expires_at,omitempty"`
	ExpiresInSeconds int64                          `json:"expires_in_seconds,omitempty"`
	MacAddress       string                         `json:"mac_address,omitempty"`
	Version          uint64                         `json:"version"` // For atomic CAS updates - incremented on each status change
}

// SetExpiry implements tasks.StatusWithExpiry interface
func (s *PipelineStatusDTO) SetExpiry(expiresAt string, expiresInSeconds int64) {
	s.ExpiresAt = expiresAt
	s.ExpiresInSeconds = expiresInSeconds
}

type PipelineRequestDTO struct {
	Language       string   `form:"language" json:"language" example:"id"`
	TargetLanguage string   `form:"target_language" json:"target_language" example:"en"`
	Context        string   `form:"context" json:"context" example:"technical meeting"`
	Style          string   `form:"style" json:"style" example:"minutes"`
	Date           string   `form:"date" json:"date"`
	Location       string   `form:"location" json:"location"`
	Participants   []string `form:"participants" json:"participants"`
	Diarize        bool     `form:"diarize" json:"diarize"`
	Refine         *bool    `form:"refine" json:"refine"` // pointer to distinguish missing vs false
	Summarize      bool     `form:"summarize" json:"summarize"`
	MacAddress     string   `form:"mac_address" json:"mac_address"`
}

type PipelineResponseDTO struct {
	TaskID     string             `json:"task_id"`
	TaskStatus *PipelineStatusDTO `json:"task_status,omitempty"`
}
