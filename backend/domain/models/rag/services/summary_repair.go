package services

import (
	"regexp"
	"sensio/domain/models/rag/dtos"
	"strings"
)

// SummaryRepairer attempts to fix validation errors in a CanonicalMeetingSummary.
// It applies safe, deterministic transformations — NOT LLM re-calling.
// Repairs include: stripping placeholder text, inferring missing metadata, normalizing fields.
type SummaryRepairer struct{}

// NewSummaryRepairer creates a new repairer instance.
func NewSummaryRepairer() *SummaryRepairer {
	return &SummaryRepairer{}
}

// RepairSummary attempts to fix validation errors in the canonical summary.
// Returns the repaired summary and a list of applied repairs.
// If no repairs were needed, returns the original summary with nil repairs.
func (r *SummaryRepairer) RepairSummary(summary *dtos.CanonicalMeetingSummary, errors *ValidationErrors) (*dtos.CanonicalMeetingSummary, []string) {
	if summary == nil || errors == nil || !errors.HasErrors() {
		return summary, nil
	}

	var repairs []string
	repaired := *summary // shallow copy
	repaired.Metadata = summary.Metadata
	repaired.MainDiscussionSections = make([]dtos.DiscussionSection, len(summary.MainDiscussionSections))
	copy(repaired.MainDiscussionSections, summary.MainDiscussionSections)
	repaired.ActionItems = make([]dtos.ActionItem, len(summary.ActionItems))
	copy(repaired.ActionItems, summary.ActionItems)
	repaired.DecisionsMade = make([]dtos.Decision, len(summary.DecisionsMade))
	copy(repaired.DecisionsMade, summary.DecisionsMade)
	repaired.OpenIssues = make([]dtos.OpenIssue, len(summary.OpenIssues))
	copy(repaired.OpenIssues, summary.OpenIssues)
	repaired.RisksAndMitigation = make([]dtos.Risk, len(summary.RisksAndMitigation))
	copy(repaired.RisksAndMitigation, summary.RisksAndMitigation)
	repaired.RolesAndResponsibilities = make([]dtos.RoleResponsibility, len(summary.RolesAndResponsibilities))
	copy(repaired.RolesAndResponsibilities, summary.RolesAndResponsibilities)

	for _, validationErr := range errors.Errors {
		repair := r.applyRepair(&repaired, validationErr)
		if repair != "" {
			repairs = append(repairs, repair)
		}
	}

	return &repaired, repairs
}

// applyRepair attempts to fix a specific validation error.
// Returns a description of the repair applied, or "" if no repair was possible.
func (r *SummaryRepairer) applyRepair(summary *dtos.CanonicalMeetingSummary, err *ValidationError) string {
	fieldPath := err.FieldPath

	switch {
	// Repair placeholder in metadata.meeting_title
	case fieldPath == "metadata.meeting_title" && err.Rule == "no_placeholder":
		if summary.Metadata.MeetingTitle == "" || isPlaceholderOnly(summary.Metadata.MeetingTitle) {
			// Try to infer from agenda
			if summary.Agenda != "" && !isPlaceholderOnly(summary.Agenda) {
				summary.Metadata.MeetingTitle = extractTopicFromText(summary.Agenda)
				return "Inferred meeting_title from agenda content"
			}
			// Try to infer from context
			if summary.Metadata.Context != "" && !isPlaceholderOnly(summary.Metadata.Context) {
				summary.Metadata.MeetingTitle = extractTopicFromText(summary.Metadata.Context)
				return "Inferred meeting_title from context"
			}
		}

	// Repair placeholder in agenda
	case fieldPath == "agenda" && err.Rule == "no_placeholder":
		if isPlaceholderOnly(summary.Agenda) {
			// Try to infer from background
			if summary.BackgroundAndObjective != "" && !isPlaceholderOnly(summary.BackgroundAndObjective) {
				summary.Agenda = extractTopicFromText(summary.BackgroundAndObjective)
				return "Inferred agenda from background/objective content"
			}
		}

	// Repair placeholder in metadata.language
	case fieldPath == "metadata.language" && err.Rule == "no_placeholder":
		if isPlaceholderOnly(summary.Metadata.Language) {
			// Default to English if unknown
			summary.Metadata.Language = "en"
			return "Set default language 'en'"
		}

	// Strip placeholders from text fields
	case err.Rule == "no_placeholder":
		stripped := stripPlaceholdersFromField(summary, fieldPath)
		if stripped {
			return "Stripped placeholder text from " + fieldPath
		}

	// Repair empty required metadata field
	case fieldPath == "metadata.meeting_title" && err.Rule == "required":
		if summary.Agenda != "" {
			summary.Metadata.MeetingTitle = extractTopicFromText(summary.Agenda)
			return "Inferred meeting_title from agenda"
		}
		if summary.Metadata.Context != "" {
			summary.Metadata.MeetingTitle = extractTopicFromText(summary.Metadata.Context)
			return "Inferred meeting_title from context"
		}

	case fieldPath == "metadata.language" && err.Rule == "required":
		summary.Metadata.Language = "en"
		return "Set default language 'en'"

	case fieldPath == "agenda" && err.Rule == "required":
		if summary.BackgroundAndObjective != "" {
			summary.Agenda = extractTopicFromText(summary.BackgroundAndObjective)
			return "Inferred agenda from background/objective"
		}
	}

	return "" // No repair possible
}

