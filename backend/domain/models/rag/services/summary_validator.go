package services

import (
	"fmt"
	"regexp"
	"strings"

	"sensio/domain/models/rag/dtos"
)

// DefaultPlaceholderPatterns contains the regex patterns used to detect placeholder text
// in summary fields. These patterns are checked against all string fields during validation.
var DefaultPlaceholderPatterns = []string{
	`\[.*\]`,               // Bracket placeholders like [Meeting Title], [Insert ...]
	`\bN/A\b`,              // N/A (case-sensitive)
	`\bn/a\b`,              // n/a (lowercase)
	`\bTBD\b`,              // TBD (uppercase)
	`\btbd\b`,              // tbd (lowercase)
	`\bNot Available\b`,    // "Not Available"
	`\[Insert .*\]`,        // [Insert ...]
	`\[Add .*\]`,           // [Add ...]
	`\[TODO.*\]`,           // [TODO ...]
	`\[TBD.*\]`,            // [TBD ...]
	`\[Placeholder.*\]`,    // [Placeholder ...]
}

// ValidationError represents a single validation failure with field path and details.
type ValidationError struct {
	FieldPath     string `json:"field_path"`     // e.g., "metadata.meeting_title", "action_items[0].task"
	Value         string `json:"value"`          // The actual value that failed validation
	Rule          string `json:"rule"`           // e.g., "required", "no_placeholder", "non_empty"
	Message       string `json:"message"`        // Human-readable error description
	SuggestedFix  string `json:"suggested_fix"`  // Guidance on how to fix the issue
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %s: %s (value: %q)", e.FieldPath, e.Message, e.Value)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors struct {
	Errors []*ValidationError `json:"errors"`
}

func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "no validation errors"
	}
	parts := make([]string, len(ve.Errors))
	for i, e := range ve.Errors {
		parts[i] = e.Error()
	}
	return strings.Join(parts, "; ")
}

// placeholderRegex is a compiled regex combining all placeholder patterns.
var placeholderRegex *regexp.Regexp

func init() {
	pattern := strings.Join(DefaultPlaceholderPatterns, "|")
	placeholderRegex = regexp.MustCompile(pattern)
}

// ValidateSummary validates a CanonicalMeetingSummary against the contract rules.
// Returns nil if validation passes, or *ValidationErrors if it fails.
func ValidateSummary(summary *dtos.CanonicalMeetingSummary) *ValidationErrors {
	if summary == nil {
		return &ValidationErrors{
			Errors: []*ValidationError{{
				FieldPath:    "summary",
				Value:        "",
				Rule:         "required",
				Message:      "summary cannot be nil",
				SuggestedFix: "Provide a valid CanonicalMeetingSummary struct",
			}},
		}
	}

	var ve ValidationErrors

	// Validate metadata
	validateMetadata(&summary.Metadata, &ve)

	// Validate agenda (required field)
	if strings.TrimSpace(summary.Agenda) == "" {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    "agenda",
			Value:        summary.Agenda,
			Rule:         "required",
			Message:      "agenda is required and must be non-empty",
			SuggestedFix: "Provide a meaningful agenda description from the meeting transcript",
		})
	} else {
		checkPlaceholder("agenda", summary.Agenda, &ve)
	}

	// Validate background and objective if present
	if strings.TrimSpace(summary.BackgroundAndObjective) != "" {
		checkPlaceholder("background_and_objective", summary.BackgroundAndObjective, &ve)
	}

	// Validate main discussion sections
	for i, section := range summary.MainDiscussionSections {
		validateDiscussionSection(section, i, &ve)
	}

	// Validate roles and responsibilities
	for i, role := range summary.RolesAndResponsibilities {
		validateRoleResponsibility(role, i, &ve)
	}

	// Validate action items
	for i, item := range summary.ActionItems {
		validateActionItem(item, i, &ve)
	}

	// Validate decisions
	for i, decision := range summary.DecisionsMade {
		validateDecision(decision, i, &ve)
	}

	// Validate open issues
	for i, issue := range summary.OpenIssues {
		validateOpenIssue(issue, i, &ve)
	}

	// Validate risks
	for i, risk := range summary.RisksAndMitigation {
		validateRisk(risk, i, &ve)
	}

	// Validate additional notes if present
	if strings.TrimSpace(summary.AdditionalNotes) != "" {
		checkPlaceholder("additional_notes", summary.AdditionalNotes, &ve)
	}

	if len(ve.Errors) > 0 {
		return &ve
	}
	return nil
}

