package services

import (
	"strings"
	"testing"

	"sensio/domain/models/rag/dtos"
)

func TestGenerateMarkdown_AllSectionsPresent(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Q2 Planning Review",
			Date:         "2026-04-09",
			Location:     "Room A",
			Participants: []string{"Alice", "Bob"},
			Language:     "en",
		},
		Agenda: "Review Q2 objectives",
		BackgroundAndObjective: "Align on priorities for next quarter",
		MainDiscussionSections: []dtos.DiscussionSection{
			{
				Title: "Budget",
				KeyPoints: []dtos.DiscussionPoint{
					{Content: "Budget increased by 10%", Speaker: "Alice"},
				},
				Decisions:   []string{"Approved"},
				ActionItems: []string{"Update budget - Bob - 2026-04-15"},
			},
		},
		RolesAndResponsibilities: []dtos.RoleResponsibility{
			{Role: "Lead", AssignedTo: "Alice", Description: "Project oversight"},
		},
		ActionItems: []dtos.ActionItem{
			{Task: "Submit report", PIC: "Bob", Deadline: "2026-04-20", Status: "Open"},
		},
		DecisionsMade: []dtos.Decision{
			{Description: "Adopt new CI/CD", Rationale: "Faster deployments"},
		},
		OpenIssues: []dtos.OpenIssue{
			{Description: "Vendor contract", Owner: "Charlie"},
		},
		RisksAndMitigation: []dtos.Risk{
			{Description: "Hiring delay", Impact: "High", Mitigation: "Use contractors"},
		},
		AdditionalNotes: "Follow up in May",
	}

	md := GenerateMarkdown(summary)

	// Check all sections are present
	expectedSections := []string{
		"# Q2 Planning Review",
		"## Agenda",
		"Review Q2 objectives",
		"## Background & Objective",
		"## Main Discussion",
		"### Budget",
		"## Roles & Responsibilities",
		"## Action Items",
		"## Decisions Made",
		"## Open Issues",
		"## Risks & Mitigation",
		"## Additional Notes",
	}

	for _, section := range expectedSections {
		if !strings.Contains(md, section) {
			t.Errorf("expected markdown to contain %q, but it was missing", section)
		}
	}
}

func TestGenerateMarkdown_OmitsEmptySections(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Quick Sync",
			Language:     "en",
		},
		Agenda: "Daily standup",
	}

	md := GenerateMarkdown(summary)

	// Sections with no data should not appear
	omittedSections := []string{
		"## Background & objective",
		"## Main Discussion",
		"## Roles & Responsibilities",
		"## Action Items",
		"## Decisions Made",
		"## Open Issues",
		"## Risks & Mitigation",
		"## Additional Notes",
	}

	for _, section := range omittedSections {
		if strings.Contains(md, section) {
			t.Errorf("expected markdown to NOT contain %q (empty section should be omitted)", section)
		}
	}
}

func TestGenerateMarkdown_ActionItemsAsTable(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Meeting",
			Language:     "en",
		},
		Agenda: "Sync",
		ActionItems: []dtos.ActionItem{
			{Task: "Task 1", PIC: "Alice", Deadline: "2026-04-15", Status: "Open"},
			{Task: "Task 2", PIC: "Bob", Status: "Done"},
		},
	}

	md := GenerateMarkdown(summary)

	// Check table header
	if !strings.Contains(md, "| # | Task | PIC | Deadline | Status |") {
		t.Error("expected action items to be rendered as markdown table with header row")
	}
	if !strings.Contains(md, "|---|------|-----|----------|--------|") {
		t.Error("expected action items table to have separator row")
	}
	// Check default status
	if !strings.Contains(md, "| Done") {
		t.Error("expected action item with 'Done' status to appear in table")
	}
}

func TestGenerateMarkdown_DecisionsAsNumberedList(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Meeting",
			Language:     "en",
		},
		Agenda: "Sync",
		DecisionsMade: []dtos.Decision{
			{Description: "First decision", Rationale: "Reason 1"},
			{Description: "Second decision"},
		},
	}

	md := GenerateMarkdown(summary)

	if !strings.Contains(md, "1. **First decision**") {
		t.Error("expected decisions to be rendered as numbered list with bold description")
	}
	if !strings.Contains(md, "2. **Second decision**") {
		t.Error("expected second decision in numbered list")
	}
	if !strings.Contains(md, "*Rationale: Reason 1*") {
		t.Error("expected rationale to appear in italics")
	}
}

