package services

import (
	"strings"
	"testing"

	"sensio/domain/models/rag/dtos"
)

func TestRepairSummary_RepairableWithPlaceholders_Succeeds(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "[Meeting Title]",
			Language:     "en",
		},
		Agenda: "Review Q2 objectives and align priorities.",
	}

	validationErr := ValidateSummary(summary)
	if validationErr == nil {
		t.Fatal("expected validation to fail for placeholder")
	}

	repairer := NewSummaryRepairer()
	repaired, repairs := repairer.RepairSummary(summary, validationErr)

	// Should infer meeting title from agenda
	if len(repairs) == 0 {
		t.Fatal("expected repairs to be applied")
	}

	// Verify meeting title was inferred
	if repaired.Metadata.MeetingTitle == "[Meeting Title]" {
		t.Error("expected meeting title to be repaired")
	}
}

func TestRepairSummary_AlreadyValid_NoRepairs(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Q2 Planning Review",
			Language:     "en",
		},
		Agenda: "Review Q2 objectives",
	}

	validationErr := ValidateSummary(summary)
	if validationErr != nil {
		t.Fatalf("expected valid summary to pass, got: %v", validationErr)
	}

	repairer := NewSummaryRepairer()
	repaired, repairs := repairer.RepairSummary(summary, validationErr)

	if len(repairs) != 0 {
		t.Errorf("expected no repairs for valid summary, got %d: %v", len(repairs), repairs)
	}
	if repaired != summary {
		t.Error("expected original summary returned when no repairs needed")
	}
}

func TestRepairSummary_NilSummary_NoRepairs(t *testing.T) {
	repairer := NewSummaryRepairer()
	repaired, repairs := repairer.RepairSummary(nil, nil)

	if repaired != nil {
		t.Error("expected nil summary returned")
	}
	if len(repairs) != 0 {
		t.Errorf("expected no repairs, got %d", len(repairs))
	}
}

func TestRepairSummary_InferTitleFromAgenda(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "[Meeting Title]",
			Language:     "en",
		},
		Agenda: "Review Q2 objectives and align on priorities for the engineering team.",
	}

	validationErr := ValidateSummary(summary)
	if validationErr == nil {
		t.Fatal("expected validation to fail for placeholder")
	}

	repairer := NewSummaryRepairer()
	repaired, repairs := repairer.RepairSummary(summary, validationErr)

	if len(repairs) == 0 {
		t.Fatal("expected repairs to be applied")
	}

	// Check if meeting title was inferred
	foundInference := false
	for _, r := range repairs {
		if strings.Contains(r, "Inferred meeting_title from agenda") {
			foundInference = true
			break
		}
	}
	if !foundInference {
		t.Errorf("expected meeting title inference repair, got: %v", repairs)
	}

	// Verify meeting title was updated
	if repaired.Metadata.MeetingTitle == "[Meeting Title]" {
		t.Error("expected meeting title to be repaired from agenda")
	}
}

func TestRepairSummary_InferAgendaFromBackground(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Team Sync",
			Language:     "en",
		},
		Agenda:                 "N/A",
		BackgroundAndObjective: "This meeting was called to finalize Q2 goals and allocate resources.",
	}

	validationErr := ValidateSummary(summary)
	if validationErr == nil {
		t.Fatal("expected validation to fail for placeholder")
	}

	repairer := NewSummaryRepairer()
	repaired, repairs := repairer.RepairSummary(summary, validationErr)

	if len(repairs) == 0 {
		t.Fatal("expected repairs to be applied")
	}

	foundAgendaInference := false
	for _, r := range repairs {
		if strings.Contains(r, "Inferred agenda") {
			foundAgendaInference = true
			break
		}
	}
	if !foundAgendaInference {
		t.Errorf("expected agenda inference repair, got: %v", repairs)
	}

	if repaired.Agenda == "N/A" {
		t.Error("expected agenda to be repaired from background")
	}
}

func TestRepairSummary_SetDefaultLanguage(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Team Sync",
			Language:     "",
		},
		Agenda: "Weekly sync",
	}

	validationErr := ValidateSummary(summary)
	if validationErr == nil {
		t.Fatal("expected validation to fail for empty language")
	}

	repairer := NewSummaryRepairer()
	repaired, repairs := repairer.RepairSummary(summary, validationErr)

	if len(repairs) == 0 {
		t.Fatal("expected repairs to be applied")
	}

	foundLangRepair := false
	for _, r := range repairs {
		if strings.Contains(r, "Set default language") {
			foundLangRepair = true
			break
		}
	}
	if !foundLangRepair {
		t.Errorf("expected language repair, got: %v", repairs)
	}

	if repaired.Metadata.Language != "en" {
		t.Errorf("expected language to be 'en', got %q", repaired.Metadata.Language)
	}
}

func TestRepairSummary_Unrepairable_MissingRequiredData(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "",
			Language:     "",
		},
		Agenda: "",
	}

	validationErr := ValidateSummary(summary)
	if validationErr == nil {
		t.Fatal("expected validation to fail")
	}

	repairer := NewSummaryRepairer()
	repaired, repairs := repairer.RepairSummary(summary, validationErr)

	// Should attempt repairs but may not fix everything
	_ = repaired
	_ = repairs

	// Language should be repaired to "en"
	if repaired.Metadata.Language != "en" {
		t.Errorf("expected language to be repaired to 'en', got %q", repaired.Metadata.Language)
	}

	// Meeting title and agenda may not be repairable (no source data)
	// That's expected behavior
}

func TestStripAllPlaceholders_RemovesPlaceholders(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[Meeting Title]", ""},
		{"N/A", ""},
		{"TBD", ""},
		{"Review N/A items", "Review items"},
		{"Plan TBD next week", "Plan next week"},
		{"Meeting about [Insert topic]", "Meeting about"},
	}

	for _, tt := range tests {
		result := stripAllPlaceholders(tt.input)
		if result != tt.expected {
			t.Errorf("stripAllPlaceholders(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractTopicFromText_ExtractsFirstSentence(t *testing.T) {
	result := extractTopicFromText("This is the first sentence. This is the second.")
	if result != "This is the first sentence." {
		t.Errorf("expected first sentence, got %q", result)
	}
}

func TestExtractTopicFromText_TruncatesLongText(t *testing.T) {
	longText := strings.Repeat("a", 200)
	result := extractTopicFromText(longText)
	if len(result) > 83 { // 80 + "..."
		t.Errorf("expected truncated result (max 83 chars), got %d chars", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("expected truncation to end with '...'")
	}
}

func TestContainsPlaceholder_DetectsPlaceholders(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"[Meeting Title]", true},
		{"N/A", true},
		{"TBD", true},
		{"Not Available", true},
		{"[Insert content]", true},
		{"Valid meeting notes", false},
		{"Budget increased by 10%", false},
	}

	for _, tt := range tests {
		result := containsPlaceholder(tt.input)
		if result != tt.expected {
			t.Errorf("containsPlaceholder(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}
