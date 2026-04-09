package services

import (
	"regexp"
	"sensio/domain/models/rag/dtos"
	"strconv"
	"strings"
)

// ============================================================
// Shared Helpers — Utility functions for normalizer
// ============================================================

// CountWords counts the approximate word count of a string.
func CountWords(s string) int {
	return len(strings.Fields(s))
}

// SanitizeEmptyParticipantEntries removes empty or whitespace-only participant entries.
func SanitizeEmptyParticipantEntries(participants []string) []string {
	var result []string
	for _, p := range participants {
		if strings.TrimSpace(p) != "" {
			result = append(result, p)
		}
	}
	return result
}

// ParseIDFromActionItem attempts to parse an integer ID from an action item task string.
func ParseIDFromActionItem(task string) int {
	re := regexp.MustCompile(`^(?:#?(\d+)[\.\)\:]\s*)`)
	matches := re.FindStringSubmatch(task)
	if len(matches) >= 2 {
		id, err := strconv.Atoi(matches[1])
		if err == nil {
			return id
		}
	}
	return 0
}

// SummaryNormalizer parses raw LLM markdown output into a CanonicalMeetingSummary struct.
// It is lenient by design — extracts what it can and leaves validation to a separate phase.
type SummaryNormalizer struct{}

// NewSummaryNormalizer creates a new normalizer instance.
func NewSummaryNormalizer() *SummaryNormalizer {
	return &SummaryNormalizer{}
}

// NormalizeToCanonical parses raw LLM markdown output into a CanonicalMeetingSummary.
// The provided metadata is used as the base for the summary's metadata fields.
// Missing sections in the markdown are mapped to nil/empty — validation catches issues downstream.
func (n *SummaryNormalizer) NormalizeToCanonical(rawMarkdown string, metadata dtos.SummaryMetadata) *dtos.CanonicalMeetingSummary {
	if rawMarkdown == "" {
		return &dtos.CanonicalMeetingSummary{
			Metadata: metadata,
		}
	}

	summary := &dtos.CanonicalMeetingSummary{
		Metadata: metadata,
	}

	lines := strings.Split(rawMarkdown, "\n")

	// Extract meeting title from first H1 header
	summary.Metadata.MeetingTitle = extractH1(lines, metadata.MeetingTitle)

	// Extract metadata from the metadata block
	extractMetadataBlock(lines, &summary.Metadata)

	// Extract sections by heading level
	sections := extractSections(lines)

	// Map sections to canonical fields
	for _, section := range sections {
		switch strings.ToLower(section.Title) {
		case "agenda":
			summary.Agenda = strings.TrimSpace(section.Content)
		case "background & objective", "background and objective", "background":
			summary.BackgroundAndObjective = strings.TrimSpace(section.Content)
		case "main discussion", "discussion", "main discussion points":
			summary.MainDiscussionSections = parseDiscussionSections(section.SubSections)
		case "roles & responsibilities", "roles and responsibilities", "roles":
			summary.RolesAndResponsibilities = parseRolesAndResponsibilities(section.Content)
		case "action items", "actions", "action points":
			summary.ActionItems = parseActionItemsFromSection(section)
		case "decisions made", "decisions", "key decisions":
			summary.DecisionsMade = parseDecisionsFromSection(section)
		case "open issues", "open questions", "unresolved":
			summary.OpenIssues = parseOpenIssues(section.Content)
		case "risks & mitigation", "risks and mitigation", "risks":
			summary.RisksAndMitigation = parseRisksFromSection(section)
		case "additional notes", "notes", "other notes":
			summary.AdditionalNotes = strings.TrimSpace(section.Content)
		}
	}

	return summary
}

// NormalizeFromStructuredArtifacts constructs a CanonicalMeetingSummary directly
// from structured artifacts (e.g., IntermediateSummaryNote from hierarchical summarization).
// This is more reliable than parsing free-form markdown.
func (n *SummaryNormalizer) NormalizeFromStructuredArtifacts(
	metadata dtos.SummaryMetadata,
	agenda string,
	actionItems []dtos.ActionItem,
	decisions []dtos.Decision,
	openIssues []dtos.OpenIssue,
	risks []dtos.Risk,
) *dtos.CanonicalMeetingSummary {
	return &dtos.CanonicalMeetingSummary{
		Metadata:           metadata,
		Agenda:             agenda,
		ActionItems:        actionItems,
		DecisionsMade:      decisions,
		OpenIssues:         openIssues,
		RisksAndMitigation: risks,
	}
}
