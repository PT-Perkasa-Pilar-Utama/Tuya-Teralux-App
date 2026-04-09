package services

import (
	"regexp"
	"sensio/domain/models/rag/dtos"
	"strconv"
	"strings"
)

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

	// Extract metadata from the metadata block (lines starting with "- **Date:**" etc.)
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

// section represents a markdown section extracted from the raw output.
type section struct {
	Title       string
	Content     string
	SubSections []discussionSubSection
}

// discussionSubSection represents a subsection within a discussion section.
type discussionSubSection struct {
	Title   string
	Content string
}

// extractH1 extracts the meeting title from the first H1 header.
func extractH1(lines []string, fallback string) string {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && len(trimmed) > 2 {
			return strings.TrimSpace(trimmed[2:])
		}
	}
	return fallback
}

// extractMetadataBlock extracts metadata from lines like "- **Date:** value".
func extractMetadataBlock(lines []string, meta *dtos.SummaryMetadata) {
	inMetadataBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect metadata block (lines after the H1 that start with "- **")
		if strings.HasPrefix(trimmed, "- **") {
			inMetadataBlock = true
			parseMetadataLine(trimmed, meta)
			continue
		}

		// Stop metadata block when we hit a non-metadata line (after we've started)
		if inMetadataBlock && trimmed != "" && !strings.HasPrefix(trimmed, "- **") {
			// Allow continuation if previous line ended with comma (multi-participant line)
			if !strings.HasSuffix(trimmed, ",") {
				inMetadataBlock = false
			}
		}
	}
}

