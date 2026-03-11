package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"sensio/domain/common/utils"
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
	logoPath := utils.GetAssetPath("images/logo.png")
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
	tmplDir := utils.GetAssetPath("templates/pdf")
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
	// Create context with hard timeout
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Use explicit system chromium to avoid auto-download and ensure stability
	path := "/usr/bin/chromium"
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("system chromium not found at %s: PDF generation requires deterministic runtime: %w", path, err)
	}

	l := launcher.New().
		Bin(path).
		Set("no-sandbox").
		Set("disable-dev-shm-usage").
		Headless(true)

	// In container, we want headless=new if possible, but go-rod handles it via flags
	l.Set("headless", "new")

	// Ensure launch is also covered by context
	controlURL, err := l.Context(ctx).Launch()
	if err != nil {
		return fmt.Errorf("failed to launch chromium (timeout or binary error): %w", err)
	}

	// Connect with context
	browser := rod.New().ControlURL(controlURL).Context(ctx)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}
	defer browser.Close()

	// Create page
	page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Set the HTML content
	if err := page.SetDocumentContent(htmlContent); err != nil {
		return fmt.Errorf("failed to set document content: %w", err)
	}

	// Wait for load with timeout (enforced by context)
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("page wait load failed: %w", err)
	}

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
		return fmt.Errorf("pdf generation failed: %w", err)
	}

	pdfBytes, err := io.ReadAll(pdfStream)
	if err != nil {
		return fmt.Errorf("failed to read pdf stream: %w", err)
	}

	if err := os.WriteFile(outputPath, pdfBytes, 0644); err != nil {
		return fmt.Errorf("failed to write pdf file: %w", err)
	}

	return nil
}
