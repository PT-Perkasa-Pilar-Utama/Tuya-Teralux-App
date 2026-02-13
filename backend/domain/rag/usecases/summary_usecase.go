package usecases

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"time"

	"github.com/google/uuid"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func (u *RAGUsecase) summaryInternal(text string, language string, context string, style string) (*dtos.RAGSummaryResponseDTO, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is empty")
	}

	if language == "" {
		language = "id"
	}

	targetLangName := "Indonesian"
	if strings.ToLower(language) == "en" {
		targetLangName = "English"
	}

	prompt := fmt.Sprintf(`### ROLE
You are a Senior Project Management Officer and Strategic Analyst. Your goal is to convert raw meeting transcripts into professional meeting intelligence using a structured reporting framework.

### INSTRUCTIONS
1. *Language Focus*: Regardless of the transcript language, the report MUST be written entirely in %s.
2. *Denoise & Professionalize*: Remove filler words, stuttering, and informal speech. Convert the text into formal business English.
3. *PPP Framework*: Within organized Discussion Points, strictly follow the Progress/Issues/Plans structure.
4. *Objectivity*: Do NOT invent facts, deadlines, or owners. If specific data is missing, use a dash (-) or omit the field.
5. *Formatting*: Use Markdown formatting. Use bolding (#, ##, **bold**) for headers and key points. Use bullet points for lists.

### CONTENT CONTEXT
- Context: %s
- Desired Style: %s

### OUTPUT FORMAT
Executive Summary
(A concise, 3-sentence high-level overview of the meeting goals and outcomes)

Key Discussion Points
(Number). (Topic Name)
(High-level summary of this topic)
(Number.Number) (Sub-topic Name) - Reporter: [Speaker Name/Role]
• Progress: (Current status/what was achieved)
• Issues: (Specific obstacles or gaps mentioned)
• Plans: (Strategic next steps, otherwise -)

Decisions Made
(List confirmed outcomes. If none, state: "No explicit decisions reached during this session")

Action Items
(List specific tasks. Include Owner and Deadline ONLY if explicitly stated in the transcript.)

Open Questions
(Points of discussion left unresolved)

AI Strategic Analysis
(AI analysis of gaps. Identify ambiguities and suggest constructive next steps for the project lead.)

### CONSTRAINTS
- Use simple dot (•) for bullet points.
- Use double line breaks between major headers for readability.
- Maintain a high-level strategic perspective.

---
<transcript>
"%s"
</transcript>

Strategic Summary (%s):`, targetLangName, context, style, text, targetLangName)

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

	pdfUrl := fmt.Sprintf("/uploads/reports/%s", pdfFilename)

	utils.LogDebug("RAG Summary: language='%s', summary_len=%d, model='%s', pdf='%s'", language, len(trimmedSummary), model, pdfUrl)
	utils.LogDebug("RAG Summary Result: %q", trimmedSummary)

	return &dtos.RAGSummaryResponseDTO{
		Summary: trimmedSummary,
		PDFUrl:  pdfUrl,
	}, nil
}

func (u *RAGUsecase) Summary(text string, language string, context string, style string) (string, error) {
	taskID := uuid.New().String()
	u.mu.Lock()
	u.taskStatus[taskID] = &dtos.RAGStatusDTO{Status: "pending"}
	u.mu.Unlock()

	if u.badger != nil {
		b, _ := json.Marshal(u.taskStatus[taskID])
		_ = u.badger.Set("rag:task:"+taskID, b)
	}

	go func() {
		result, err := u.summaryInternal(text, language, context, style)
		u.mu.Lock()
		if err != nil {
			utils.LogError("RAG Summary Task %s: Failed with error: %v", taskID, err)
			u.taskStatus[taskID] = &dtos.RAGStatusDTO{Status: "failed", Result: err.Error()}
		} else {
			utils.LogInfo("RAG Summary Task %s: Completed successfully", taskID)
			// For summary, we might want to store the structured result as JSON in StatusDTO.Result
			// or just use ExecutionResult. For consistency with Control, let's use Result as string if possible,
			// or just store the whole DTO as ExecutionResult.
			u.taskStatus[taskID] = &dtos.RAGStatusDTO{
				Status:          "completed",
				ExecutionResult: result,
				Result:          result.Summary, // Fallback for simple consumers
			}
		}
		status := u.taskStatus[taskID]
		u.mu.Unlock()

		if u.badger != nil {
			b, _ := json.Marshal(status)
			_ = u.badger.SetPreserveTTL("rag:task:"+taskID, b)
		}
	}()

	return taskID, nil
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
	
	// Define known section headers (Focus on English-only as per strategy)
	headers := map[string]bool{
		"Executive Summary":     true,
		"Key Discussion Points": true,
		"Decisions Made":        true,
		"Action Items":          true,
		"Open Questions":        true,
		"AI Strategic Analysis": true,
		// Indonesian headers
		"Ringkasan Eksekutif":   true,
		"Poin Diskusi Utama":    true,
		"Keputusan Utama":       true,
		"Tindakan Lanjut":       true,
		"Pertanyaan Terbuka":    true,
		"Analisis Strategis AI": true,
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
