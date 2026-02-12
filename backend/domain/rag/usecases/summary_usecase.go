package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// Summary generates professional meeting minutes from the provided text using the LLM.
func (u *RAGUsecase) Summary(text string, language string, context string, style string) (*dtos.RAGSummaryResponseDTO, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is empty")
	}

	if language == "" {
		language = "id"
	}
	
	targetLang := "Indonesian"
	if strings.ToLower(language) == "en" {
		targetLang = "English"
	}

	prompt := fmt.Sprintf(`### ROLE
You are a Senior Project Management Officer and Strategic Analyst. Your goal is to convert raw meeting transcripts into professional meeting intelligence using a structured reporting framework.

### INSTRUCTIONS
1. *Language Mirroring*: Detect the transcript language. Headings and content MUST be in %s.
2. *Denoise & Professionalize*: Remove filler words and convert informal speech into formal business language.
3. *PPP Framework*: Within the Discussion Points, use the Progress/Masalah/Rencana (Progress/Issues/Plans) structure.
4. *No Hallucination*: Do NOT invent deadlines, owners, or facts. If a field like "Plans" or "Deadline" is not mentioned, use a dash (-) or omit it as per the format.
5. *Formatting*: Strictly NO Markdown (no #, *, or |). Use double line breaks between sections for readability.

### CONTEXT
%s

### STYLE
%s

### OUTPUT FORMAT
Executive Summary
(3-sentence high-level overview of the meeting)

Key Discussion Points
(Number). (Topic Name)
(High-level summary of this topic)
(Number.Number) (Sub-topic Name) - Pelapor: [Speaker Name/Role]
• Kemajuan: (Current status/what was discussed)
• Masalah: (Specific challenges or gaps mentioned)
• Rencana: (Explicit next steps mentioned in text, otherwise -)

Decisions Made
(Confirmed decisions or deferred items. If none, state: "Tidak ada keputusan yang secara eksplisit disepakati")

Action Items
(List each task. Include Owner and Deadline ONLY if explicitly mentioned. If not mentioned, simply list the task without inventing data.)

Open Questions
(Issues raised but not resolved)

Saran AI
(AI analysis of gaps. Identify items that lack clear follow-up and suggest a constructive next step to resolve the ambiguity.)

### CONSTRAINTS
- Use double line breaks between all major sections.
- Use a simple dot (•) for bullet points.
- Maintain strict objectivity in the main sections; keep suggestions only in the Saran AI section.

---
Transcript:
"%s"

Summary (%s):`, targetLang, context, style, text, targetLang)

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	summary, err := u.llm.CallModel(prompt, model)
	if err != nil {
		return nil, err
	}

	trimmedSummary := strings.TrimSpace(summary)

	// Generate PDF
	pdfFilename := fmt.Sprintf("summary_%d.pdf", time.Now().Unix())
	pdfPath := filepath.Join("uploads", "reports", pdfFilename)

	// Create reports directory if not exists
	os.MkdirAll(filepath.Dir(pdfPath), 0755)

	if err := u.generateProfessionalPDF(trimmedSummary, pdfPath); err != nil {
		utils.LogWarn("Warning: Failed to generate PDF: %v", err)
	}

	pdfUrl := fmt.Sprintf("/api/static/reports/%s", pdfFilename)

	utils.LogDebug("RAG Summary: language='%s', summary_len=%d, model='%s', pdf='%s'", language, len(trimmedSummary), model, pdfUrl)
	utils.LogDebug("RAG Summary Result: %q", trimmedSummary)

	return &dtos.RAGSummaryResponseDTO{
		Summary: trimmedSummary,
		PDFUrl:  pdfUrl,
	}, nil
}

func (u *RAGUsecase) generateProfessionalPDF(summary string, path string) error {
	cfg := config.NewBuilder().
		WithPageNumber().
		Build()

	m := maroto.New(cfg)

	// Header
	m.AddRows(
		row.New(20).Add(
			col.New(12).Add(
				text.New("MEETING SUMMARY REPORT", props.Text{
					Top:    5,
					Size:   20,
					Style:  fontstyle.Bold,
					Align:  align.Center,
					Color:  &props.Color{Red: 10, Green: 20, Blue: 100},
				}),
			),
		),
		row.New(10).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Generated on: %s", time.Now().Format("02 Jan 2006 15:04")), props.Text{
					Size:  10,
					Align: align.Right,
					Style: fontstyle.Italic,
				}),
			),
		),
	)

	// Content
	lines := strings.Split(summary, "\n")
	
	// Define known section headers (EN and ID)
	headers := map[string]bool{
		"Executive Summary":    true,
		"Ringkasan Eksekutif":  true,
		"Key Discussion Points": true,
		"Poin Diskusi Utama":   true,
		"Decisions Made":       true,
		"Keputusan":            true,
		"Action Items":         true,
		"Tindak Lanjut":        true,
		"Open Questions":       true,
		"Pertanyaan Terbuka":   true,
		"Saran AI":             true,
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			m.AddRows(row.New(5))
			continue
		}

		// Check if line is a section header (bold and larger font)
		isHeader := false
		for h := range headers {
			if strings.EqualFold(line, h) {
				isHeader = true
				break
			}
		}

		if isHeader {
			m.AddRows(row.New(10).Add(
				col.New(12).Add(
					text.New(strings.ToUpper(line), props.Text{
						Size:  14,
						Style: fontstyle.Bold,
						Top:   2,
						Color: &props.Color{Red: 50, Green: 50, Blue: 50},
					}),
				),
			))
		} else if strings.HasPrefix(line, "• ") || strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			content := line
			if strings.HasPrefix(line, "- ") {
				content = "• " + strings.TrimPrefix(line, "- ")
			} else if strings.HasPrefix(line, "* ") {
				content = "• " + strings.TrimPrefix(line, "* ")
			}
			m.AddRows(row.New(8).Add(
				col.New(12).Add(
					text.New(content, props.Text{
						Size: 11,
						Left: 5,
					}),
				),
			))
		} else if (len(line) > 0 && line[0] >= '0' && line[0] <= '9') {
			// Lines starting with numbers (Task/Topic numbering) - also bolded or slightly emphasized
			m.AddRows(row.New(8).Add(
				col.New(12).Add(
					text.New(line, props.Text{
						Size:  11,
						Style: fontstyle.Bold,
					}),
				),
			))
		} else {
			m.AddRows(row.New(8).Add(
				col.New(12).Add(
					text.New(line, props.Text{
						Size: 11,
					}),
				),
			))
		}
	}

	doc, err := m.Generate()
	if err != nil {
		return err
	}

	return doc.Save(path)
}
