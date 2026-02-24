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
3. **Formatting**: Markdown headers, tables, status indicators (☐/⏳/✅ for actions, 🟢🟡🔴 for risks).

### ANTI-HALLUCINATION & DYNAMIC UX RULES (CRITICAL)
1. **DO NOT INVENT NAMES**: If a person's name, role, or specific assignment is not explicitly stated in the transcript, DO NOT make one up.
2. **OMIT EMPTY SECTIONS**: If there is no information in the audio for a specific section (e.g., no Action Items, no Decisions, or no Risks), **DO NOT write "Tidak ada" or placeholders. Completely remove/omit that section and its header from your final output.**
3. **FLEXIBLE STRUCTURE**: The template below is your maximum structure. You must adapt it to the actual conversation. Keep only the sections that have real data.`, pc.Language)

	// Output Structure based on User Template
	var outputStructure string
	if strings.Contains(strings.ToLower(pc.Language), "english") {
		outputStructure = `### OUTPUT STRUCTURE (MANDATORY FORMAT)

# 📋 MEETING SUMMARY TEMPLATE

## [Meeting Code/Date]: [Meeting Title]

**Date & Time:** [YYYY-MM-DD HH:MM] (Use ` + pc.Date + ` if available)
**Location:** [Meeting Location] (Use ` + pc.Location + ` if available)
**Participants:**
* [Name 1]
* [Name 2]
 (Use ` + pc.Participants + ` as reference. If nobody is mentioned and reference is empty, write "Not Mentioned")

---

# 1️⃣ Background & Objective
Brief description of the meeting context and objective.

---

# 2️⃣ Main Discussion
## 2.1 [Main Topic 1]
Brief explanation of this topic.
### 📌 Breakdown (If there are sessions/days)
#### Day One – [Date]
* Agenda & Notes...

## 2.2 [Main Topic 2]
* Bullet points...

---

# 3️⃣ Roles & Responsibilities
| Party / Name | Responsibility |
|---|---|
| (Do not invent names) | ... |

---

# 4️⃣ Action Items
| No | Task | PIC | Deadline | Status |
|---|---|---|---|---|
| 1 | description | name (do not invent) | date | ☐/⏳/✅ |

---

# 5️⃣ Decisions Made
1. List decisions...

---

# 6️⃣ Open Issues / Pending Discussion
1. List open issues...

---

# 7️⃣ Risks & Mitigation
| Risk | Impact | Mitigation |
|---|---|---|

---

# 8️⃣ Additional Notes
Other things to note.`
	} else {
		outputStructure = `### OUTPUT STRUCTURE (MANDATORY FORMAT)

# 📋 TEMPLATE SUMMARY MEETING

## [Kode/Tanggal] Rapat: [Judul Rapat]

**Tanggal & Waktu:** [YYYY-MM-DD HH:MM] (Use ` + pc.Date + ` if available)
**Lokasi:** [Lokasi Rapat] (Use ` + pc.Location + ` if available)
**Peserta:**
* [Nama 1]
* [Nama 2]
 (Use ` + pc.Participants + ` as reference. If nobody is mentioned and reference is empty, write "Tidak Disebutkan")

---

# 1️⃣ Latar Belakang & Tujuan
Deskripsi singkat mengenai konteks dan tujuan rapat.

---

# 2️⃣ Pembahasan Utama
## 2.1 [Topik Besar 1]
Penjelasan ringkas mengenai topik ini.
### 📌 Sub-Acara / Breakdown (Jika ada breakdown sesi/hari)
#### Hari Pertama – [Tanggal]
* Agenda & Catatan...

## 2.2 [Topik Besar 2]
* Bullet points...

---

# 3️⃣ Pembagian Peran & Tanggung Jawab
| Pihak / Nama | Tanggung Jawab |
|---|---|
| (Do not invent names) | ... |

---

# 4️⃣ Rencana Kerja (Action Items)
| No | Tugas | PIC | Deadline | Status |
|---|---|---|---|---|
| 1 | description | name (do not invent) | date | ☐/⏳/✅ |

---

# 5️⃣ Keputusan yang Disepakati
1. List keputusan...

---

# 6️⃣ Isu Terbuka / Pending Discussion
1. List isu terbuka...

---

# 7️⃣ Risiko & Mitigasi
| Risiko | Dampak | Mitigasi |
|---|---|---|

---

# 8️⃣ Catatan Tambahan
Hal lain yang perlu diperhatikan.`
	}

	prompt := fmt.Sprintf(`### ROLE & MANDATE
You are a Chief of Staff and Strategic Analyst. Extract strategic value, risks, and recommended actions.
Follow the MANDATORY FORMAT strictly.

%s

%s

%s

### TONE & EDITORIAL
%s

### CONTEXT
- Context: %s
- Style: %s (Format as MoM if style is 'minutes' or 'notulensi')

%s

### INPUT TRANSCRIPT
<transcript>
%s
</transcript>

**BEGIN OUTPUT:**
`, pc.AudienceGuidance(), pc.RiskScoringGuidance(), structural, pc.AssertivenessPhrasing(), pc.Context, pc.Style, outputStructure, transcript)

	return prompt
}