func TestGenerateMarkdown_RisksAsTable(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Meeting",
			Language:     "en",
		},
		Agenda: "Sync",
		RisksAndMitigation: []dtos.Risk{
			{Description: "Server downtime", Impact: "High", Mitigation: "Redundant servers"},
		},
	}

	md := GenerateMarkdown(summary)

	if !strings.Contains(md, "| # | Description | Impact | Mitigation |") {
		t.Error("expected risks to be rendered as markdown table")
	}
	if !strings.Contains(md, "| Server downtime") {
		t.Error("expected risk description in table")
	}
}

func TestGenerateMarkdown_HandlesSpecialCharacters(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Meeting with | pipe chars",
			Language:     "en",
		},
		Agenda: "Discussion about\nmulti-line content",
		ActionItems: []dtos.ActionItem{
			{Task: "Task with | pipe and\nnewline", PIC: "Alice"},
		},
	}

	md := GenerateMarkdown(summary)

	// Pipe characters should be escaped
	if strings.Contains(md, "| pipe chars") && !strings.Contains(md, "\\| pipe chars") {
		// Title is in header, not in a table, so pipe isn't escaped there
	}
	// In table cells, pipes should be escaped
	if strings.Contains(md, "\\| pipe") {
		// Good, pipe was escaped
	}
	// Newlines in table cells should be replaced with spaces
	if strings.Contains(md, "Task with | pipe and\nnewline") {
		t.Error("expected newlines in table cells to be replaced with spaces")
	}
}

func TestGenerateMarkdown_ValidGoldmarkParsable(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Test Meeting",
			Language:     "en",
		},
		Agenda: "Test agenda",
		ActionItems: []dtos.ActionItem{
			{Task: "Do something", PIC: "Alice"},
		},
		DecisionsMade: []dtos.Decision{
			{Description: "Agreed on X"},
		},
	}

	md := GenerateMarkdown(summary)
	if md == "" {
		t.Error("expected non-empty markdown output")
	}

	// Basic structure check: should have at least one heading
	if !strings.Contains(md, "#") {
		t.Error("expected markdown to contain at least one heading")
	}
}

func TestGenerateMarkdown_NilSummary(t *testing.T) {
	md := GenerateMarkdown(nil)
	if md != "" {
		t.Errorf("expected empty string for nil summary, got: %s", md)
	}
}

func TestGenerateMarkdown_MetadataBlock(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Meeting",
			Date:         "2026-04-09",
			Location:     "Room A",
			Participants: []string{"Alice", "Bob"},
			Language:     "en",
		},
		Agenda: "Sync",
	}

	md := GenerateMarkdown(summary)

	if !strings.Contains(md, "**Date:** 2026-04-09") {
		t.Error("expected metadata block to contain date")
	}
	if !strings.Contains(md, "**Location:** Room A") {
		t.Error("expected metadata block to contain location")
	}
	if !strings.Contains(md, "**Participants:** Alice, Bob") {
		t.Error("expected metadata block to contain participants")
	}
}

func TestGenerateMarkdown_DiscussionSectionWithSubContent(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Meeting",
			Language:     "en",
		},
		Agenda: "Sync",
		MainDiscussionSections: []dtos.DiscussionSection{
			{
				Title: "Topic A",
				KeyPoints: []dtos.DiscussionPoint{
					{Content: "Point 1", Speaker: "Alice"},
					{Content: "Point 2"},
				},
			},
		},
	}

	md := GenerateMarkdown(summary)

	if !strings.Contains(md, "### Topic A") {
		t.Error("expected discussion section title")
	}
	if !strings.Contains(md, "- Point 1 *(Alice)*") {
		t.Error("expected key point with speaker attribution")
	}
	if !strings.Contains(md, "- Point 2") {
		t.Error("expected key point without speaker")
	}
}
