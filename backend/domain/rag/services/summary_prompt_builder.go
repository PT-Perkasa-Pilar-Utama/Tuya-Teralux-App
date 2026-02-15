package services

import (
	"fmt"
	"strings"
)

// PromptConfig encapsulates parameterized prompt generation for meeting summaries.
// Controls tone, depth, and output structure independently.
type PromptConfig struct {
	// Assertiveness: 1-10 scale
	// 1-3: neutral, safe, highly hedged
	// 4-6: balanced, some conviction
	// 7-9: assertive, willing to call out gaps/risks
	// 10: aggressive, blunt (rarely recommended)
	Assertiveness int

	// Audience level affects depth & terminology
	// "c-level": CEO, CFO, CTO — strategic brevity, ROI focus
	// "mixed": C-level + VPs/Directors — slightly more detail, cross-org context
	// "execution": team leads, individual contributors — tactical detail
	Audience string

	// RiskScale defines how AI scores severity
	// "binary": High/Medium/Low only
	// "granular": 1-10 scale (1-3 low, 4-6 medium, 7-10 high)
	RiskScale string

	// Context provided by user (e.g., "product roadmap planning", "hiring review")
	Context string

	// Style provided by user (e.g., "executive brief", "tactical breakdown")
	Style string

	// Language for output ("Indonesian", "English")
	Language string
}

// AssertivenessPhrasing returns prompt guidance based on assertiveness level
func (pc *PromptConfig) AssertivenessPhrasing() string {
	switch {
	case pc.Assertiveness <= 3:
		return `
Tone: Objective and cautious. Present findings with qualifier phrases like "suggests", "may indicate", "worth exploring". 
Avoid making strong claims. Emphasize what is known and what requires further investigation.
Editorial voice: Clinical, neutral, non-opinionated.`

	case pc.Assertiveness <= 6:
		return `
Tone: Balanced and analytical. Present findings with confidence when evidence supports it, but include caveats where appropriate.
Use phrases like "indicates", "demonstrates", "points to" for supported claims.
Editorial voice: Professional analyst offering well-reasoned perspective.`

	default: // 7-10
		return `
Tone: Assertive and analytical. Call out gaps, risks, and misalignments directly. Be willing to name problems.
Use phrases like "reveals", "exposes", "highlights critical gap", "strategic risk", "stagnation risk".
Editorial voice: Chief of Staff analyzing strategic health; willing to deliver hard truths when evidence supports it.
Do NOT hedge legitimate concerns. If decision clarity is absent, say so explicitly with impact projection.`
	}
}

// AudienceGuidance returns prompt guidance based on target audience
func (pc *PromptConfig) AudienceGuidance() string {
	switch strings.ToLower(pc.Audience) {
	case "c-level":
		return `
Audience: C-suite executives (CEO, CFO, CTO). 
Depth: Strategic only. Skip tactical details unless directly relevant to strategic decision-making.
Focus: Business impact, competitive implications, resource allocation, timeline risk, strategic alignment.
Length: Ultra-condensed. Assume reader has limited attention; every word must earn its place.
Terminology: Business outcomes, not implementation details.`

	case "mixed":
		return `
Audience: C-level + VPs/Directors (product, engineering, sales, ops, finance).
Depth: Strategic + some operational context. Explain cross-org implications.
Focus: Business impact, resource allocation, timeline, strategic alignment, plus tactical feasibility.
Length: Moderately condensed. More space for cross-functional implications.
Terminology: Mix of strategic outcomes and operational feasibility.`

	default: // "execution"
		return `
Audience: Execution team (team leads, ICs, project managers).
Depth: Tactical and strategic. Include decision rationale and implementation implications.
Focus: What was decided, why it matters to execution, what changes, what stays same, timeline impact.
Length: Detailed but organized clearly. More space for "so what does this mean for us" specificity.
Terminology: Technical but business-aware.`
	}
}

// RiskScoringGuidance returns prompt guidance for risk assessment scoring
func (pc *PromptConfig) RiskScoringGuidance() string {
	if strings.ToLower(pc.RiskScale) == "granular" {
		return `
RISK SCORING (1-10 granular scale):
1-3: Low risk. Easily managed, low probability or low impact. Can defer without consequence.
4-6: Medium risk. Noticeable impact or moderate probability. Requires active management within 30 days.
7-10: High risk. Severe impact or high probability. Requires immediate action (within 7 days for 9-10, within 14 days for 7-8).

For each risk, provide:
- Risk description
- Likelihood score (1-10)
- Impact score (1-10)  
- Composite risk score = AVERAGE(likelihood, impact) — round to nearest integer
- Mitigation strategy`
	}

	return `
RISK SCORING (binary scale):
Low: Easily managed, low probability/impact.
Medium: Noticeable impact, moderate probability.
High: Severe impact or high probability, needs immediate action.

For each risk, provide: description, score, impact statement, mitigation.`
}

