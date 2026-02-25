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

	// Output Structure based on User Template
	var outputStructure string
	if strings.Contains(strings.ToLower(pc.Language), "english") {
		outputStructure = `### SUGGESTED OUTPUT STRUCTURE (FLEXIBLE: OMIT SECTIONS WITH NO CONTENT)

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
	} else {
		outputStructure = `### SUGGESTED OUTPUT STRUCTURE (FLEXIBLE: OMIT SECTIONS WITH NO CONTENT)

# AGENDA: [Ringkasan singkat konteks/tujuan rapat yang disimpulkan]

# Latar Belakang & Tujuan
Deskripsi singkat mengenai konteks, tujuan, dan sasaran utama rapat berdasarkan transkrip.

---

# Pembahasan Utama
Berikan analisis mendalam untuk setiap topik besar. Pecah menjadi sub-bagian logis jika ada tema yang berbeda.
## [Section #].1 [Topik Besar 1]
Penjelasan mendalam mengenai topik ini, poin-poin kunci yang diangkat, dan signifikansinya.
### 📌 Sub-Acara / Breakdown (Jika ada)
* Agenda & Catatan spesifik...

## [Section #].2 [Topik Besar 2]
* Analisis komprehensif mengenai hasil diskusi atau perdebatan...

---

# Pembagian Peran & Tanggung Jawab
Detailkan siapa yang bertanggung jawab atas apa, gunakan nama yang disebutkan.
| Pihak / Nama | Tanggung Jawab |
|---|---|
| (Jangan mengarang nama) | ... |

---

# Rencana Kerja (Action Items)
| No | Tugas | PIC | Deadline | Status |
|---|---|---|---|---|
| 1 | deskripsi tugas | nama (jangan mengarang) | tanggal | status teks |

---

# Keputusan yang Disepakati
1. Daftar semua keputusan penting yang disepakati selama rapat.

---

# Isu Terbuka / Pending Discussion
1. Daftar topik yang dibahas namun belum mencapai keputusan akhir.

---

# Risiko & Mitigasi
| Risiko | Dampak | Mitigasi |
|---|---|---|

---

# Catatan Tambahan
Hal-hal penting lainnya yang perlu dicatat.`
	}

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
