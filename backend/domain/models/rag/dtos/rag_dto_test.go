package dtos

import (
	"encoding/json"
	"testing"
)

func TestRAGStatusDTO_SerializesCanonicalSummaryFields(t *testing.T) {
	dto := RAGStatusDTO{
		Status:  "completed",
		Summary: "Meeting summary markdown",
		CanonicalSummary: &CanonicalMeetingSummary{
			Metadata: SummaryMetadata{
				MeetingTitle: "Test Meeting",
				Language:     "en",
			},
			Agenda: "Test agenda",
			ActionItems: []ActionItem{
				{Task: "Do something", PIC: "Alice"},
			},
		},
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("failed to marshal RAGStatusDTO: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Check canonical_summary field exists
	canonical, ok := result["canonical_summary"].(map[string]interface{})
	if !ok {
		t.Fatal("expected canonical_summary field in JSON output")
	}

	// Check nested fields
	metadata, ok := canonical["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("expected metadata field in canonical_summary")
	}
	if metadata["meeting_title"] != "Test Meeting" {
		t.Errorf("expected meeting_title to be 'Test Meeting', got %v", metadata["meeting_title"])
	}

	// Check action items
	actionItems, ok := canonical["action_items"].([]interface{})
	if !ok || len(actionItems) == 0 {
		t.Fatal("expected action_items in canonical_summary")
	}
}

func TestRAGSummaryResponseDTO_BackwardCompatibleFields(t *testing.T) {
	dto := RAGSummaryResponseDTO{
		Summary: "Summary markdown content",
		PDFUrl:  "https://example.com/pdf/123",
		ActionItems: []ActionItem{
			{Task: "Task 1", PIC: "Bob"},
		},
		Decisions: []Decision{
			{Description: "Decision 1"},
		},
		CanonicalSummary: &CanonicalMeetingSummary{
			Metadata: SummaryMetadata{
				MeetingTitle: "Meeting",
				Language:     "en",
			},
			Agenda: "Agenda",
		},
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("failed to marshal RAGSummaryResponseDTO: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Backward compatible fields must exist
	if result["summary"] != "Summary markdown content" {
		t.Error("expected 'summary' field to be backward compatible")
	}
	if result["pdf_url"] != "https://example.com/pdf/123" {
		t.Error("expected 'pdf_url' field to be backward compatible")
	}
	if result["canonical_summary"] == nil {
		t.Error("expected 'canonical_summary' field to be present")
	}
}

func TestRAGStatusDTO_JSONMarshaling_OmitsEmptyOptionalFields(t *testing.T) {
	dto := RAGStatusDTO{
		Status: "completed",
		// CanonicalSummary is nil (empty)
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("failed to marshal RAGStatusDTO: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Empty optional fields should be omitted
	if _, exists := result["canonical_summary"]; exists {
		t.Error("expected canonical_summary to be omitted when nil")
	}
	if _, exists := result["action_items"]; exists {
		t.Error("expected action_items to be omitted when nil/empty")
	}
	if _, exists := result["pdf_url"]; exists {
		t.Error("expected pdf_url to be omitted when empty")
	}
}

func TestRAGSummaryResponseDTO_JSONMarshaling_OmitsEmptyOptionalFields(t *testing.T) {
	dto := RAGSummaryResponseDTO{
		Summary: "Summary text",
		// CanonicalSummary is nil
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("failed to marshal RAGSummaryResponseDTO: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Empty optional fields should be omitted
	if _, exists := result["canonical_summary"]; exists {
		t.Error("expected canonical_summary to be omitted when nil")
	}
}

func TestCanonicalMeetingSummary_JSONSerialization(t *testing.T) {
	summary := CanonicalMeetingSummary{
		Metadata: SummaryMetadata{
			MeetingTitle: "Test",
			Participants: []string{"Alice", "Bob"},
			Language:     "en",
		},
		Agenda: "Agenda text",
	}

	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("failed to marshal CanonicalMeetingSummary: %v", err)
	}

	var result CanonicalMeetingSummary
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal CanonicalMeetingSummary: %v", err)
	}

	if result.Metadata.MeetingTitle != "Test" {
		t.Errorf("expected meeting_title 'Test', got %s", result.Metadata.MeetingTitle)
	}
	if len(result.Metadata.Participants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(result.Metadata.Participants))
	}
	if result.Agenda != "Agenda text" {
		t.Errorf("expected agenda 'Agenda text', got %s", result.Agenda)
	}
}

func TestCanonicalMeetingSummary_OmitsEmptyOptionalFields(t *testing.T) {
	summary := CanonicalMeetingSummary{
		Metadata: SummaryMetadata{
			MeetingTitle: "Test",
			Language:     "en",
		},
		Agenda: "Agenda",
	}

	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Optional empty fields should be omitted
	optionalFields := []string{
		"background_and_objective",
		"main_discussion_sections",
		"roles_and_responsibilities",
		"action_items",
		"decisions_made",
		"open_issues",
		"risks_and_mitigation",
		"additional_notes",
	}

	for _, field := range optionalFields {
		if _, exists := result[field]; exists {
			t.Errorf("expected %s to be omitted when empty", field)
		}
	}
}
