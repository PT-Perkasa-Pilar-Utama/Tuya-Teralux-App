package services

import (
	"regexp"
	"sensio/domain/models/rag/dtos"
	"strings"
)

// ============================================================
// Section Extraction — Parses markdown structure into sections
// ============================================================

// section represents a markdown section extracted from raw output.
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

		if strings.HasPrefix(trimmed, "- **") {
			inMetadataBlock = true
			parseMetadataLine(trimmed, meta)
			continue
		}

		if inMetadataBlock && trimmed != "" && !strings.HasPrefix(trimmed, "- **") {
			if !strings.HasSuffix(trimmed, ",") {
				inMetadataBlock = false
			}
		}
	}
}

// parseMetadataLine parses a line like "- **Date:** 2026-04-09".
func parseMetadataLine(line string, meta *dtos.SummaryMetadata) {
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

		if strings.HasPrefix(trimmed, "## ") && len(trimmed) > 3 {
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

		if strings.HasPrefix(trimmed, "### ") && len(trimmed) > 4 && current != nil {
			if currentSub != nil {
				current.SubSections = append(current.SubSections, *currentSub)
			}
			currentSub = &discussionSubSection{
				Title:   strings.TrimSpace(trimmed[4:]),
				Content: "",
			}
			continue
		}

		if current != nil {
			if currentSub != nil {
				currentSub.Content += line + "\n"
			} else {
				current.Content += line + "\n"
			}
		}
	}

	if current != nil {
		if currentSub != nil {
			current.SubSections = append(current.SubSections, *currentSub)
		}
		sections = append(sections, *current)
	}

	return sections
}

// ExtractSectionContent extracts content between two section headers.
func ExtractSectionContent(markdown string, sectionTitle string) string {
	lines := strings.Split(markdown, "\n")
	var result strings.Builder
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") {
			headerTitle := strings.TrimSpace(trimmed[3:])
			if strings.EqualFold(headerTitle, sectionTitle) {
				inSection = true
				continue
			} else if inSection {
				break
			}
		}

		if inSection {
			result.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(result.String())
}
