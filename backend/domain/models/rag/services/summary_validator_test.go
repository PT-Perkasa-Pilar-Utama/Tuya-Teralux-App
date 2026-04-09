package services

import (
	"testing"

	"sensio/domain/models/rag/dtos"
)

func TestValidateSummary_ValidSummaryWithAllFields_Passes(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Q2 Planning Review",
			Date:         "2026-04-09",
			Location:     "Meeting Room A",
			Participants: []string{"Alice", "Bob", "Charlie"},
			Context:      "Quarterly planning",
			Style:        "minutes",
			Language:     "en",
		},
		Agenda:                 "Review Q2 objectives and align on priorities",
		BackgroundAndObjective: "This meeting was called to finalize Q2 goals",
		MainDiscussionSections: []dtos.DiscussionSection{
			{
				Title: "Budget Allocation",
				KeyPoints: []dtos.DiscussionPoint{
					{Content: "Budget increased by 10% for engineering", Speaker: "Alice"},
				},
				Decisions:   []string{"Approve 10% budget increase"},
				ActionItems: []string{"Submit revised budget proposal - Bob - 2026-04-15"},
			},
		},
		RolesAndResponsibilities: []dtos.RoleResponsibility{
			{Role: "Project Lead", AssignedTo: "Alice", Description: "Oversee Q2 delivery"},
		},
		ActionItems: []dtos.ActionItem{
			{Task: "Update project roadmap", PIC: "Bob", Deadline: "2026-04-20", Status: "Open"},
		},
		DecisionsMade: []dtos.Decision{
			{Description: "Adopt new CI/CD pipeline", Rationale: "Faster deployment cycles"},
		},
		OpenIssues: []dtos.OpenIssue{
			{Description: "Vendor contract renewal pending", Owner: "Charlie"},
		},
		RisksAndMitigation: []dtos.Risk{
			{Description: "Delayed hiring", Impact: "High", Mitigation: "Use contractors temporarily"},
		},
		AdditionalNotes: "Next review scheduled for May",
	}

	err := ValidateSummary(summary)
	if err != nil {
		t.Errorf("expected valid summary to pass validation, got: %v", err)
	}
}

func TestValidateSummary_EmptyOptionalFields_Passes(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Quick Sync",
			Language:     "en",
		},
		Agenda: "Daily standup update",
	}

	err := ValidateSummary(summary)
	if err != nil {
		t.Errorf("expected summary with empty optional fields to pass, got: %v", err)
	}
}

func TestValidateSummary_MeetingTitlePlaceholder_Fails(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "[Meeting Title]",
			Language:     "en",
		},
		Agenda: "Discussion about project status",
	}

	err := ValidateSummary(summary)
	if err == nil {
		t.Error("expected validation to fail for placeholder in meeting title")
		return
	}
	found := false
	for _, e := range err.Errors {
		if e.FieldPath == "metadata.meeting_title" && e.Rule == "no_placeholder" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected validation error for metadata.meeting_title placeholder, got: %v", err)
	}
}

func TestValidateSummary_NAInActionItemPIC_Fails(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Team Meeting",
			Language:     "en",
		},
		Agenda: "Weekly sync",
		ActionItems: []dtos.ActionItem{
			{Task: "Review PR #123", PIC: "N/A", Deadline: "2026-04-15"},
		},
	}

	err := ValidateSummary(summary)
	if err == nil {
		t.Error("expected validation to fail for N/A in action item PIC")
		return
	}
	found := false
	for _, e := range err.Errors {
		if e.FieldPath == "action_items[0].pic" && e.Rule == "no_placeholder" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected validation error for action_items[0].pic placeholder, got: %v", err)
	}
}

func TestValidateSummary_TBDInDeadline_Fails(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Planning Session",
			Language:     "en",
		},
		Agenda: "Sprint planning",
		ActionItems: []dtos.ActionItem{
			{Task: "Implement feature X", PIC: "Alice", Deadline: "TBD"},
		},
	}

	err := ValidateSummary(summary)
	if err == nil {
		t.Error("expected validation to fail for TBD in deadline")
		return
	}
	found := false
	for _, e := range err.Errors {
		if e.FieldPath == "action_items[0].deadline" && e.Rule == "no_placeholder" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected validation error for action_items[0].deadline placeholder, got: %v", err)
	}
}

