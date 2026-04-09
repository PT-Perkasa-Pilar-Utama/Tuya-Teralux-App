package services

import (
	"sensio/domain/models/rag/dtos"
	"strings"
)

// ============================================================
// Table Parsing — Extracts structured data from markdown tables
// ============================================================

// splitMarkdownTableRow splits a markdown table row into columns.
func splitMarkdownTableRow(row string) []string {
	row = strings.TrimSpace(row)
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")

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

// parseActionItemsFromSection extracts action items from a section's content or table.
func parseActionItemsFromSection(s section) []dtos.ActionItem {
	var items []dtos.ActionItem

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

// ParseActionItemsFromMarkdown parses action items from a markdown table.
func ParseActionItemsFromMarkdown(markdown string) []dtos.ActionItem {
	var items []dtos.ActionItem
	lines := strings.Split(markdown, "\n")
	skipNext := false
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

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