// parseMetadataLine parses a line like "- **Date:** 2026-04-09".
func parseMetadataLine(line string, meta *dtos.SummaryMetadata) {
	// Extract key and value from "- **Key:** Value"
	re := regexp.MustCompile(`^- \*\*(.+?):\*\*\s*(.*)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 3 {
		return
	}

	key := strings.ToLower(strings.TrimSpace(matches[1]))
	value := strings.TrimSpace(matches[2])

	if value == "" {
		return
	}

	switch key {
	case "date":
		meta.Date = value
	case "location":
		meta.Location = value
	case "participants":
		// Parse comma-separated participants
		parts := strings.Split(value, ",")
		var participants []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				participants = append(participants, p)
			}
		}
		meta.Participants = participants
	case "context":
		meta.Context = value
	case "style":
		meta.Style = value
	}
}

// extractSections extracts H2 (##) sections from the markdown.
func extractSections(lines []string) []section {
	var sections []section
	var current *section
	var currentSub *discussionSubSection

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// H2 section header
		if strings.HasPrefix(trimmed, "## ") && len(trimmed) > 3 {
			// Save previous section
			if current != nil {
				if currentSub != nil {
					current.SubSections = append(current.SubSections, *currentSub)
					currentSub = nil
				}
				sections = append(sections, *current)
			}
			current = &section{
				Title:   strings.TrimSpace(trimmed[3:]),
				Content: "",
			}
			continue
		}

		// H3 section header (subsection within discussion)
		if strings.HasPrefix(trimmed, "### ") && len(trimmed) > 4 && current != nil {
			// Save previous subsection
			if currentSub != nil {
				current.SubSections = append(current.SubSections, *currentSub)
			}
			currentSub = &discussionSubSection{
				Title:   strings.TrimSpace(trimmed[4:]),
				Content: "",
			}
			continue
		}

		// Accumulate content
		if current != nil {
			if currentSub != nil {
				currentSub.Content += line + "\n"
			} else {
				current.Content += line + "\n"
			}
		}
	}

	// Save last section
	if current != nil {
		if currentSub != nil {
			current.SubSections = append(current.SubSections, *currentSub)
		}
		sections = append(sections, *current)
	}

	return sections
}

// parseDiscussionSections parses discussion subsections into DiscussionSection structs.
func parseDiscussionSections(subSections []discussionSubSection) []dtos.DiscussionSection {
	var result []dtos.DiscussionSection

	for _, sub := range subSections {
		if sub.Title == "" {
			continue
		}

		ds := dtos.DiscussionSection{
			Title: sub.Title,
		}

		lines := strings.Split(sub.Content, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Parse bullet points as key points
			if strings.HasPrefix(trimmed, "- ") && len(trimmed) > 2 {
				point := strings.TrimSpace(trimmed[2:])
				if point != "" {
					// Extract speaker if present: "content *(Speaker)*"
					content, speaker := extractSpeaker(point)
					ds.KeyPoints = append(ds.KeyPoints, dtos.DiscussionPoint{
						Content: content,
						Speaker: speaker,
					})
				}
			}

			// Parse inline decisions: "**Decisions:**" header followed by numbered list
			if strings.HasPrefix(trimmed, "**Decisions:**") || strings.HasPrefix(trimmed, "**Decisions**") {
				// Subsequent numbered lines are decisions
				continue
			}

			// Parse numbered decisions
			if decision := parseNumberedItem(trimmed); decision != "" {
				// Check if we're in a decisions context (previous line had "Decisions:")
				ds.Decisions = append(ds.Decisions, decision)
			}

			// Parse inline action items table rows
			if strings.HasPrefix(trimmed, "|") && strings.Contains(trimmed, "|") {
				// This might be a table row for action items
				// We'll handle it at the section level
			}
		}

		if len(ds.KeyPoints) > 0 || len(ds.Decisions) > 0 {
			result = append(result, ds)
		}
	}

	return result
}

// extractSpeaker extracts speaker info from content like "content *(Speaker Name)*".
func extractSpeaker(content string) (text string, speaker string) {
	re := regexp.MustCompile(`^(.+?)\s+\*\((.+?)\)\s*$`)
	matches := re.FindStringSubmatch(content)
	if len(matches) >= 3 {
		return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
	}
	return content, ""
}

// parseNumberedItem extracts content from a numbered list item like "1. Do something".
func parseNumberedItem(line string) string {
	re := regexp.MustCompile(`^\d+\.\s+(.+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// parseRolesAndResponsibilities parses a markdown table into RoleResponsibility structs.
func parseRolesAndResponsibilities(content string) []dtos.RoleResponsibility {
	var result []dtos.RoleResponsibility
	lines := strings.Split(content, "\n")

	skipNext := false // Skip the separator row
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if skipNext {
			skipNext = false
			continue
		}

		if strings.HasPrefix(trimmed, "|") && strings.Contains(trimmed, "|") {
			// Check if it's a header row
			if strings.Contains(trimmed, "Role") || strings.Contains(trimmed, "role") {
				skipNext = true
				continue
			}
			// Parse table row
			cols := splitMarkdownTableRow(trimmed)
			if len(cols) >= 2 {
				result = append(result, dtos.RoleResponsibility{
					Role:        strings.TrimSpace(cols[0]),
					AssignedTo:  strings.TrimSpace(cols[1]),
					Description: getMarkdownTableCol(cols, 2),
				})
			}
		}
	}

	return result
}

// parseActionItemsFromSection extracts action items from a section's content or table.
func parseActionItemsFromSection(s section) []dtos.ActionItem {
	var items []dtos.ActionItem

	// Check if section has a table (action items as markdown table)
	lines := strings.Split(s.Content, "\n")
	hasTable := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "|") && strings.Contains(strings.TrimSpace(line), "Task") {
			hasTable = true
			break
		}
	}

	if hasTable {
		skipNext := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || !strings.HasPrefix(trimmed, "|") {
				continue
			}
			if skipNext {
				skipNext = false
				continue
			}
			// Skip header
			if strings.Contains(trimmed, "Task") || strings.Contains(trimmed, "task") {
				skipNext = true
				continue
			}
			cols := splitMarkdownTableRow(trimmed)
			if len(cols) >= 2 {
				task := strings.TrimSpace(cols[1])
				if task == "" || task == "---" {
					continue
				}
				item := dtos.ActionItem{
					Task: task,
					PIC:  getMarkdownTableCol(cols, 2),
				}
				if len(cols) > 3 {
					item.Deadline = getMarkdownTableCol(cols, 3)
				}
				if len(cols) > 4 {
					item.Status = getMarkdownTableCol(cols, 4)
				}
				items = append(items, item)
			}
		}
		return items
	}

	// Fallback: parse bullet points as action items
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") && len(trimmed) > 2 {
			content := strings.TrimSpace(trimmed[2:])
			if content == "" {
				continue
			}
			// Try to parse "task - PIC - deadline" format
			parts := strings.SplitN(content, " - ", 3)
			item := dtos.ActionItem{
				Task: strings.TrimSpace(parts[0]),
			}
			if len(parts) >= 2 {
				item.PIC = strings.TrimSpace(parts[1])
			}
			if len(parts) >= 3 {
				item.Deadline = strings.TrimSpace(parts[2])
			}
			if item.Task != "" {
				items = append(items, item)
			}
		}
	}

	return items
}

