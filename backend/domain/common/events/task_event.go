package events

import (
	"time"
)

// TaskEventV1 represents the standard MQTT payload for task lifecycle and progress tracking.
// This supports real-time pipeline status updates on the client side.
type TaskEventV1 struct {
	Version         int    `json:"version"` // Always 1 for this version
	TaskID          string `json:"task_id"`
	TaskType        string `json:"task_type"`      // e.g. "MeetingPipeline", "Transcription"
	Event           string `json:"event"`          // "accepted", "started", "stage_update", "completed", "failed"
	OverallStatus   string `json:"overall_status"` // "pending", "processing", "completed", "failed"
	Stage           string `json:"stage,omitempty"`
	StageStatus     string `json:"stage_status,omitempty"` // "pending", "processing", "completed", "failed", "skipped"
	ProgressPercent int    `json:"progress_percent,omitempty"`
	Error           string `json:"error,omitempty"`
	Timestamp       string `json:"ts"` // RFC3339 formatted time
}

// NewTaskEventV1 creates a new TaskEventV1 with the default version and current timestamp.
func NewTaskEventV1(taskID, taskType, event, overallStatus string) *TaskEventV1 {
	return &TaskEventV1{
		Version:       1,
		TaskID:        taskID,
		TaskType:      taskType,
		Event:         event,
		OverallStatus: overallStatus,
		Timestamp:     time.Now().Format(time.RFC3339),
	}
}