// BuildPrompt generates the full prompt string with structured sections
func (pc *PromptConfig) BuildPrompt(transcript string) string {
	if pc.assertivenessValid() {
		// Apply defaults
		pc.Assertiveness = 8
	}
	if pc.Audience == "" {
		pc.Audience = "mixed"
	}
	if pc.RiskScale == "" {
		pc.RiskScale = "granular"
	}
	if pc.Language == "" {
		pc.Language = "Indonesian"
	}

	prompt := fmt.Sprintf(`### ROLE & MANDATE
You are a Chief of Staff and Strategic Analyst producing a meeting intelligence report.
Your mandate: Extract strategic value, call out risks and gaps, and recommend concrete actions.

%s

%s

%s

---

### STRUCTURAL REQUIREMENTS

1. **Language**: All output in %s, regardless of transcript language. Convert colloquialisms to formal equivalents.

2. **Evidence Traceability**: Every key finding must anchor to something explicitly said in the meeting.
   If unsure, write "(inference)" or "Not specified." Do NOT invent details.

3. **Specificity Over Generic**: 
   - Use concrete names, team abbreviations, and numbers when present
   - Avoid vague terms like "optimization", "improvement", "enhancement"
   - Replace with specific mechanics: what changes, by how much, with what timeline

4. **Analysis Over Description**:
   - Not just "what was said" but "what it means"
   - Include impact, risk, and recommended response
   - Connect dots across topics

5. **Formatting Rules**:
    - Markdown headers: # for main, ## for subsection, ### for detail. ALWAYS use a single relevant emoji in the main (#) headers.
    - Lists: Use bullet (*) for unordered, numbered for sequences.
    - Tables: Use when comparing 3+ attributes (clarity, ownership, risk, etc).
    - Dividers: Use "---" between major sections.
    - Emphasis: Use **bold** for key metrics/decisions; use _italic_ for caveats.
    - Status Indicators: Use 🔴 (Red) for At Risk/Blocked, 🟡 (Yellow) for Partial/Caution, 🟢 (Green) for Healthy/Strategic.
    - Keep nesting shallow (max 1-2 levels to avoid visual clutter).

---

### OUTPUT STRUCTURE

---

# 🧭 Executive Signals
Provide a high-level strategic health dashboard:
- **Strategic Status**: [🔴/🟡/🟢 Indicator] **[Status Name]** — [One sentence rationale]
- **Decision Density**: [⚪ Indicator] **[High/Medium/Low]** — [One sentence rationale]
- **Execution Readiness**: **[1-10 score]** — [Short explanation of owners/timelines]
- **Alignment Score**: **[1-10 score]** — [Short explanation of stakeholder agreement]
- **30-Day Risk Outlook**: [⚠️/✅ Indicator] **[Level]** — [Short explanation]

---

# 📌 Executive Summary
REQUIRED: 3-4 assertive bullet points summarizing the "So what?"
### Core Outcome
What was actually settled?
### Strategic Impact
How does this change the roadmap, status quo, or bottom line?
### Urgency Signal
What happens if we don't act in the next 7 days?

---

# 🧠 Strategic Interpretation (The "AI Brain" Layer)
Analyze the subtext and blind spots:
### What Was Not Said
Missing prerequisites, ignored risks, or silent stakeholders.
### Tension Points
Conflicting priorities or underlying disagreements identified.
### Blind Spots
Areas assumed to be fine but lacking evidence.

---

# 🔎 Key Discussion Themes
For each major topic (max 3-4):

## [Topic Name]
**Digest**
Core debate/update

**Status**
[🔴/🟡/🟢 Status Name] — [One sentence rationale]

**Strategic Implication / Risk Vector**
Specific cost of delay or impact on budget, talent, or market positioning.

---

# 📋 Decisions Made
Format: | Decision | Owner | Impact |

---

# 🛠 Analyst Recommended Actions
Based on gaps identified, what SHOULD be assigned immediately:
Format: | Recommended Action | Suggested Owner | Strategic Rationale |

---

# ⚠️ Risk Matrix
| Risk | Likelihood | Impact | Commentary |

---

# ❓ Open Questions
| Question | Why it matters | Recommended Owner |

---

# 🎯 Strategic Commentary
Final synthesis for leadership:
- **Overall Assessment**: [One sentence]
- **Resource Constraints**: [One sentence]
- **The One Thing**: [The single most important takeaway for C-level]

---

### TONE & EDITORIAL GUIDELINES

%s

**Constraints:**
- Do NOT use hedging phrases reflexively (e.g., "It seems", "Arguably", "One could argue")
- Do NOT over-explain obvious items
- Do NOT assume reader saw the meeting; be self-contained
- Do NOT ignore ambiguity; call it out as "Not specified" or "(clarity gap)"
- Do NOT pad with generic corporate language

---

### CONTEXT & FOCUS
- Meeting Context: %s
- Desired Style: %s

---

### INPUT TRANSCRIPT
<transcript>
%s
</transcript>

---

**BEGIN OUTPUT:**
`, pc.AssertivenessPhrasing(), pc.AudienceGuidance(), pc.RiskScoringGuidance(), pc.Language,
		pc.AssertivenessPhrasing(),
		pc.Context, pc.Style, transcript)

	return prompt
}

// Helper: validate assertiveness
func (pc *PromptConfig) assertivenessValid() bool {
	return pc.Assertiveness < 1 || pc.Assertiveness > 10
}

// Helper: risk score label for output structure
func (pc *PromptConfig) riskScoreLabel() string {
	if strings.ToLower(pc.RiskScale) == "granular" {
		return "Score (1-10)"
	}
	return "Level (Low/Med/High)"
}
