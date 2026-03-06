package events

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTaskEventV1_JSONSerialization(t *testing.T) {
	event := NewTaskEventV1("test-task-123", "MeetingPipeline", "started", "processing")

	// Set some optional fields
	event.Stage = "transcription"
	event.StageStatus = "processing"
	event.ProgressPercent = 50

	b, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal TaskEventV1: %v", err)
	}

	jsonStr := string(b)

	// Validate required fields
	expectedFields := []string{
		`"version":1`,
		`"task_id":"test-task-123"`,
		`"task_type":"MeetingPipeline"`,
		`"event":"started"`,
		`"overall_status":"processing"`,
		`"stage":"transcription"`,
		`"stage_status":"processing"`,
		`"progress_percent":50`,
		`"ts":`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("Expected JSON to contain %s, got: %s", field, jsonStr)
		}
	}

	// Validate omitempty fields
	if strings.Contains(jsonStr, `"error"`) {
		t.Errorf("Expected JSON to omit 'error' when empty, got: %s", jsonStr)
	}

	// Now test omitempty on stage and progress
	emptyEvent := NewTaskEventV1("task-2", "Type", "accepted", "pending")
	bEmpty, _ := json.Marshal(emptyEvent)
	jsonStrEmpty := string(bEmpty)

	if strings.Contains(jsonStrEmpty, `"stage"`) || strings.Contains(jsonStrEmpty, `"stage_status"`) || strings.Contains(jsonStrEmpty, `"progress_percent"`) {
		t.Errorf("Expected JSON to omit empty optional fields, got: %s", jsonStrEmpty)
	}
}
