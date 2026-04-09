package services

import (
	"strings"
	"testing"

	"sensio/domain/models/rag/dtos"
)

// ============================================================
// Summary Normalizer — Production Tests
// ============================================================

func TestNormalizeToCanonical_TypicalLLMOutput_AllSectionsPopulated(t *testing.T) {
	markdown := `# Q2 Planning Review

- **Date:** 2026-04-09
- **Location:** Meeting Room A
- **Participants:** Alice, Bob, Charlie
- **Context:** Quarterly planning session
- **Style:** minutes

## Agenda

Review Q2 objectives and align on priorities for the engineering team.

## Background & Objective

This meeting was called to finalize Q2 goals and allocate resources across teams.

## Main Discussion

### Budget Allocation

- Budget increased by 10% for engineering *(Alice)*
- New hiring plan approved for Q2

**Decisions:**

1. **Approve 10% budget increase**
   *Rationale: To support accelerated roadmap*

### Timeline Review

- Sprint velocity improving over last quarter
- CI/CD pipeline needs attention

## Action Items

| # | Task | PIC | Deadline | Status |
|---|------|-----|----------|--------|
| 1 | Update project roadmap | Bob | 2026-04-20 | Open |
| 2 | Submit budget proposal | Alice | 2026-04-15 | Open |

## Decisions Made

1. **Adopt new CI/CD pipeline**
   *Rationale: Faster deployment cycles*
2. **Hire 2 senior engineers**

## Open Issues

- **Vendor contract renewal pending** *(Owner: Charlie)*
- **Office space allocation for new hires**

## Risks & Mitigation

| # | Description | Impact | Mitigation |
|---|-------------|--------|------------|
| 1 | Delayed hiring | High | Use contractors temporarily |
| 2 | Budget overrun | Medium | Monthly review checkpoints |

## Additional Notes

Next review scheduled for May.`

	meta := dtos.SummaryMetadata{
		MeetingTitle: "Q2 Planning Review",
		Date:         "2026-04-09",
		Location:     "Meeting Room A",
		Participants: []string{"Alice", "Bob", "Charlie"},
		Context:      "Quarterly planning session",
		Style:        "minutes",
		Language:     "en",
	}

	n := NewSummaryNormalizer()
	summary := n.NormalizeToCanonical(markdown, meta)

	// Verify metadata
	if summary.Metadata.MeetingTitle != "Q2 Planning Review" {
		t.Errorf("expected meeting title 'Q2 Planning Review', got %q", summary.Metadata.MeetingTitle)
	}

	// Verify agenda
	if !strings.Contains(summary.Agenda, "Review Q2 objectives") {
		t.Errorf("expected agenda to contain 'Review Q2 objectives', got %q", summary.Agenda)
	}

	// Verify background
	if !strings.Contains(summary.BackgroundAndObjective, "finalize Q2 goals") {
		t.Errorf("expected background to contain 'finalize Q2 goals', got %q", summary.BackgroundAndObjective)
	}

	// Verify discussion sections
	if len(summary.MainDiscussionSections) == 0 {
		t.Error("expected discussion sections to be populated")
	} else {
		found := false
		for _, s := range summary.MainDiscussionSections {
			if s.Title == "Budget Allocation" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'Budget Allocation' discussion section")
		}
	}

	// Verify action items
	if len(summary.ActionItems) == 0 {
		t.Error("expected action items to be populated")
	} else {
		if summary.ActionItems[0].Task != "Update project roadmap" {
			t.Errorf("expected first action item task 'Update project roadmap', got %q", summary.ActionItems[0].Task)
		}
		if summary.ActionItems[0].PIC != "Bob" {
			t.Errorf("expected PIC 'Bob', got %q", summary.ActionItems[0].PIC)
		}
	}

	// Verify decisions
	if len(summary.DecisionsMade) == 0 {
		t.Error("expected decisions to be populated")
	}

	// Verify risks
	if len(summary.RisksAndMitigation) == 0 {
		t.Error("expected risks to be populated")
	}

	// Verify additional notes
	if !strings.Contains(summary.AdditionalNotes, "Next review") {
		t.Errorf("expected additional notes to contain 'Next review', got %q", summary.AdditionalNotes)
	}
}

