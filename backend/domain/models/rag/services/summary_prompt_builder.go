package services

import (
	"fmt"
	"strings"
)

// PromptConfig encapsulates parameterized prompt generation for meeting summaries.
type PromptConfig struct {
	Assertiveness int
	Audience      string
	RiskScale     string
	Context       string
	Style         string
	Language      string

	// Metadata for MoM
	Date         string
	Location     string
	Participants string
}

func (pc *PromptConfig) AssertivenessPhrasing() string {
	switch {
	case pc.Assertiveness <= 3:
		return "Tone: Objective and cautious. Present findings with qualifier phrases like 'suggests', 'may indicate'."
	case pc.Assertiveness <= 6:
		return "Tone: Balanced and analytical. Professional analyst offering well-reasoned perspective."
	default:
		return "Tone: Assertive and analytical. Call out gaps and risks directly."
	}
}

func (pc *PromptConfig) AudienceGuidance() string {
	switch strings.ToLower(pc.Audience) {
	case "c-level":
		return "Audience: C-suite executives. ROI focus, strategic brevity."
	case "mixed":
		return "Audience: C-level + Directors. Strategic + operational context."
	default:
		return "Audience: Execution team. Tactical and strategic detail."
	}
}

func (pc *PromptConfig) RiskScoringGuidance() string {
	if strings.EqualFold(pc.RiskScale, "granular") {
		return "RISK SCORING (1-10 granular scale)."
	}
	return "RISK SCORING (binary scale): Low/Medium/High."
}

func (pc *PromptConfig) BuildPrompt(transcript string) string {
	if pc.Assertiveness < 1 || pc.Assertiveness > 10 {
		pc.Assertiveness = 8
	}
	if pc.Language == "" {
		pc.Language = "Indonesian"
	}

	// Structural Requirements
	structural := fmt.Sprintf(`### STRUCTURAL REQUIREMENTS
1. **Language**: All output in %s.
2. **Analysis Over Description**: Not just "what was said" but "what it means".
3. **Formatting**: Markdown headers, tables, status indicators.
4. **Sequential Numbering**: Number all primary sections sequentially (1, 2, 3...) in the final output. If a section is omitted, the next section must continue the correct sequence without gaps. Sub-sections must follow the parent number (e.g., 2.1, 2.2).

### ANTI-HALLUCINATION & DYNAMIC UX RULES (CRITICAL)
1. **DO NOT INVENT NAMES**: If a person's name, role, or specific assignment is not explicitly stated in the transcript, DO NOT make one up.
2. **OMIT EMPTY SECTIONS**: This is extremely important. If there is no information in the audio for a specific section (e.g., no Action Items, no Decisions, or no Risks), DO NOT write "Tidak ada", "N/A", or placeholders. **COMPLETELY REMOVE / OMIT that section and its header from your final output.**
3. **FLEXIBLE STRUCTURE**: The template below is just a guide. You are encouraged to merge topics, change section titles, or add new sections if it better fits the flow of the meeting. Keep sections that bring value based on the transcript.`, pc.Language)

	outputStructure := `### SUGGESTED OUTPUT STRUCTURE (FLEXIBLE: OMIT SECTIONS WITH NO CONTENT)

# AGENDA: [Inferred concise meeting context/objective]

# Background & Objective
Briefly describe the meeting context, objective, and overall goal as inferred from the transcript.

---

# Main Discussion
Provide a deep, multi-topic analysis of the conversations. Split this into logical sub-sections if there are different themes.
## [Section #].1 [Main Topic 1]
Detailed explanation of this topic, key points raised, and their significance.
### 📌 Breakdown (If applicable)
* Specific notes or sub-topics...

## [Section #].2 [Main Topic 2]
* Comprehensive bullet points analyzing the outcomes or disagreements...

---

# Roles & Responsibilities
Detail who is responsible for what, identifying roles explicitly mentioned.
| Party / Name | Responsibility |
|---|---|
| (Do not invent names) | ... |

---

# Action Items
| No | Task | PIC | Deadline | Status |
|---|---|---|---|---|
| 1 | description | name (do not invent) | date | status text |

---

# Decisions Made
1. List all key decisions agreed upon during the meeting.

---

# Open Issues / Pending Discussion
1. List topics that were discussed but not resolved.

---

# Risks & Mitigation
| Risk | Impact | Mitigation |
|---|---|---|

---

# Additional Notes
Other noteworthy observations.`

	prompt := fmt.Sprintf(`### ROLE & MANDATE
You are a Chief of Staff and Strategic Analyst. Extract strategic value, risks, and recommended actions from the provided transcript.
**CRITICAL**: Adapt the suggested structure dynamically. DO NOT keep bracketed placeholders like [Meeting Title] or [Name 1] in your output. Replace them with specific information found in the transcript or metadata. If information is missing, OMIT that section or specific line entirely.

%s

%s

%s

### TONE & EDITORIAL
%s

### CONTEXT
- Context: %s
- Style: %s (Format as MoM if style is 'minutes' or 'notulensi')

### METADATA (REFERENCE ONLY)
- Date: %s
- Location: %s
- Participants: %s

**IMPORTANT**: Use the metadata above to fill in the header sections. If metadata is empty or "[...]"(brackets), try to infer from the transcript. If still unknown, omit the field.

%s

### INPUT TRANSCRIPT
<transcript>
%s
</transcript>

**BEGIN OUTPUT (STRIP ALL PLACEHOLDERS, OMIT EMPTY SECTIONS):**
`, pc.AudienceGuidance(), pc.RiskScoringGuidance(), structural, pc.AssertivenessPhrasing(), pc.Context, pc.Style, pc.Date, pc.Location, pc.Participants, outputStructure, transcript)

	return prompt
}

