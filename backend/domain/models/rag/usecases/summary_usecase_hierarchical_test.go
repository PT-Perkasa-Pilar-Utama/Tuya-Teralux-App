package usecases

import (
	"strings"
	"testing"
)

func TestValidateIntermediateNote_ValidNoteWithSummary_Passes(t *testing.T) {
	note := &IntermediateSummaryNote{
		WindowID:  0,
		Summary:   "The team discussed the new feature rollout plan.",
		Decisions: []string{"Approve phase 1"},
	}

	u := &summaryUseCase{}
	err := u.validateIntermediateNote(note)
	if err != nil {
		t.Errorf("expected valid note to pass, got: %v", err)
	}
}

func TestValidateIntermediateNote_ValidNoteWithStructuredFields_Passes(t *testing.T) {
	note := &IntermediateSummaryNote{
		WindowID:    1,
		Summary:     "",
		Decisions:   []string{"Decide on architecture"},
		ActionItems: []string{"Create PRD - Alice - 2026-04-15"},
	}

	u := &summaryUseCase{}
	err := u.validateIntermediateNote(note)
	if err != nil {
		t.Errorf("expected valid note with structured fields to pass, got: %v", err)
	}
}

func TestValidateIntermediateNote_EmptyNote_Fails(t *testing.T) {
	note := &IntermediateSummaryNote{
		WindowID:    2,
		Summary:     "",
		Decisions:   nil,
		ActionItems: nil,
		Risks:       nil,
		OpenQuestions: nil,
	}

	u := &summaryUseCase{}
	err := u.validateIntermediateNote(note)
	if err == nil {
		t.Error("expected empty note to fail validation")
	}
	if !strings.Contains(err.Error(), "no meaningful content") {
		t.Errorf("expected 'no meaningful content' error, got: %v", err)
	}
}

func TestValidateIntermediateNote_PlaceholderSummary_Fails(t *testing.T) {
	note := &IntermediateSummaryNote{
		WindowID: 3,
		Summary:  "[Meeting Title]",
	}

	u := &summaryUseCase{}
	err := u.validateIntermediateNote(note)
	if err == nil {
		t.Error("expected placeholder summary to fail validation")
	}
	if !strings.Contains(err.Error(), "placeholder") {
		t.Errorf("expected 'placeholder' error, got: %v", err)
	}
}

func TestValidateIntermediateNote_NAPlaceholder_Fails(t *testing.T) {
	note := &IntermediateSummaryNote{
		WindowID: 4,
		Summary:  "N/A",
	}

	u := &summaryUseCase{}
	err := u.validateIntermediateNote(note)
	if err == nil {
		t.Error("expected N/A summary to fail validation")
	}
}

func TestValidateIntermediateNote_EmptyActionItem_Fails(t *testing.T) {
	note := &IntermediateSummaryNote{
		WindowID:    5,
		Summary:     "Some summary content",
		ActionItems: []string{"", "Valid action item"},
	}

	u := &summaryUseCase{}
	err := u.validateIntermediateNote(note)
	if err == nil {
		t.Error("expected empty action item to fail validation")
	}
	if !strings.Contains(err.Error(), "empty action item") {
		t.Errorf("expected 'empty action item' error, got: %v", err)
	}
}

func TestValidateIntermediateNote_EmptyDecision_Fails(t *testing.T) {
	note := &IntermediateSummaryNote{
		WindowID:  6,
		Summary:   "Some summary content",
		Decisions: []string{"", "Valid decision"},
	}

	u := &summaryUseCase{}
	err := u.validateIntermediateNote(note)
	if err == nil {
		t.Error("expected empty decision to fail validation")
	}
	if !strings.Contains(err.Error(), "empty decision") {
		t.Errorf("expected 'empty decision' error, got: %v", err)
	}
}

func TestValidateIntermediateNote_ValidNoteWithRisks_Passes(t *testing.T) {
	note := &IntermediateSummaryNote{
		WindowID: 7,
		Summary:  "",
		Risks:    []string{"Server downtime risk"},
	}

	u := &summaryUseCase{}
	err := u.validateIntermediateNote(note)
	if err != nil {
		t.Errorf("expected note with risks to pass, got: %v", err)
	}
}

func TestValidateIntermediateNote_ValidNoteWithOpenQuestions_Passes(t *testing.T) {
	note := &IntermediateSummaryNote{
		WindowID:      8,
		Summary:       "",
		OpenQuestions: []string{"Who owns the budget?"},
	}

	u := &summaryUseCase{}
	err := u.validateIntermediateNote(note)
	if err != nil {
		t.Errorf("expected note with open questions to pass, got: %v", err)
	}
}

func TestIsPlaceholderText_DetectsPlaceholders(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"N/A", true},
		{"TBD", true},
		{"Not Available", true},
		{"[Meeting Title]", true},
		{"[Insert content here]", true},
		{"[TODO]", true},
		{"[Date]", true},
		{"Valid meeting summary about Q2 planning", false},
		{"The team decided to use microservices", false},
		{"Budget increased by 10%", false},
	}

	for _, tt := range tests {
		result := isPlaceholderText(tt.input)
		if result != tt.expected {
			t.Errorf("isPlaceholderText(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}
