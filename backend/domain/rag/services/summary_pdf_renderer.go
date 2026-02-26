package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type SummaryPDFMeta struct {
	Language     string
	Context      string
	Style        string
	Date         string
	Location     string
	Participants string
	CustomerName string
	CompanyName  string
}

type SummaryPDFRenderer interface {
	Render(summary string, path string, meta SummaryPDFMeta) error
}

type HTMLSummaryPDFRenderer struct{}

func NewHTMLSummaryPDFRenderer() *HTMLSummaryPDFRenderer {
	return &HTMLSummaryPDFRenderer{}
}

type templateData struct {
	SummaryPDFMeta
	SummaryHTML        template.HTML
	LogoDataURI        template.URL
	LblMeetingInfo     string
	LblDate            string
	LblLocation        string
	LblParticipants    string
	LblContext         string
	LblFooterRights    string
	LblFooterGenerated string
}

func (r *HTMLSummaryPDFRenderer) Render(summary string, pdfPath string, meta SummaryPDFMeta) error {
	basePath, _ := os.Getwd()

	// Convert Markdown to HTML
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM), // GitHub Flavored Markdown (tables, etc.)
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(summary), &buf); err != nil {
		return fmt.Errorf("failed to convert markdown to html: %w", err)
	}
	summaryHTML := template.HTML(buf.String())

	// Read and encode logo
	logoPath := filepath.Join(basePath, "assets/images/logo.png")
	logoBase64 := ""
	if imgData, err := os.ReadFile(logoPath); err == nil {
		logoBase64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(imgData)
	}

	// Prepare dynamic labels based on language
	isEnglish := strings.Contains(strings.ToLower(meta.Language), "english")
	data := templateData{
		SummaryPDFMeta: meta,
		SummaryHTML:    summaryHTML,
		LogoDataURI:    template.URL(logoBase64),
	}

	if isEnglish {
		data.LblMeetingInfo = "Meeting Information"
		data.LblDate = "Date"
		data.LblLocation = "Location"
		data.LblParticipants = "Participants"
		data.LblContext = "Agenda Context"
		data.LblFooterRights = "All rights reserved."
		data.LblFooterGenerated = "This document was automatically generated and summarized by an artificial intelligence system."
	} else {
		data.LblMeetingInfo = "Informasi Pertemuan"
		data.LblDate = "Tanggal"
		data.LblLocation = "Lokasi"
		data.LblParticipants = "Peserta"
		data.LblContext = "Konteks Agenda"
		data.LblFooterRights = "Seluruh hak cipta dilindungi undang-undang."
		data.LblFooterGenerated = "Dokumen ini dibuat secara otomatis oleh sistem kecerdasan buatan dan telah dirangkum untuk kemudahan analisis."
	}

	// Load templates
	tmplDir := filepath.Join(basePath, "templates", "pdf")
	t, err := template.ParseGlob(filepath.Join(tmplDir, "*.html"))
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	var htmlBuf bytes.Buffer
	if err := t.ExecuteTemplate(&htmlBuf, "mom_main.html", data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Generate PDF using go-rod
	err = generatePDFFromHTML(htmlBuf.String(), pdfPath)
	if err != nil {
		return fmt.Errorf("go-rod pdf generation failed: %w", err)
	}

	return nil
}

func generatePDFFromHTML(htmlContent string, outputPath string) error {
	// Create common browser launcher with flags
	l := rod.New().ControlURL("") // Local

	browser := l.MustConnect()
	defer browser.MustClose()

	// Use a 20s timeout for the entire operation
	return rod.Try(func() {
		page := browser.MustPage()
		defer page.MustClose()

		// Set the HTML content
		page.MustSetDocumentContent(htmlContent)
		page.MustWaitLoad()

		marginTop := 0.75
		marginBottom := 0.75
		marginLeft := 0.6
		marginRight := 0.6
		pdfStream, err := page.PDF(&proto.PagePrintToPDF{
			PrintBackground: true,
			MarginTop:       &marginTop,
			MarginBottom:    &marginBottom,
			MarginLeft:      &marginLeft,
			MarginRight:     &marginRight,
		})
		if err != nil {
			panic(err)
		}

		pdfBytes, err := io.ReadAll(pdfStream)
		if err != nil {
			panic(err)
		}

		_ = os.WriteFile(outputPath, pdfBytes, 0644)
	})
}