func TestNormalizeToCanonical_MinimalOutput_OnlyRequiredFields(t *testing.T) {
	markdown := `# Quick Sync

## Agenda

Daily standup update

- Discussed progress on feature X
- Blockers: waiting on design review`

	meta := dtos.SummaryMetadata{
		MeetingTitle: "Quick Sync",
		Language:     "en",
	}

	n := NewSummaryNormalizer()
	summary := n.NormalizeToCanonical(markdown, meta)

	if summary.Metadata.MeetingTitle != "Quick Sync" {
		t.Errorf("expected meeting title 'Quick Sync', got %q", summary.Metadata.MeetingTitle)
	}
	if summary.Metadata.Language != "en" {
		t.Errorf("expected language 'en', got %q", summary.Metadata.Language)
	}
	if !strings.Contains(summary.Agenda, "Daily standup") {
		t.Errorf("expected agenda to contain 'Daily standup', got %q", summary.Agenda)
	}

	// Optional sections should be empty/nil
	if len(summary.ActionItems) != 0 {
		t.Errorf("expected no action items, got %d", len(summary.ActionItems))
	}
	if len(summary.DecisionsMade) != 0 {
		t.Errorf("expected no decisions, got %d", len(summary.DecisionsMade))
	}
}

func TestNormalizeToCanonical_EmptyInput_ReturnsEmptyCanonical(t *testing.T) {
	meta := dtos.SummaryMetadata{
		MeetingTitle: "Test",
		Language:     "en",
	}

	n := NewSummaryNormalizer()
	summary := n.NormalizeToCanonical("", meta)

	if summary == nil {
		t.Fatal("expected non-nil summary for empty input")
	}
	if summary.Metadata.MeetingTitle != "Test" {
		t.Errorf("expected metadata to be preserved, got %q", summary.Metadata.MeetingTitle)
	}
	if summary.Agenda != "" {
		t.Errorf("expected empty agenda, got %q", summary.Agenda)
	}
}

func TestNormalizeToCanonical_MalformedOutput_NoPanic(t *testing.T) {
	markdown := `This is just random text with no structure at all.
No headers, no tables, nothing meaningful.`

	meta := dtos.SummaryMetadata{
		MeetingTitle: "Unknown",
		Language:     "en",
	}

	n := NewSummaryNormalizer()
	// Should not panic
	summary := n.NormalizeToCanonical(markdown, meta)

	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	// Should preserve metadata
	if summary.Metadata.Language != "en" {
		t.Errorf("expected language preserved, got %q", summary.Metadata.Language)
	}
}

func TestNormalizeToCanonical_ActionItemsTable_Parsed(t *testing.T) {
	markdown := `# Meeting

## Action Items

| # | Task | PIC | Deadline | Status |
|---|------|-----|----------|--------|
| 1 | Review PR #123 | Alice | 2026-04-15 | Open |
| 2 | Update docs | Bob | 2026-04-20 | In Progress |`

	meta := dtos.SummaryMetadata{
		MeetingTitle: "Meeting",
		Language:     "en",
	}

	n := NewSummaryNormalizer()
	summary := n.NormalizeToCanonical(markdown, meta)

	if len(summary.ActionItems) != 2 {
		t.Fatalf("expected 2 action items, got %d", len(summary.ActionItems))
	}
	if summary.ActionItems[0].Task != "Review PR #123" {
		t.Errorf("expected task 'Review PR #123', got %q", summary.ActionItems[0].Task)
	}
	if summary.ActionItems[0].PIC != "Alice" {
		t.Errorf("expected PIC 'Alice', got %q", summary.ActionItems[0].PIC)
	}
	if summary.ActionItems[0].Deadline != "2026-04-15" {
		t.Errorf("expected deadline '2026-04-15', got %q", summary.ActionItems[0].Deadline)
	}
	if summary.ActionItems[0].Status != "Open" {
		t.Errorf("expected status 'Open', got %q", summary.ActionItems[0].Status)
	}
}

