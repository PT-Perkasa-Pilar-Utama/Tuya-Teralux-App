package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/johnfercher/maroto/v2/pkg/repository"
)

type SummaryPDFMeta struct {
	Language string
	Context  string
	Style    string
}

type SummaryPDFRenderer interface {
	Render(summary string, path string, meta SummaryPDFMeta) error
}

type MarotoSummaryPDFRenderer struct{}

func NewMarotoSummaryPDFRenderer() *MarotoSummaryPDFRenderer {
	return &MarotoSummaryPDFRenderer{}
}

func (r *MarotoSummaryPDFRenderer) Render(summary string, path string, meta SummaryPDFMeta) error {
	basePath, _ := os.Getwd()
	customFonts, err := repository.New().
		AddUTF8Font("notoserif", fontstyle.Normal, "/usr/share/fonts/noto/NotoSerif-Regular.ttf").
		AddUTF8Font("notoserif", fontstyle.Bold, "/usr/share/fonts/noto/NotoSerif-Bold.ttf").
		AddUTF8Font("notoserif", fontstyle.Italic, "/usr/share/fonts/noto/NotoSerif-Italic.ttf").
		AddUTF8Font("notoserif", fontstyle.BoldItalic, "/usr/share/fonts/noto/NotoSerif-BoldItalic.ttf").
		Load()
	if err != nil {
		return fmt.Errorf("failed to load custom fonts: %w", err)
	}

	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(15).
		WithCustomFonts(customFonts).
		WithDefaultFont(&props.Font{
			Family: "notoserif",
			Size:   13,
		}).
		Build()

	m := maroto.New(cfg)

	// Enhanced color palette
	brandDark := &props.Color{Red: 7, Green: 89, Blue: 133}  // Deep teal
	brandMid := &props.Color{Red: 12, Green: 130, Blue: 156} // Teal
	textDark := &props.Color{Red: 40, Green: 44, Blue: 52}
	textMuted := &props.Color{Red: 99, Green: 109, Blue: 122}

	// 1. BRANDED HEADER
	logoPath := filepath.Join(basePath, "assets/images/logo.png")
	m.AddRows(
		row.New(25).Add(
			col.New(2).Add(
				image.NewFromFile(logoPath, props.Rect{
					Center:  true,
					Percent: 80,
					Top:     2,
				}),
			),
			col.New(10).Add(
				text.New("MEETING INTELLIGENCE REPORT", props.Text{
					Size:   16,
					Style:  fontstyle.Bold,
					Family: "notoserif",
					Align:  align.Left,
					Color:  brandDark,
					Top:    8,
				}),
			),
		),
	)

	m.AddRows(row.New(2).Add(col.New(12).Add(text.New(" ", props.Text{}))).WithStyle(&props.Cell{BorderType: border.Bottom, BorderColor: brandMid}))
	m.AddRows(row.New(4))

	// 2. METADATA ROW
	reportDate := time.Now().Format("January 02, 2006")
	metaText := fmt.Sprintf("Report Date: %s | Language: %s", reportDate, meta.Language)
	if strings.TrimSpace(meta.Context) != "" {
		metaText = fmt.Sprintf("%s | Context: %s", metaText, meta.Context)
	}

	m.AddRows(
		row.New(6).Add(
			col.New(12).Add(
				text.New(metaText, props.Text{
					Size:   9,
					Family: "notoserif",
					Style:  fontstyle.Italic,
					Color:  textMuted,
					Align:  align.Left,
				}),
			),
		),
	)

	m.AddRows(row.New(5))

	// 3. Parse summary into sections
	lines := normalizeLines(summary)

	// 4. PROCESS LINES WITH CONTEXT-AWARE STYLING
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			m.AddRows(row.New(4))
			continue
		}

		// Detect section switches
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "executive signals") {
			addSectionHeader(m, "EXECUTIVE SIGNALS", brandDark)
			continue
		} else if strings.Contains(lowerLine, "executive summary") {
			addSectionHeader(m, "EXECUTIVE SUMMARY", brandDark)
			continue
		} else if strings.Contains(lowerLine, "strategic interpretation") {
			addSectionHeader(m, "STRATEGIC INTERPRETATION", brandDark)
			continue
		} else if strings.Contains(lowerLine, "risk assessment") {
			addSectionHeader(m, "RISK ASSESSMENT", brandDark)
			continue
		} else if strings.Contains(lowerLine, "strategic commentary") {
			addSectionHeader(m, "STRATEGIC COMMENTARY", brandDark)
			continue
		} else // Detect dividers
		if line == "---" {
			m.AddRows(row.New(1).Add(col.New(12).Add(text.New(" ", props.Text{}))).WithStyle(&props.Cell{BorderType: border.Bottom, BorderColor: textMuted}))
			m.AddRows(row.New(4))
			continue
		}

		if strings.HasPrefix(line, "# ") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			addSectionHeader(m, strings.ToUpper(cleanText(title)), brandDark)
			continue
		}

		// Body text processing
		fontSize := 13.0
		family := "notoserif"

		// Header detection (## or ###)
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") {
			section := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "### "), "## "))
			m.AddRows(row.New(10).Add(col.New(12).Add(text.New(cleanText(section), props.Text{
				Size:   fontSize + 1,
				Style:  fontstyle.Bold,
				Family: "notoserif",
				Color:  brandDark,
				Top:    2,
			}))))
			continue
		}

		// Tables are now treated as plain text list/points per user request
		if isTableRow(line) {
			if strings.Contains(line, ":---") || strings.Contains(line, "| --- |") {
				continue
			}
			renderTableAsText(m, line, fontSize, family, textDark)
			continue
		}

		// Bullet point detection
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "• ") || strings.HasPrefix(line, "* ") {
			content, level, _ := normalizeBullet(line)
			isBold := strings.Contains(content, "**")
			cleaned := cleanText(content)

			indent := float64(level) * 6.0
			rowHeight := estimateRowHeight(cleaned, 90-indent, fontSize)

			textProps := props.Text{
				Size:   fontSize,
				Family: family,
				Color:  textDark,
				Left:   indent,
				Top:    1,
			}
			if isBold {
				textProps.Style = fontstyle.Bold
			}

			m.AddRows(row.New(rowHeight).Add(
				col.New(1).Add(text.New("•", props.Text{Size: fontSize, Family: family, Color: brandMid, Align: align.Center})),
				col.New(11).Add(text.New(cleaned, textProps)),
			))
			continue
		}

		// Numbered items
		if len(line) > 0 && (line[0] >= '0' && line[0] <= '9') && strings.Contains(line, ".") {
			cleaned := cleanText(line)
			rowHeight := estimateRowHeight(cleaned, 100, fontSize)
			m.AddRows(row.New(rowHeight).Add(
				col.New(12).Add(
					text.New(cleaned, props.Text{
						Size:   fontSize,
						Style:  fontstyle.Bold,
						Family: family,
						Color:  textDark,
						Top:    1,
					}),
				),
			))
			continue
		}

		// Regular paragraph text
		cleaned := cleanText(line)
		rowHeight := estimateRowHeight(cleaned, 100, fontSize)

		m.AddRows(row.New(rowHeight).Add(
			col.New(12).Add(text.New(cleaned, props.Text{
				Size:   fontSize,
				Family: family,
				Color:  textDark,
				Align:  align.Left,
				Top:    1,
			})),
		))
	}

	// 5. FOOTER
	m.AddRows(row.New(4))
	m.AddRows(
		row.New(6).Add(
			col.New(12).Add(
				text.New("Generated by Teralux Meeting Intelligence Engine", props.Text{
					Size:  7,
					Color: textMuted,
					Align: align.Center,
					Style: fontstyle.Italic,
				}),
			),
		),
	)

	doc, err := m.Generate()
	if err != nil {
		return err
	}

	return doc.Save(path)
}

