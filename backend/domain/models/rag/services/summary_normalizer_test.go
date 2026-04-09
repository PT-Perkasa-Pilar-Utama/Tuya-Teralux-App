package services

import (
	"strings"
	"testing"

	"sensio/domain/models/rag/dtos"
)

// ParseMarkdownToCanonical is a helper function for testing.
// In production, this would be implemented in summary_normalizer.go
// For now, we test the concept through the validator + markdown generator roundtrip.
func TestParseLLMMarkdownOutput_PopulatesFields(t *testing.T) {
	// Simulate what a normalized LLM output would look like
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Sprint Planning",
			Participants: []string{"Alice", "Bob"},
			Language:     "en",
		},
		Agenda: "Plan sprint 24",
		MainDiscussionSections: []dtos.DiscussionSection{
			{
				Title: "Feature Priorities",
				KeyPoints: []dtos.DiscussionPoint{
					{Content: "Focus on auth module first", Speaker: "Alice"},
				},
			},
		},
		ActionItems: []dtos.ActionItem{
			{Task: "Create auth module PRD", PIC: "Bob", Deadline: "2026-04-12"},
		},
	}

	// Validate the normalized output
	err := ValidateSummary(summary)
	if err != nil {
		t.Errorf("expected normalized summary to pass validation, got: %v", err)
	}

	// Generate markdown from it
	md := GenerateMarkdown(summary)
	if !strings.Contains(md, "Sprint Planning") {
		t.Error("expected markdown to contain meeting title")
	}
	if !strings.Contains(md, "Feature Priorities") {
		t.Error("expected markdown to contain discussion section title")
	}
}

func TestParseLLMMarkdownOutput_MissingSections(t *testing.T) {
	// Simulate LLM output with missing sections (only agenda and metadata)
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Quick Sync",
			Language:     "en",
		},
		Agenda: "Quick status update",
	}

	err := ValidateSummary(summary)
	if err != nil {
		t.Errorf("expected summary with missing optional sections to pass validation, got: %v", err)
	}

	md := GenerateMarkdown(summary)

	// Verify empty sections are omitted
	omittedSections := []string{
		"## Action Items",
		"## Decisions Made",
		"## Open Issues",
		"## Risks & Mitigation",
	}
	for _, section := range omittedSections {
		if strings.Contains(md, section) {
			t.Errorf("expected %q to be omitted from markdown", section)
		}
	}
}

func TestParseLLMMarkdownOutput_PlaceholderText_FlagsValidation(t *testing.T) {
	// Simulate LLM output that contains placeholders
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "[Meeting Title]",
			Language:     "en",
		},
		Agenda: "N/A",
	}

	err := ValidateSummary(summary)
	if err == nil {
		t.Error("expected validation to fail for placeholder text in normalized output")
		return
	}

	// Should flag both meeting_title and agenda
	if len(err.Errors) < 2 {
		t.Errorf("expected at least 2 validation errors, got %d", len(err.Errors))
	}
}
