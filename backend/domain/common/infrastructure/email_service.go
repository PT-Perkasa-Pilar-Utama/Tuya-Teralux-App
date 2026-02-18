package infrastructure

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"

	"teralux_app/domain/common/utils"
)

type EmailService struct {
	config      *utils.Config
	TemplateDir string
}

func NewEmailService(cfg *utils.Config) *EmailService {
	// Determine template directory relative to exec path or standard location
	wd, _ := os.Getwd()
	// Default guess
	tmplDir := filepath.Join(wd, "domain", "common", "templates")
	
	// Fallback check for "backend/domain..."
	if _, err := os.Stat(tmplDir); os.IsNotExist(err) {
		altDir := filepath.Join(wd, "backend", "domain", "common", "templates")
		if _, err := os.Stat(altDir); err == nil {
			tmplDir = altDir
		}
	}

	return &EmailService{
		config:      cfg,
		TemplateDir: tmplDir, // Default
	}
}

// SetTemplateDir allows setting a custom template directory (e.g. for testing)
func (s *EmailService) SetTemplateDir(dir string) {
	s.TemplateDir = dir
}

func (s *EmailService) SendEmail(to []string, subject string, body string) error {
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// Standard SMTP headers
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subjectHeader := fmt.Sprintf("Subject: %s\n", subject)
	// To header is a bit complex for multiple recipients, usually handled by SMTP server but good practice to include
	toHeader := fmt.Sprintf("To: %s\n", to[0]) // Simplification for first recipient

	msg := []byte(toHeader + subjectHeader + mime + body)

	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)
	if err := smtp.SendMail(addr, auth, s.config.SMTPFrom, to, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (s *EmailService) SendEmailWithTemplate(to []string, subject string, templateName string) error {
	data := map[string]interface{}{
		"Timestamp":    "Just Now",
		"Date":         "Today",
		"SummaryItems": []string{"Item 1", "Item 2"},
	}

	// Use the configured TemplateDir
	tmplPath := filepath.Join(s.TemplateDir, templateName+".html")

	// Check if exists
	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		return fmt.Errorf("template %s not found in %s", templateName, s.TemplateDir)
	}

	parsedTemplate, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
	}

	var body bytes.Buffer
	if err := parsedTemplate.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String())
}