func normalizeBullet(line string) (string, int, bool) {
	clean := strings.TrimLeft(line, " \t")
	level := 0
	for {
		switch {
		case strings.HasPrefix(clean, "•"):
			clean = strings.TrimLeft(strings.TrimPrefix(clean, "•"), " \t")
			level++
		case strings.HasPrefix(clean, "-"):
			clean = strings.TrimLeft(strings.TrimPrefix(clean, "-"), " \t")
			level++
		case strings.HasPrefix(clean, "*"):
			clean = strings.TrimLeft(strings.TrimPrefix(clean, "*"), " \t")
			level++
		default:
			return clean, level, level > 0
		}
	}
}

func normalizeLines(raw string) []string {
	items := strings.Split(raw, "\n")
	out := make([]string, 0, len(items))
	for _, line := range items {
		line = strings.TrimSpace(line)
		if line == "" {
			out = append(out, "")
			continue
		}
		for strings.HasPrefix(line, "• •") || strings.HasPrefix(line, "- -") || strings.HasPrefix(line, "* *") {
			line = strings.ReplaceAll(line, "• •", "•")
			line = strings.ReplaceAll(line, "- -", "-")
			line = strings.ReplaceAll(line, "* *", "*")
			line = strings.TrimSpace(line)
		}
		out = append(out, line)
	}
	return out
}

