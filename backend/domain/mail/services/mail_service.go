package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"teralux_app/domain/common/utils"
)

type MailService struct {
	config      *utils.Config
	TemplateDir string
}

func NewMailService(cfg *utils.Config) *MailService {
	// Determine template directory relative to exec path or standard location
	wd, _ := os.Getwd()
	// New location: domain/mail/templates
	tmplDir := filepath.Join(wd, "domain", "mail", "templates")

	// Fallback check for "backend/domain..."
	if _, err := os.Stat(tmplDir); os.IsNotExist(err) {
		altDir := filepath.Join(wd, "backend", "domain", "mail", "templates")
		if _, err := os.Stat(altDir); err == nil {
			tmplDir = altDir
		}
	}

	return &MailService{
		config:      cfg,
		TemplateDir: tmplDir,
	}
}

// SetTemplateDir allows setting a custom template directory (e.g. for testing)
func (s *MailService) SetTemplateDir(dir string) {
	s.TemplateDir = dir
}

func (s *MailService) SendEmail(to []string, subject string, body string) error {
	if s.config.SMTPHost == "" || s.config.SMTPPort == "" {
		return fmt.Errorf("SMTP configuration is missing")
	}

	// Use custom LoginAuth for better compatibility with Hostinger
	auth := LoginAuth(s.config.SMTPUsername, s.config.SMTPPassword)

	// Standard SMTP headers
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subjectHeader := fmt.Sprintf("Subject: %s\n", subject)
	
	toHeader := fmt.Sprintf("To: %s\n", strings.Join(to, ","))

	msg := []byte(toHeader + subjectHeader + mime + body)

	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)
	if err := smtp.SendMail(addr, auth, s.config.SMTPFrom, to, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// loginAuth matches the loginAuth pattern used in diagnostics
type loginAuth struct {
	username, password string
	step               int
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username: username, password: password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch a.step {
		case 0:
			a.step++
			return []byte(a.username), nil
		case 1:
			a.step++
			return []byte(a.password), nil
		}
	}
	return nil, nil
}

func (s *MailService) SendEmailWithTemplate(to []string, subject string, templateName string, data interface{}) error {
	if data == nil {
		// Default dummy data if none provided
		data = map[string]interface{}{
			"Timestamp":    "Just Now",
			"Date":         "Today",
			"SummaryItems": []string{"Item 1", "Item 2"},
		}
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