func TestNormalizeToCanonical_DecisionsNumberedList_Parsed(t *testing.T) {
	markdown := `# Meeting

## Decisions Made

1. **Adopt new framework**
   *Rationale: Better performance*
2. **Migrate to microservices**
3. **Keep monolith for now**
   *Rationale: Too risky to migrate now*`

	meta := dtos.SummaryMetadata{
		MeetingTitle: "Meeting",
		Language:     "en",
	}

	n := NewSummaryNormalizer()
	summary := n.NormalizeToCanonical(markdown, meta)

	if len(summary.DecisionsMade) != 3 {
		t.Fatalf("expected 3 decisions, got %d", len(summary.DecisionsMade))
	}
	if summary.DecisionsMade[0].Description != "Adopt new framework" {
		t.Errorf("expected decision 'Adopt new framework', got %q", summary.DecisionsMade[0].Description)
	}
	if summary.DecisionsMade[0].Rationale != "Better performance" {
		t.Errorf("expected rationale 'Better performance', got %q", summary.DecisionsMade[0].Rationale)
	}
	// Decision without rationale
	if summary.DecisionsMade[1].Rationale != "" {
		t.Errorf("expected empty rationale, got %q", summary.DecisionsMade[1].Rationale)
	}
}

func TestNormalizeToCanonical_RisksTable_Parsed(t *testing.T) {
	markdown := `# Meeting

## Risks & Mitigation

| # | Description | Impact | Mitigation |
|---|-------------|--------|------------|
| 1 | Server downtime | High | Redundant servers |
| 2 | Data loss | Critical | Daily backups |`

	meta := dtos.SummaryMetadata{
		MeetingTitle: "Meeting",
		Language:     "en",
	}

	n := NewSummaryNormalizer()
	summary := n.NormalizeToCanonical(markdown, meta)

	if len(summary.RisksAndMitigation) != 2 {
		t.Fatalf("expected 2 risks, got %d", len(summary.RisksAndMitigation))
	}
	if summary.RisksAndMitigation[0].Description != "Server downtime" {
		t.Errorf("expected risk 'Server downtime', got %q", summary.RisksAndMitigation[0].Description)
	}
	if summary.RisksAndMitigation[0].Impact != "High" {
		t.Errorf("expected impact 'High', got %q", summary.RisksAndMitigation[0].Impact)
	}
	if summary.RisksAndMitigation[0].Mitigation != "Redundant servers" {
		t.Errorf("expected mitigation 'Redundant servers', got %q", summary.RisksAndMitigation[0].Mitigation)
	}
}

func TestNormalizeToCanonical_OpenIssuesBulletedList_Parsed(t *testing.T) {
	markdown := `# Meeting

## Open Issues

- **Vendor contract renewal pending** *(Owner: Charlie)*
- **Office space allocation for new hires**`

	meta := dtos.SummaryMetadata{
		MeetingTitle: "Meeting",
		Language:     "en",
	}

	n := NewSummaryNormalizer()
	summary := n.NormalizeToCanonical(markdown, meta)

	if len(summary.OpenIssues) != 2 {
		t.Fatalf("expected 2 open issues, got %d", len(summary.OpenIssues))
	}
	if summary.OpenIssues[0].Description != "Vendor contract renewal pending" {
		t.Errorf("expected issue description 'Vendor contract renewal pending', got %q", summary.OpenIssues[0].Description)
	}
	if summary.OpenIssues[0].Owner != "Charlie" {
		t.Errorf("expected owner 'Charlie', got %q", summary.OpenIssues[0].Owner)
	}
	if summary.OpenIssues[1].Owner != "" {
		t.Errorf("expected empty owner, got %q", summary.OpenIssues[1].Owner)
	}
}

