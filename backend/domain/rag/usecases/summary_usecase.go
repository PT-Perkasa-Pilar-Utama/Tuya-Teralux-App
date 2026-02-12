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
	
	targetLangName := "Indonesian"
	if strings.ToLower(language) == "en" {
		targetLangName = "English"
	}

	prompt := fmt.Sprintf(`You are an expert Executive Assistant specializing in distilling the essence of meetings and conversations into professional summaries.
Your goal is to extract CONCLUSIONS, DECISIONS, and KEY TAKEAWAYS. 
STRICT RULE: DO NOT retell the story chronologically. DO NOT menceritakan ulang isi teks secara naratif. Focus on what was achieved or decided.

STRICT RULES:
1. BE CONCISE. Use bullet points for details.
2. Remove all filler words (e.g., "uh", "um", "like", "so").
3. OUTPUT LANGUAGE: You must write the summary in %s even if the input is in a different language.

CONTEXT: %s
STYLE: %s

OUTPUT STRUCTURE:
# [Meeting Title / Topik Utama]

## 1. Summary (Ringkasan)
Summarize the essence of the conversation in 1-2 concise sentences.

## 2. Key Points (Poin Penting)
- **[Topic A]**: Main conclusion or result regarding this topic.
- **[Topic B]**: ...

## 3. Decisions (Keputusan)
- [Decision 1]
- [Decision 2]
(If none, state "No specific decisions recorded")

## 4. Action Items (Tindak Lanjut)
- [ ] **[PIC/Owner]**: [Task Description] (Deadline if any)

Text: "%s"
Summary (%s):`, targetLangName, context, style, text, targetLangName)

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
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			m.AddRows(row.New(5))
			continue
		}

		if strings.HasPrefix(line, "# ") {
			m.AddRows(row.New(12).Add(
				col.New(12).Add(
					text.New(strings.TrimPrefix(line, "# "), props.Text{
						Size:  16,
						Style: fontstyle.Bold,
						Color: &props.Color{Red: 50, Green: 50, Blue: 50},
					}),
				),
			))
		} else if strings.HasPrefix(line, "## ") {
			m.AddRows(row.New(10).Add(
				col.New(12).Add(
					text.New(strings.TrimPrefix(line, "## "), props.Text{
						Size:  14,
						Style: fontstyle.Bold,
						Top:   2,
					}),
				),
			))
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			content := line
			if strings.HasPrefix(line, "- ") {
				content = "• " + strings.TrimPrefix(line, "- ")
			} else {
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