// BuildMinutesPrompt builds a meeting-minutes style prompt (default for mobile flow)
// This is less abstract than the strategic analyst prompt and focuses on factual recording
func (pc *PromptConfig) BuildMinutesPrompt(transcript string) string {
	if pc.Language == "" {
		pc.Language = "Indonesian"
	}

	targetLangName := pc.Language

	return fmt.Sprintf(`### ROLE & MANDATE
You are a professional meeting secretary creating official meeting minutes.

**CRITICAL RULES**:
1. **DO NOT INVENT INFORMATION**: If a name, role, decision, or action item is not explicitly stated in the transcript, DO NOT make it up.
2. **OMIT EMPTY SECTIONS**: If there are no decisions, action items, or risks in the transcript, COMPLETELY REMOVE those sections from your output. Do NOT write "Tidak ada", "N/A", or placeholders.
3. **PRESERVE UNCERTAINTY**: If speakers express uncertainty ("might", "possibly", "we should discuss"), preserve these markers. Do not smooth them away.
4. **PRESERVE DISAGREEMENTS**: If there were unresolved disagreements, document them explicitly in "Open Issues".
5. **NO INFERRED OWNERSHIP**: Do not assign action items to people unless explicitly stated in the transcript.

### OUTPUT LANGUAGE
All output must be in %s.

### REQUIRED SECTIONS (omit if no content)
1. **Meeting Info**: Date, Location, Participants (from metadata or inferred from transcript)
2. **Agenda**: Brief statement of meeting purpose
3. **Decisions**: List of decisions made (only if explicitly stated)
4. **Action Items**: Table with Task, PIC (only if named), Deadline (only if stated), Status
5. **Open Issues**: Unresolved topics requiring future discussion
6. **Risks**: Identified risks with mitigation strategies (if discussed)

### FORMAT
Use clean Markdown formatting. Use tables for Action Items and Risks.

### METADATA
- Context: %s
- Style: %s
- Date: %s
- Location: %s
- Participants: %s

### TRANSCRIPT:
%s

**REMINDER**: Create meeting minutes following the structure above. Omit any section that has no content from the transcript.`, targetLangName, pc.Context, pc.Style, pc.Date, pc.Location, pc.Participants, transcript)
}

// BuildExecutiveMinutesPrompt builds an executive-summary style prompt for C-level audiences
func (pc *PromptConfig) BuildExecutiveMinutesPrompt(transcript string) string {
	if pc.Language == "" {
		pc.Language = "Indonesian"
	}

	return fmt.Sprintf(`### ROLE & MANDATE
You are creating executive meeting minutes for C-level leadership.

**CRITICAL**:
- **BREVITY**: Executive summary style - concise, actionable, strategic focus
- **NO INVENTION**: Do not invent names, decisions, or ownership
- **OMIT EMPTY SECTIONS**: Remove sections with no content entirely
- **ROI MINDSET**: Emphasize business impact, decisions, and strategic actions

### OUTPUT LANGUAGE
All output in %s.

### REQUIRED STRUCTURE
1. **Executive Summary**: 2-3 sentence overview of meeting purpose and key outcomes
2. **Key Decisions**: Bullet list of strategic decisions (only if explicitly made)
3. **Strategic Actions**: High-level action items with business impact (PIC only if named)
4. **Escalations/Blockers**: Issues requiring leadership attention
5. **Risks**: Strategic risks with business impact (if discussed)

### METADATA
- Context: %s
- Date: %s
- Location: %s
- Participants: %s

### TRANSCRIPT:
%s

**OUTPUT**: Executive meeting minutes following the structure above.`, pc.Language, pc.Context, pc.Date, pc.Location, pc.Participants, transcript)
}

// BuildActionItemsPrompt builds a prompt focused specifically on extracting action items
func (pc *PromptConfig) BuildActionItemsPrompt(transcript string) string {
	if pc.Language == "" {
		pc.Language = "Indonesian"
	}

	return fmt.Sprintf(`### ROLE & MANDATE
You are extracting action items and commitments from a meeting transcript.

**CRITICAL RULES**:
1. **EXPLICIT COMMITMENTS ONLY**: Only extract action items that were explicitly committed to ("I will", "we should", "let's do")
2. **NO INFERRED OWNERSHIP**: If PIC is not explicitly named, leave it blank or mark as "TBD"
3. **PRESERVE DEADLINES**: If a deadline was mentioned, include it. Otherwise leave blank.
4. **INCLUDE CONTEXT**: Briefly note why each action item matters

### OUTPUT LANGUAGE
All output in %s.

### OUTPUT FORMAT
Create a Markdown table with these columns:
| No | Action Item | PIC | Deadline | Context/Notes |

### METADATA
- Context: %s
- Date: %s

### TRANSCRIPT:
%s

**OUTPUT**: Extract all action items into the table format above. If no action items exist, output "No action items identified."`, pc.Language, pc.Context, pc.Date, transcript)
}

// GetStylePrompt returns the appropriate prompt based on style parameter
func (pc *PromptConfig) GetStylePrompt(transcript string) string {
	style := strings.ToLower(pc.Style)

	switch {
	case style == "executive" || style == "exec":
		return pc.BuildExecutiveMinutesPrompt(transcript)
	case style == "action_items" || style == "actions" || style == "tasks":
		return pc.BuildActionItemsPrompt(transcript)
	case style == "minutes" || style == "notulensi" || style == "standard":
		fallthrough
	default:
		return pc.BuildMinutesPrompt(transcript)
	}
}