// stripPlaceholdersFromField removes placeholder text from the specified field.
// Returns true if a repair was applied.
func stripPlaceholdersFromField(summary *dtos.CanonicalMeetingSummary, fieldPath string) bool {
	value := getField(summary, fieldPath)
	if value == "" {
		return false
	}
	cleaned := stripAllPlaceholders(value)
	if cleaned != value && cleaned != "" {
		setField(summary, fieldPath, cleaned)
		return true
	}
	return false
}

// getField gets a text field value from the summary by field path.
func getField(summary *dtos.CanonicalMeetingSummary, fieldPath string) string {
	switch fieldPath {
	case "metadata.meeting_title":
		return summary.Metadata.MeetingTitle
	case "metadata.language":
		return summary.Metadata.Language
	case "agenda":
		return summary.Agenda
	case "background_and_objective":
		return summary.BackgroundAndObjective
	case "additional_notes":
		return summary.AdditionalNotes
	case "metadata.date":
		return summary.Metadata.Date
	case "metadata.location":
		return summary.Metadata.Location
	case "metadata.context":
		return summary.Metadata.Context
	case "metadata.style":
		return summary.Metadata.Style
	}
	return ""
}

// setField sets a text field value in the summary by field path.
func setField(summary *dtos.CanonicalMeetingSummary, fieldPath string, value string) {
	switch fieldPath {
	case "metadata.meeting_title":
		summary.Metadata.MeetingTitle = value
	case "metadata.language":
		summary.Metadata.Language = value
	case "agenda":
		summary.Agenda = value
	case "background_and_objective":
		summary.BackgroundAndObjective = value
	case "additional_notes":
		summary.AdditionalNotes = value
	case "metadata.date":
		summary.Metadata.Date = value
	case "metadata.location":
		summary.Metadata.Location = value
	case "metadata.context":
		summary.Metadata.Context = value
	case "metadata.style":
		summary.Metadata.Style = value
	}
}

// isPlaceholderOnly checks if a string consists entirely of placeholder text.
func isPlaceholderOnly(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	return containsPlaceholder(s)
}

// containsPlaceholder checks if a string contains any placeholder patterns.
func containsPlaceholder(s string) bool {
	placeholderPatterns := []string{
		"N/A", "n/a", "TBD", "tbd", "Not Available",
		"[Insert", "[Add", "[TODO", "[TBD", "[Placeholder",
		"[Meeting Title]", "[Date]", "[Location]",
	}
	for _, p := range placeholderPatterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return true
	}
	return false
}

// stripAllPlaceholders removes all placeholder patterns from text.
// Returns cleaned text with placeholders replaced by empty string.
func stripAllPlaceholders(text string) string {
	patterns := []string{
		`\[.*?\]`,
		`\bN/A\b`, `\bn/a\b`,
		`\bTBD\b`, `\btbd\b`,
		`\bNot Available\b`,
	}

	result := text
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "")
	}

	// Clean up extra spaces left by removals
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}

	return strings.TrimSpace(result)
}

// extractTopicFromText extracts a short topic/title from a longer text string.
// Takes the first sentence or first 80 characters.
func extractTopicFromText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// Take first sentence
	if idx := strings.Index(text, "."); idx > 0 && idx < 100 {
		return strings.TrimSpace(text[:idx+1])
	}

	// Take first 80 characters
	if len(text) > 80 {
		return strings.TrimSpace(text[:80]) + "..."
	}

	return text
}