func TestValidateSummary_EmptyRequiredMetadata_Fails(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "",
			Language:     "",
		},
		Agenda: "",
	}

	err := ValidateSummary(summary)
	if err == nil {
		t.Error("expected validation to fail for empty required metadata")
		return
	}
	// Should have errors for meeting_title, language, and agenda
	fieldPaths := make(map[string]bool)
	for _, e := range err.Errors {
		fieldPaths[e.FieldPath] = true
	}
	for _, expected := range []string{"metadata.meeting_title", "metadata.language", "agenda"} {
		if !fieldPaths[expected] {
			t.Errorf("expected validation error for %s, got errors for: %v", expected, fieldPaths)
		}
	}
}

func TestValidateSummary_MalformedActionItem_EmptyTask_Fails(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Review",
			Language:     "en",
		},
		Agenda: "Code review",
		ActionItems: []dtos.ActionItem{
			{Task: "", PIC: "Bob", Deadline: "2026-04-15"},
		},
	}

	err := ValidateSummary(summary)
	if err == nil {
		t.Error("expected validation to fail for empty action item task")
		return
	}
	found := false
	for _, e := range err.Errors {
		if e.FieldPath == "action_items[0].task" && e.Rule == "required" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected validation error for action_items[0].task required, got: %v", err)
	}
}

func TestValidateSummary_NilSummary_Fails(t *testing.T) {
	err := ValidateSummary(nil)
	if err == nil {
		t.Error("expected validation to fail for nil summary")
		return
	}
	if len(err.Errors) != 1 {
		t.Errorf("expected 1 error for nil summary, got %d", len(err.Errors))
	}
	if err.Errors[0].FieldPath != "summary" {
		t.Errorf("expected error on 'summary' field, got %s", err.Errors[0].FieldPath)
	}
}

func TestValidateSummary_LowercaseTBD_Fails(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Meeting",
			Language:     "en",
		},
		Agenda: "tbd",
	}

	err := ValidateSummary(summary)
	if err == nil {
		t.Error("expected validation to fail for lowercase tbd in agenda")
	}
}

func TestValidateSummary_BracketPlaceholder_Fails(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Sprint Review",
			Language:     "en",
		},
		Agenda: "[Insert agenda here]",
	}

	err := ValidateSummary(summary)
	if err == nil {
		t.Error("expected validation to fail for bracket placeholder in agenda")
	}
}

func TestValidateSummary_NotAvailablePlaceholder_Fails(t *testing.T) {
	summary := &dtos.CanonicalMeetingSummary{
		Metadata: dtos.SummaryMetadata{
			MeetingTitle: "Meeting",
			Language:     "en",
		},
		Agenda: "Not Available",
	}

	err := ValidateSummary(summary)
	if err == nil {
		t.Error("expected validation to fail for 'Not Available' placeholder")
	}
}

func TestValidationErrors_Error_Message(t *testing.T) {
	ve := &ValidationErrors{
		Errors: []*ValidationError{
			{FieldPath: "agenda", Rule: "required", Message: "agenda is required"},
		},
	}

	msg := ve.Error()
	if msg == "" {
		t.Error("expected non-empty error message from ValidationErrors")
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	veEmpty := &ValidationErrors{}
	if veEmpty.HasErrors() {
		t.Error("expected empty ValidationErrors to report no errors")
	}

	veWithErrors := &ValidationErrors{Errors: []*ValidationError{{}}}
	if !veWithErrors.HasErrors() {
		t.Error("expected ValidationErrors with errors to report has errors")
	}
}

func TestValidationError_Error_String(t *testing.T) {
	e := &ValidationError{
		FieldPath: "metadata.meeting_title",
		Rule:      "required",
		Message:   "must not be empty",
		Value:     "",
	}

	msg := e.Error()
	if msg == "" {
		t.Error("expected non-empty error string from ValidationError")
	}
}