// parseDecisionsFromSection extracts decisions from a section's numbered list.
func parseDecisionsFromSection(s section) []dtos.Decision {
	var decisions []dtos.Decision
	lines := strings.Split(s.Content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		re := regexp.MustCompile(`^\d+\.\s+(.+)`)
		matches := re.FindStringSubmatch(trimmed)
		if len(matches) >= 2 {
			content := strings.TrimSpace(matches[1])
			if content == "" {
				continue
			}

			// Strip bold markers: **text** → text
			content = stripBoldMarkers(content)

			// Check next line for rationale
			rationale := ""
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				rationaleRe := regexp.MustCompile(`^\*Rationale:\s*(.+?)\*`)
				rMatches := rationaleRe.FindStringSubmatch(nextLine)
				if len(rMatches) >= 2 {
					rationale = strings.TrimSpace(rMatches[1])
				}
			}

			decisions = append(decisions, dtos.Decision{
				Description: content,
				Rationale:   rationale,
			})
		}
	}

	return decisions
}

// stripBoldMarkers removes ** markers from text.
func stripBoldMarkers(text string) string {
	text = strings.ReplaceAll(text, "**", "")
	return strings.TrimSpace(text)
}

// extractRationale extracts rationale from text like "decision *Rationale: reason*".
func extractRationale(text string) (desc string, rationale string) {
	re := regexp.MustCompile(`^(.+?)\s+\*Rationale:\s*(.+?)\*`)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 3 {
		return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
	}
	return text, ""
}

// parseOpenIssues parses a markdown bulleted list into OpenIssue structs.
func parseOpenIssues(content string) []dtos.OpenIssue {
	var issues []dtos.OpenIssue
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") && len(trimmed) > 2 {
			content := strings.TrimSpace(trimmed[2:])
			if content == "" {
				continue
			}
			// Extract owner if present: "issue *(Owner: name)*"
			desc, owner := extractOwner(content)
			issues = append(issues, dtos.OpenIssue{
				Description: desc,
				Owner:       owner,
			})
		}
	}

	return issues
}

// extractOwner extracts owner from text like "issue *(Owner: name)*".
func extractOwner(text string) (desc string, owner string) {
	// Owner may be wrapped in italics: *(Owner: name)*
	re := regexp.MustCompile(`^(.+?)\s+\*?\(Owner:\s*(.+?)\)\*?\s*$`)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 3 {
		return stripBoldMarkers(strings.TrimSpace(matches[1])), strings.TrimSpace(matches[2])
	}
	return stripBoldMarkers(text), ""
}

// parseRisksFromSection extracts risks from a markdown table.
func parseRisksFromSection(s section) []dtos.Risk {
	var risks []dtos.Risk
	lines := strings.Split(s.Content, "\n")

	skipNext := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		// Skip header
		if strings.Contains(trimmed, "Description") && strings.Contains(trimmed, "Impact") {
			skipNext = true
			continue
		}
		cols := splitMarkdownTableRow(trimmed)
		if len(cols) >= 2 {
			desc := strings.TrimSpace(cols[1])
			if desc == "" || desc == "---" {
				continue
			}
			risk := dtos.Risk{
				Description: desc,
			}
			if len(cols) > 2 {
				risk.Impact = strings.TrimSpace(cols[2])
			}
			if len(cols) > 3 {
				risk.Mitigation = strings.TrimSpace(cols[3])
			}
			risks = append(risks, risk)
		}
	}

	return risks
}

// splitMarkdownTableRow splits a markdown table row into columns.
func splitMarkdownTableRow(row string) []string {
	row = strings.TrimSpace(row)
	// Remove leading and trailing pipes
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")

	// Split by pipe, being careful about escaped pipes
	var cols []string
	var current strings.Builder
	escaped := false

	for _, ch := range row {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '|' {
			cols = append(cols, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(ch)
	}
	cols = append(cols, current.String())

	return cols
}

// getMarkdownTableCol safely gets a column from a split table row.
func getMarkdownTableCol(cols []string, index int) string {
	if index < len(cols) {
		return strings.TrimSpace(cols[index])
	}
	return ""
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
	summary := &dtos.CanonicalMeetingSummary{
		Metadata:       metadata,
		Agenda:         agenda,
		ActionItems:    actionItems,
		DecisionsMade:  decisions,
		OpenIssues:     openIssues,
		RisksAndMitigation: risks,
	}

	return summary
}

// ParseActionItemsFromMarkdown parses action items from a markdown table.
// Useful for parsing LLM output that contains action items in table format.
func ParseActionItemsFromMarkdown(markdown string) []dtos.ActionItem {
	var items []dtos.ActionItem
	lines := strings.Split(markdown, "\n")
	skipNext := false
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect table header
		if strings.HasPrefix(trimmed, "|") && (strings.Contains(trimmed, "Task") || strings.Contains(trimmed, "task")) {
			inTable = true
			skipNext = true
			continue
		}

		if !inTable {
			continue
		}

		if skipNext {
			skipNext = false
			continue
		}

		if !strings.HasPrefix(trimmed, "|") {
			inTable = false
			continue
		}

		cols := splitMarkdownTableRow(trimmed)
		if len(cols) >= 2 {
			task := strings.TrimSpace(cols[1])
			if task == "" || task == "---" {
				continue
			}
			item := dtos.ActionItem{
				Task: task,
				PIC:  getMarkdownTableCol(cols, 2),
			}
			if len(cols) > 3 {
				item.Deadline = getMarkdownTableCol(cols, 3)
			}
			if len(cols) > 4 {
				item.Status = getMarkdownTableCol(cols, 4)
			}
			items = append(items, item)
		}
	}

	return items
}

// ParseDecisionsFromMarkdown parses decisions from numbered list markdown.
func ParseDecisionsFromMarkdown(markdown string) []dtos.Decision {
	var decisions []dtos.Decision
	lines := strings.Split(markdown, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		re := regexp.MustCompile(`^\d+\.\s+\*\*(.+?)\*\*`)
		matches := re.FindStringSubmatch(trimmed)
		if len(matches) >= 2 {
			decision := dtos.Decision{
				Description: strings.TrimSpace(matches[1]),
			}
			// Check next line for rationale
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				rationaleRe := regexp.MustCompile(`^\*Rationale:\s*(.+?)\*`)
				rMatches := rationaleRe.FindStringSubmatch(nextLine)
				if len(rMatches) >= 2 {
					decision.Rationale = strings.TrimSpace(rMatches[1])
				}
			}
			decisions = append(decisions, decision)
		}
	}

	return decisions
}