func TestNormalizeToCanonical_MetadataBlockParsed(t *testing.T) {
	markdown := `# Planning Session

- **Date:** 2026-04-10
- **Location:** Building B, Floor 3
- **Participants:** Alice, Bob, Charlie, Diana
- **Context:** Sprint planning
- **Style:** executive

## Agenda

Plan sprint 24`

	meta := dtos.SummaryMetadata{}

	n := NewSummaryNormalizer()
	summary := n.NormalizeToCanonical(markdown, meta)

	if summary.Metadata.Date != "2026-04-10" {
		t.Errorf("expected date '2026-04-10', got %q", summary.Metadata.Date)
	}
	if summary.Metadata.Location != "Building B, Floor 3" {
		t.Errorf("expected location 'Building B, Floor 3', got %q", summary.Metadata.Location)
	}
	if len(summary.Metadata.Participants) != 4 {
		t.Errorf("expected 4 participants, got %d", len(summary.Metadata.Participants))
	}
	if summary.Metadata.Context != "Sprint planning" {
		t.Errorf("expected context 'Sprint planning', got %q", summary.Metadata.Context)
	}
	if summary.Metadata.Style != "executive" {
		t.Errorf("expected style 'executive', got %q", summary.Metadata.Style)
	}
}

func TestNormalizeFromStructuredArtifacts_DirectConstruction(t *testing.T) {
	meta := dtos.SummaryMetadata{
		MeetingTitle: "Team Sync",
		Language:     "en",
	}
	actionItems := []dtos.ActionItem{
		{Task: "Fix bug #123", PIC: "Alice"},
	}
	decisions := []dtos.Decision{
		{Description: "Use Go 1.22"},
	}
	risks := []dtos.Risk{
		{Description: "Timeline tight", Impact: "Medium"},
	}

	n := NewSummaryNormalizer()
	summary := n.NormalizeFromStructuredArtifacts(meta, "Sprint sync", actionItems, decisions, []dtos.OpenIssue{}, risks)

	if summary.Metadata.MeetingTitle != "Team Sync" {
		t.Errorf("expected title 'Team Sync', got %q", summary.Metadata.MeetingTitle)
	}
	if summary.Agenda != "Sprint sync" {
		t.Errorf("expected agenda 'Sprint sync', got %q", summary.Agenda)
	}
	if len(summary.ActionItems) != 1 {
		t.Errorf("expected 1 action item, got %d", len(summary.ActionItems))
	}
	if len(summary.DecisionsMade) != 1 {
		t.Errorf("expected 1 decision, got %d", len(summary.DecisionsMade))
	}
	if len(summary.RisksAndMitigation) != 1 {
		t.Errorf("expected 1 risk, got %d", len(summary.RisksAndMitigation))
	}
}

func TestExtractSectionContent_ExtractsCorrectSection(t *testing.T) {
	markdown := `# Meeting

## Agenda

This is the agenda content.

## Discussion

This is discussion content.

## Action Items

These are action items.`

	agenda := ExtractSectionContent(markdown, "Agenda")
	if !strings.Contains(agenda, "This is the agenda content") {
		t.Errorf("expected agenda content, got %q", agenda)
	}

	discussion := ExtractSectionContent(markdown, "Discussion")
	if !strings.Contains(discussion, "This is discussion content") {
		t.Errorf("expected discussion content, got %q", discussion)
	}
}