// Helper: Add section header to maroto
func addSectionHeader(m core.Maroto, title string, textColor *props.Color) {
	m.AddRows(row.New(10).Add(
		col.New(12).Add(text.New(title, props.Text{
			Size:   14,
			Style:  fontstyle.Bold,
			Family: "notoserif",
			Color:  textColor,
			Top:    2,
		})),
	))
	m.AddRows(row.New(1).Add(col.New(12).Add(text.New(" ", props.Text{}))).WithStyle(&props.Cell{BorderType: border.Bottom, BorderColor: textColor}))
	m.AddRows(row.New(2))
}

// Helper: Render table as plain text points
func renderTableAsText(m core.Maroto, line string, fontSize float64, family string, textColor *props.Color) {
	cells := strings.Split(line, "|")
	output := []string{}
	for _, c := range cells {
		c = strings.TrimSpace(c)
		if c != "" && c != "-" && c != "–" {
			output = append(output, stripMarkdown(c))
		}
	}

	if len(output) == 0 {
		return
	}

	content := strings.Join(output, ": ")
	cleaned := cleanText(content)
	rowHeight := estimateRowHeight(cleaned, 100, fontSize)
	m.AddRows(row.New(rowHeight).Add(
		col.New(12).Add(text.New("• "+cleaned, props.Text{
			Size:   fontSize,
			Family: family,
			Color:  textColor,
			Top:    1,
		})),
	))
}

// Helper: Robust text cleaning for PDF stability
func cleanText(s string) string {
	// 1. Strip Markdown
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "#", "")

	// 2. Handle specific Executive Indicators
	s = strings.ReplaceAll(s, "🔴", "[!!!]")
	s = strings.ReplaceAll(s, "🟡", "[!]")
	s = strings.ReplaceAll(s, "🟢", "[OK]")
	s = strings.ReplaceAll(s, "⚪", "[?]")
	s = strings.ReplaceAll(s, "⚠️", "[WARN]")
	s = strings.ReplaceAll(s, "✅", "[DONE]")

	// 3. Handle problematic whitespace/symbols
	s = strings.ReplaceAll(s, "\u202f", " ") // Narrow No-Break Space
	s = strings.ReplaceAll(s, "≈", "~")      // Almost Equal

	// 4. Strip ALL non-BMP characters (Codepoints > 0xFFFF)
	// This is the CRITICAL fix for the CIDFontMap panic
	var b strings.Builder
	for _, r := range s {
		if r <= 0xFFFF {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func stripMarkdown(s string) string {
	return cleanText(s) // Use centralized cleaner
}

// Helper: Strip complex emojis for PDF stability (Deprecated, use cleanText)
func stripEmojis(s string) string {
	return cleanText(s)
}

// Helper: Estimate row height based on content
func estimateRowHeight(text string, widthChars float64, fontSize float64) float64 {
	if text == "" {
		return 2
	}
	// Rough heuristic: line length / width + padding
	lines := float64(len(text)) / widthChars
	if lines < 1 {
		lines = 1
	}
	// Add some buffer for word wrapping
	return (lines * fontSize * 0.8) + 4
}

// Helper: Detect if line is a table row
func isTableRow(line string) bool {
	return strings.Count(line, "|") >= 2
}