// ParseRisksFromMarkdown parses risks from a markdown table.
func ParseRisksFromMarkdown(markdown string) []dtos.Risk {
	var risks []dtos.Risk
	lines := strings.Split(markdown, "\n")
	skipNext := false
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "|") && strings.Contains(trimmed, "Description") && strings.Contains(trimmed, "Impact") {
			inTable = true
			skipNext = true
			continue
		}

		if !inTable {
			continue
		}

		if skipNext {
			skipNext = false
			continue
		}

		if !strings.HasPrefix(trimmed, "|") {
			inTable = false
			continue
		}

		cols := splitMarkdownTableRow(trimmed)
		if len(cols) >= 2 {
			desc := strings.TrimSpace(cols[1])
			if desc == "" || desc == "---" {
				continue
			}
			risk := dtos.Risk{
				Description: desc,
			}
			if len(cols) > 2 {
				risk.Impact = strings.TrimSpace(cols[2])
			}
			if len(cols) > 3 {
				risk.Mitigation = strings.TrimSpace(cols[3])
			}
			risks = append(risks, risk)
		}
	}

	return risks
}

// ParseOpenIssuesFromMarkdown parses open issues from a bulleted list.
func ParseOpenIssuesFromMarkdown(markdown string) []dtos.OpenIssue {
	var issues []dtos.OpenIssue
	lines := strings.Split(markdown, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") && len(trimmed) > 2 {
			content := strings.TrimSpace(trimmed[2:])
			if content == "" {
				continue
			}
			// Parse format: "**issue** *(Owner: name)*" or just "**issue**"
			// Note: Owner may be wrapped in italics: *(Owner: name)*
			re := regexp.MustCompile(`^\*\*(.+?)\*\*\s+\*?\(Owner:\s*(.+?)\)\*?`)
			matches := re.FindStringSubmatch(content)
			if len(matches) >= 3 {
				issues = append(issues, dtos.OpenIssue{
					Description: strings.TrimSpace(matches[1]),
					Owner:       strings.TrimSpace(matches[2]),
				})
			} else {
				// Just bold text: "**issue**"
				re2 := regexp.MustCompile(`^\*\*(.+?)\*\*`)
				matches2 := re2.FindStringSubmatch(content)
				if len(matches2) >= 2 {
					issues = append(issues, dtos.OpenIssue{
						Description: strings.TrimSpace(matches2[1]),
					})
				} else {
					issues = append(issues, dtos.OpenIssue{
						Description: stripBoldMarkers(content),
					})
				}
			}
		}
	}

	return issues
}

// ExtractSectionContent extracts content between two section headers.
func ExtractSectionContent(markdown string, sectionTitle string) string {
	lines := strings.Split(markdown, "\n")
	var result strings.Builder
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this line starts the target section
		if strings.HasPrefix(trimmed, "## ") {
			headerTitle := strings.TrimSpace(trimmed[3:])
			if strings.EqualFold(headerTitle, sectionTitle) {
				inSection = true
				continue
			} else if inSection {
				// We've hit the next section
				break
			}
		}

		if inSection {
			result.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(result.String())
}

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
	// Look for leading number patterns like "1. Task" or "#1 Task"
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
