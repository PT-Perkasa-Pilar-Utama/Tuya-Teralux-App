package services

import (
	"fmt"
	"strings"

	"sensio/domain/models/rag/dtos"
)

// GenerateMarkdown generates well-formatted markdown from a validated CanonicalMeetingSummary.
// Empty sections are omitted entirely — no placeholders or "N/A" text is rendered.
// The output is consistent regardless of which AI provider generated the underlying data.
func GenerateMarkdown(summary *dtos.CanonicalMeetingSummary) string {
	if summary == nil {
		return ""
	}

	var sb strings.Builder

	// Meeting title (required)
	if summary.Metadata.MeetingTitle != "" {
		sb.WriteString(fmt.Sprintf("# %s\n\n", summary.Metadata.MeetingTitle))
	}

	// Metadata block
	writeMetadataBlock(&sb, &summary.Metadata)

	// Agenda (required)
	if summary.Agenda != "" {
		sb.WriteString("## Agenda\n\n")
		sb.WriteString(summary.Agenda)
		sb.WriteString("\n\n")
	}

	// Background and Objective
	if summary.BackgroundAndObjective != "" {
		sb.WriteString("## Background & Objective\n\n")
		sb.WriteString(summary.BackgroundAndObjective)
		sb.WriteString("\n\n")
	}

	// Main Discussion Sections
	if len(summary.MainDiscussionSections) > 0 {
		sb.WriteString("## Main Discussion\n\n")
		for _, section := range summary.MainDiscussionSections {
			writeDiscussionSection(&sb, section)
		}
	}

	// Roles and Responsibilities
	if len(summary.RolesAndResponsibilities) > 0 {
		sb.WriteString("## Roles & Responsibilities\n\n")
		writeRolesAndResponsibilities(&sb, summary.RolesAndResponsibilities)
	}

	// Action Items
	if len(summary.ActionItems) > 0 {
		sb.WriteString("## Action Items\n\n")
		writeActionItems(&sb, summary.ActionItems)
	}

	// Decisions Made
	if len(summary.DecisionsMade) > 0 {
		sb.WriteString("## Decisions Made\n\n")
		writeDecisions(&sb, summary.DecisionsMade)
	}

	// Open Issues
	if len(summary.OpenIssues) > 0 {
		sb.WriteString("## Open Issues\n\n")
		writeOpenIssues(&sb, summary.OpenIssues)
	}

	// Risks and Mitigation
	if len(summary.RisksAndMitigation) > 0 {
		sb.WriteString("## Risks & Mitigation\n\n")
		writeRisks(&sb, summary.RisksAndMitigation)
	}

	// Additional Notes
	if summary.AdditionalNotes != "" {
		sb.WriteString("## Additional Notes\n\n")
		sb.WriteString(summary.AdditionalNotes)
		sb.WriteString("\n\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// writeMetadataBlock writes the meeting metadata as a formatted block.
func writeMetadataBlock(sb *strings.Builder, meta *dtos.SummaryMetadata) {
	var lines []string

	if meta.Date != "" {
		lines = append(lines, fmt.Sprintf("- **Date:** %s", meta.Date))
	}
	if meta.Location != "" {
		lines = append(lines, fmt.Sprintf("- **Location:** %s", meta.Location))
	}
	if len(meta.Participants) > 0 {
		lines = append(lines, fmt.Sprintf("- **Participants:** %s", strings.Join(meta.Participants, ", ")))
	}
	if meta.Context != "" {
		lines = append(lines, fmt.Sprintf("- **Context:** %s", meta.Context))
	}
	if meta.Style != "" {
		lines = append(lines, fmt.Sprintf("- **Style:** %s", meta.Style))
	}

	if len(lines) > 0 {
		sb.WriteString(strings.Join(lines, "\n"))
		sb.WriteString("\n\n")
	}
}

// writeDiscussionSection writes a single discussion section to the markdown builder.
func writeDiscussionSection(sb *strings.Builder, section dtos.DiscussionSection) {
	if section.Title == "" {
		return
	}

	sb.WriteString(fmt.Sprintf("### %s\n\n", section.Title))

	// Key points
	if len(section.KeyPoints) > 0 {
		for _, point := range section.KeyPoints {
			if point.Content == "" {
				continue
			}
			sb.WriteString(fmt.Sprintf("- %s", point.Content))
			if point.Speaker != "" {
				sb.WriteString(fmt.Sprintf(" *(%s)*", point.Speaker))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Decisions within this section
	if len(section.Decisions) > 0 {
		sb.WriteString("**Decisions:**\n\n")
		for i, decision := range section.Decisions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, decision))
		}
		sb.WriteString("\n")
	}

	// Action items within this section
	if len(section.ActionItems) > 0 {
		sb.WriteString("**Action Items:**\n\n")
		sb.WriteString("| # | Task | PIC | Deadline |\n")
		sb.WriteString("|---|------|-----|----------|\n")
		for i, item := range section.ActionItems {
			// Parse simple action item format: "task - PIC - deadline" or just "task"
			task := item
			pic := ""
			deadline := ""
			parts := strings.SplitN(item, " - ", 3)
			if len(parts) >= 1 {
				task = strings.TrimSpace(parts[0])
			}
			if len(parts) >= 2 {
				pic = strings.TrimSpace(parts[1])
			}
			if len(parts) >= 3 {
				deadline = strings.TrimSpace(parts[2])
			}
			sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n", i+1, escapeMarkdown(task), escapeMarkdown(pic), escapeMarkdown(deadline)))
		}
		sb.WriteString("\n")
	}
}

// writeRolesAndResponsibilities writes roles and responsibilities as a markdown table.
func writeRolesAndResponsibilities(sb *strings.Builder, roles []dtos.RoleResponsibility) {
	sb.WriteString("| Role | Assigned To | Description |\n")
	sb.WriteString("|------|-------------|-------------|\n")
	for _, role := range roles {
		if role.Role == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
			escapeMarkdown(role.Role),
			escapeMarkdown(role.AssignedTo),
			escapeMarkdown(role.Description)))
	}
	sb.WriteString("\n")
}

// writeActionItems writes action items as a markdown table.
func writeActionItems(sb *strings.Builder, items []dtos.ActionItem) {
	sb.WriteString("| # | Task | PIC | Deadline | Status |\n")
	sb.WriteString("|---|------|-----|----------|--------|\n")
	for i, item := range items {
		if item.Task == "" {
			continue
		}
		status := item.Status
		if status == "" {
			status = "Open"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			i+1,
			escapeMarkdown(item.Task),
			escapeMarkdown(item.PIC),
			escapeMarkdown(item.Deadline),
			escapeMarkdown(status)))
	}
	sb.WriteString("\n")
}

