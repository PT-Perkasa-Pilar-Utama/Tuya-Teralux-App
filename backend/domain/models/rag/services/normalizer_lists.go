package services

import (
	"regexp"
	"sensio/domain/models/rag/dtos"
	"strings"
)

// ============================================================
// List Parsing — Extracts structured data from markdown lists
// ============================================================

// stripBoldMarkers removes ** markers from text.
func stripBoldMarkers(text string) string {
	text = strings.ReplaceAll(text, "**", "")
	return strings.TrimSpace(text)
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

// extractRationale extracts rationale from text like "decision *Rationale: reason*".
func extractRationale(text string) (desc string, rationale string) {
	re := regexp.MustCompile(`^(.+?)\s+\*Rationale:\s*(.+?)\*`)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 3 {
		return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
	}
	return text, ""
}

// parseRolesAndResponsibilities parses a markdown table into RoleResponsibility structs.
func parseRolesAndResponsibilities(content string) []dtos.RoleResponsibility {
	var result []dtos.RoleResponsibility
	lines := strings.Split(content, "\n")

	skipNext := false
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
			if strings.Contains(trimmed, "Role") || strings.Contains(trimmed, "role") {
				skipNext = true
				continue
			}
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

			if strings.HasPrefix(trimmed, "- ") && len(trimmed) > 2 {
				point := strings.TrimSpace(trimmed[2:])
				if point != "" {
					content, speaker := extractSpeaker(point)
					ds.KeyPoints = append(ds.KeyPoints, dtos.DiscussionPoint{
						Content: content,
						Speaker: speaker,
					})
				}
			}

			if strings.HasPrefix(trimmed, "**Decisions:**") || strings.HasPrefix(trimmed, "**Decisions**") {
				continue
			}

			if decision := parseNumberedItem(trimmed); decision != "" {
				ds.Decisions = append(ds.Decisions, decision)
			}
		}

		if len(ds.KeyPoints) > 0 || len(ds.Decisions) > 0 {
			result = append(result, ds)
		}
	}

	return result
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

			content = stripBoldMarkers(content)

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

// parseOpenIssues parses a markdown bulleted list into OpenIssue structs.
func parseOpenIssues(content string) []dtos.OpenIssue {
	var issues []dtos.OpenIssue
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") && len(trimmed) > 2 {
			raw := strings.TrimSpace(trimmed[2:])
			if raw == "" {
				continue
			}
			// Try to extract owner: may have italics wrapper *(Owner: name)*
			ownerRe := regexp.MustCompile(`^(.+?)\s+\*?\(Owner:\s*(.+?)\)\*?`)
			oMatches := ownerRe.FindStringSubmatch(raw)
			if len(oMatches) >= 3 {
				issues = append(issues, dtos.OpenIssue{
					Description: stripBoldMarkers(strings.TrimSpace(oMatches[1])),
					Owner:       strings.TrimSpace(oMatches[2]),
				})
			} else {
				issues = append(issues, dtos.OpenIssue{
					Description: stripBoldMarkers(raw),
				})
			}
		}
	}

	return issues
}

// extractOwner extracts owner from text like "issue *(Owner: name)*".
func extractOwner(text string) (desc string, owner string) {
	re := regexp.MustCompile(`^(.+?)\s+\*?\(Owner:\s*(.+?)\)\*?\s*$`)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 3 {
		return stripBoldMarkers(strings.TrimSpace(matches[1])), strings.TrimSpace(matches[2])
	}
	return stripBoldMarkers(text), ""
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
			re := regexp.MustCompile(`^\*\*(.+?)\*\*\s+\*?\(Owner:\s*(.+?)\)\*?`)
			matches := re.FindStringSubmatch(content)
			if len(matches) >= 3 {
				issues = append(issues, dtos.OpenIssue{
					Description: strings.TrimSpace(matches[1]),
					Owner:       strings.TrimSpace(matches[2]),
				})
			} else {
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