func TestParseActionItemsFromMarkdown_TableFormat(t *testing.T) {
	markdown := `| # | Task | PIC | Deadline | Status |
|---|------|-----|----------|--------|
| 1 | Deploy v2 | Alice | 2026-05-01 | Open |
| 2 | Write tests | Bob | 2026-04-20 | In Progress |`

	items := ParseActionItemsFromMarkdown(markdown)

	if len(items) != 2 {
		t.Fatalf("expected 2 action items, got %d", len(items))
	}
	if items[0].Task != "Deploy v2" {
		t.Errorf("expected task 'Deploy v2', got %q", items[0].Task)
	}
	if items[0].PIC != "Alice" {
		t.Errorf("expected PIC 'Alice', got %q", items[0].PIC)
	}
	if items[0].Deadline != "2026-05-01" {
		t.Errorf("expected deadline '2026-05-01', got %q", items[0].Deadline)
	}
	if items[0].Status != "Open" {
		t.Errorf("expected status 'Open', got %q", items[0].Status)
	}
}

func TestParseDecisionsFromMarkdown_NumberedList(t *testing.T) {
	markdown := `1. **First decision**
   *Rationale: Good idea*
2. **Second decision**
3. **Third decision**
   *Rationale: Necessary*`

	decisions := ParseDecisionsFromMarkdown(markdown)

	if len(decisions) != 3 {
		t.Fatalf("expected 3 decisions, got %d", len(decisions))
	}
	if decisions[0].Description != "First decision" {
		t.Errorf("expected 'First decision', got %q", decisions[0].Description)
	}
	if decisions[0].Rationale != "Good idea" {
		t.Errorf("expected rationale 'Good idea', got %q", decisions[0].Rationale)
	}
}

func TestParseRisksFromMarkdown_TableFormat(t *testing.T) {
	markdown := `| # | Description | Impact | Mitigation |
|---|-------------|--------|------------|
| 1 | Server crash | High | Backup server |
| 2 | Data loss | Critical | Backup daily |`

	risks := ParseRisksFromMarkdown(markdown)

	if len(risks) != 2 {
		t.Fatalf("expected 2 risks, got %d", len(risks))
	}
	if risks[0].Description != "Server crash" {
		t.Errorf("expected 'Server crash', got %q", risks[0].Description)
	}
	if risks[0].Impact != "High" {
		t.Errorf("expected impact 'High', got %q", risks[0].Impact)
	}
}

func TestParseOpenIssuesFromMarkdown_BulletedList(t *testing.T) {
	markdown := `- **Issue one** *(Owner: Alice)*
- **Issue two**
- **Issue three** *(Owner: Bob)*`

	issues := ParseOpenIssuesFromMarkdown(markdown)

	if len(issues) != 3 {
		t.Fatalf("expected 3 issues, got %d", len(issues))
	}
	if issues[0].Description != "Issue one" {
		t.Errorf("expected 'Issue one', got %q", issues[0].Description)
	}
	if issues[0].Owner != "Alice" {
		t.Errorf("expected owner 'Alice', got %q", issues[0].Owner)
	}
	if issues[1].Owner != "" {
		t.Errorf("expected empty owner, got %q", issues[1].Owner)
	}
}

func TestSanitizeEmptyParticipantEntries_RemovesEmpty(t *testing.T) {
	input := []string{"Alice", "", "  ", "Bob", ""}
	result := SanitizeEmptyParticipantEntries(input)

	if len(result) != 2 {
		t.Errorf("expected 2 participants, got %d", len(result))
	}
	if result[0] != "Alice" || result[1] != "Bob" {
		t.Errorf("expected ['Alice', 'Bob'], got %v", result)
	}
}

func TestCountWords_CountsCorrectly(t *testing.T) {
	if CountWords("hello world") != 2 {
		t.Errorf("expected 2 words, got %d", CountWords("hello world"))
	}
	if CountWords("") != 0 {
		t.Errorf("expected 0 words for empty string, got %d", CountWords(""))
	}
	if CountWords("one") != 1 {
		t.Errorf("expected 1 word, got %d", CountWords("one"))
	}
}