// writeDecisions writes decisions as a numbered list.
func writeDecisions(sb *strings.Builder, decisions []dtos.Decision) {
	for i, decision := range decisions {
		if decision.Description == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("%d. **%s**", i+1, escapeMarkdown(decision.Description)))
		if decision.Rationale != "" {
			sb.WriteString(fmt.Sprintf("\n   *Rationale: %s*", escapeMarkdown(decision.Rationale)))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

// writeOpenIssues writes open issues as a bulleted list.
func writeOpenIssues(sb *strings.Builder, issues []dtos.OpenIssue) {
	for _, issue := range issues {
		if issue.Description == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("- **%s**", escapeMarkdown(issue.Description)))
		if issue.Owner != "" {
			sb.WriteString(fmt.Sprintf(" *(Owner: %s)*", escapeMarkdown(issue.Owner)))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

// writeRisks writes risks as a markdown table.
func writeRisks(sb *strings.Builder, risks []dtos.Risk) {
	sb.WriteString("| # | Description | Impact | Mitigation |\n")
	sb.WriteString("|---|-------------|--------|------------|\n")
	for i, risk := range risks {
		if risk.Description == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
			i+1,
			escapeMarkdown(risk.Description),
			escapeMarkdown(risk.Impact),
			escapeMarkdown(risk.Mitigation)))
	}
	sb.WriteString("\n")
}

// escapeMarkdown escapes special markdown characters to prevent formatting issues.
func escapeMarkdown(s string) string {
	if s == "" {
		return ""
	}
	// Escape pipe characters which would break table formatting
	s = strings.ReplaceAll(s, "|", "\\|")
	// Escape newlines which would break table rows
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}