// validateMetadata checks the SummaryMetadata struct for required fields and placeholders.
func validateMetadata(meta *dtos.SummaryMetadata, ve *ValidationErrors) {
	if strings.TrimSpace(meta.MeetingTitle) == "" {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    "metadata.meeting_title",
			Value:        meta.MeetingTitle,
			Rule:         "required",
			Message:      "meeting_title is required and must be non-empty",
			SuggestedFix: "Extract the meeting title from the transcript or meeting context",
		})
	} else {
		checkPlaceholder("metadata.meeting_title", meta.MeetingTitle, ve)
	}

	if strings.TrimSpace(meta.Language) == "" {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    "metadata.language",
			Value:        meta.Language,
			Rule:         "required",
			Message:      "language is required and must be non-empty",
			SuggestedFix: "Set the output language (e.g., 'id' for Indonesian, 'en' for English)",
		})
	}

	// Optional fields: check for placeholders if present
	if strings.TrimSpace(meta.Date) != "" {
		checkPlaceholder("metadata.date", meta.Date, ve)
	}
	if strings.TrimSpace(meta.Location) != "" {
		checkPlaceholder("metadata.location", meta.Location, ve)
	}
	if strings.TrimSpace(meta.Context) != "" {
		checkPlaceholder("metadata.context", meta.Context, ve)
	}
	if strings.TrimSpace(meta.Style) != "" {
		checkPlaceholder("metadata.style", meta.Style, ve)
	}

	// Check participants for placeholders
	for i, p := range meta.Participants {
		if strings.TrimSpace(p) == "" {
			ve.Errors = append(ve.Errors, &ValidationError{
				FieldPath:    fmt.Sprintf("metadata.participants[%d]", i),
				Value:        p,
				Rule:         "non_empty",
				Message:      "participant name must not be empty",
				SuggestedFix: "Remove empty participant entries or provide actual participant names",
			})
		} else {
			checkPlaceholder(fmt.Sprintf("metadata.participants[%d]", i), p, ve)
		}
	}
}

// validateDiscussionSection validates a DiscussionSection struct.
func validateDiscussionSection(section dtos.DiscussionSection, index int, ve *ValidationErrors) {
	prefix := fmt.Sprintf("main_discussion_sections[%d]", index)

	if strings.TrimSpace(section.Title) == "" {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    fmt.Sprintf("%s.title", prefix),
			Value:        section.Title,
			Rule:         "required",
			Message:      "discussion section title must not be empty",
			SuggestedFix: "Provide a meaningful title for this discussion section",
		})
	} else {
		checkPlaceholder(fmt.Sprintf("%s.title", prefix), section.Title, ve)
	}

	for j, point := range section.KeyPoints {
		validateDiscussionPoint(point, j, prefix, ve)
	}
}

// validateDiscussionPoint validates a DiscussionPoint struct.
func validateDiscussionPoint(point dtos.DiscussionPoint, index int, parentPrefix string, ve *ValidationErrors) {
	prefix := fmt.Sprintf("%s.key_points[%d]", parentPrefix, index)

	if strings.TrimSpace(point.Content) == "" {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    fmt.Sprintf("%s.content", prefix),
			Value:        point.Content,
			Rule:         "required",
			Message:      "discussion point content must not be empty",
			SuggestedFix: "Provide meaningful content for this discussion point or remove it",
		})
	} else {
		checkPlaceholder(fmt.Sprintf("%s.content", prefix), point.Content, ve)
	}
}

// validateRoleResponsibility validates a RoleResponsibility struct.
func validateRoleResponsibility(role dtos.RoleResponsibility, index int, ve *ValidationErrors) {
	prefix := fmt.Sprintf("roles_and_responsibilities[%d]", index)

	if strings.TrimSpace(role.Role) == "" {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    fmt.Sprintf("%s.role", prefix),
			Value:        role.Role,
			Rule:         "required",
			Message:      "role must not be empty",
			SuggestedFix: "Specify the role or responsibility being assigned",
		})
	} else {
		checkPlaceholder(fmt.Sprintf("%s.role", prefix), role.Role, ve)
	}

	if strings.TrimSpace(role.AssignedTo) != "" {
		checkPlaceholder(fmt.Sprintf("%s.assigned_to", prefix), role.AssignedTo, ve)
	}
}

// validateActionItem validates an ActionItem struct.
func validateActionItem(item dtos.ActionItem, index int, ve *ValidationErrors) {
	prefix := fmt.Sprintf("action_items[%d]", index)

	if strings.TrimSpace(item.Task) == "" {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    fmt.Sprintf("%s.task", prefix),
			Value:        item.Task,
			Rule:         "required",
			Message:      "action item task must not be empty",
			SuggestedFix: "Provide a concrete task description or remove this action item",
		})
	} else {
		checkPlaceholder(fmt.Sprintf("%s.task", prefix), item.Task, ve)
	}

	if strings.TrimSpace(item.PIC) != "" {
		checkPlaceholder(fmt.Sprintf("%s.pic", prefix), item.PIC, ve)
	}

	if strings.TrimSpace(item.Deadline) != "" {
		checkPlaceholder(fmt.Sprintf("%s.deadline", prefix), item.Deadline, ve)
	}
}

// validateDecision validates a Decision struct.
func validateDecision(decision dtos.Decision, index int, ve *ValidationErrors) {
	prefix := fmt.Sprintf("decisions_made[%d]", index)

	if strings.TrimSpace(decision.Description) == "" {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    fmt.Sprintf("%s.description", prefix),
			Value:        decision.Description,
			Rule:         "required",
			Message:      "decision description must not be empty",
			SuggestedFix: "Provide a meaningful description of the decision made",
		})
	} else {
		checkPlaceholder(fmt.Sprintf("%s.description", prefix), decision.Description, ve)
	}
}

// validateOpenIssue validates an OpenIssue struct.
func validateOpenIssue(issue dtos.OpenIssue, index int, ve *ValidationErrors) {
	prefix := fmt.Sprintf("open_issues[%d]", index)

	if strings.TrimSpace(issue.Description) == "" {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    fmt.Sprintf("%s.description", prefix),
			Value:        issue.Description,
			Rule:         "required",
			Message:      "open issue description must not be empty",
			SuggestedFix: "Provide a description of the open issue or remove it",
		})
	} else {
		checkPlaceholder(fmt.Sprintf("%s.description", prefix), issue.Description, ve)
	}
}

// validateRisk validates a Risk struct.
func validateRisk(risk dtos.Risk, index int, ve *ValidationErrors) {
	prefix := fmt.Sprintf("risks_and_mitigation[%d]", index)

	if strings.TrimSpace(risk.Description) == "" {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    fmt.Sprintf("%s.description", prefix),
			Value:        risk.Description,
			Rule:         "required",
			Message:      "risk description must not be empty",
			SuggestedFix: "Provide a description of the risk or remove it",
		})
	} else {
		checkPlaceholder(fmt.Sprintf("%s.description", prefix), risk.Description, ve)
	}
}

// checkPlaceholder checks a string value for placeholder patterns and adds a validation error if found.
func checkPlaceholder(fieldPath string, value string, ve *ValidationErrors) {
	if placeholderRegex.MatchString(value) {
		ve.Errors = append(ve.Errors, &ValidationError{
			FieldPath:    fieldPath,
			Value:        value,
			Rule:         "no_placeholder",
			Message:      fmt.Sprintf("field contains placeholder text matching pattern"),
			SuggestedFix: "Replace placeholder with actual content from the meeting transcript",
		})
	}
}
