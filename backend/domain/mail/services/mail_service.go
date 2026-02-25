package services

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html/template"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"teralux_app/domain/common/utils"
)

const smtpDialTimeout = 180 * time.Second

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

// GetConfig returns the mail service configuration
func (s *MailService) GetConfig() *utils.Config {
	return s.config
}

func (s *MailService) SendEmail(to []string, subject string, body string, attachmentPath string) error {
	if s.config.SMTPHost == "" || s.config.SMTPPort == "" {
		return fmt.Errorf("SMTP configuration is missing")
	}

	auth := LoginAuth(s.config.SMTPUsername, s.config.SMTPPassword)
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)

	var msg []byte
	var err error

	// We always use multipart if we want to support inline images (CID)
	// or if there's an actual attachment.
	msg, err = s.buildMultipartMessage(to, subject, body, attachmentPath)
	if err != nil {
		return fmt.Errorf("failed to build multipart message: %w", err)
	}

	host, _, _ := net.SplitHostPort(addr)
	return s.sendEmailDirect(addr, host, auth, s.config.SMTPFrom, to, msg)
}

func (s *MailService) sendEmailDirect(addr string, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial failed: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("smtp StartTLS failed: %w", err)
	}

	// Try PlainAuth first, fallback to LoginAuth
	plainAuth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, host)
	if err := client.Auth(plainAuth); err != nil {
		utils.LogWarn("smtp PlainAuth failed (falling back to LoginAuth): %v", err)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed (both Plain and Login): %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp MAIL FROM failed: %w", err)
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp RCPT TO %s failed: %w", rcpt, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA failed: %w", err)
	}
	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("smtp write body failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp body close failed: %w", err)
	}

	return client.Quit()
}

func (s *MailService) buildMultipartMessage(to []string, subject string, body string, attachmentPath string) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	// Headers
	// Headers
	fmt.Fprintf(buf, "From: %s\r\n", s.config.SMTPFrom)
	fmt.Fprintf(buf, "To: %s\r\n", strings.Join(to, ","))
	fmt.Fprintf(buf, "Subject: %s\r\n", subject)
	fmt.Fprintf(buf, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(buf, "Message-ID: <%d.%d@teralux.app>\r\n", time.Now().UnixNano(), os.Getpid())
	fmt.Fprintf(buf, "MIME-Version: 1.0\r\n")

	// If we have an attachment, we use multipart/mixed.
	fmt.Fprintf(buf, "Content-Type: multipart/mixed; boundary=%s\r\n", writer.Boundary())
	fmt.Fprintf(buf, "\r\n") // Empty line indicates end of headers

	// 1. Text/HTML Body Part
	bodyPartHeader := make(textproto.MIMEHeader)
	bodyPartHeader.Set("Content-Type", "text/html; charset=\"UTF-8\"")
	bodyPartHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	bodyPartWriter, err := writer.CreatePart(bodyPartHeader)
	if err != nil {
		return nil, err
	}

	qp := quotedprintable.NewWriter(bodyPartWriter)
	qp.Write([]byte(body))
	qp.Close()

	// 2. Logo Part (CID embedding)
	// Try to find logo in assets/images/logo.png
	wd, _ := os.Getwd()
	logoPath := filepath.Join(wd, "assets", "images", "logo.png")
	// Fallback check
	if _, err := os.Stat(logoPath); os.IsNotExist(err) {
		altLogo := filepath.Join(wd, "backend", "assets", "images", "logo.png")
		if _, err := os.Stat(altLogo); err == nil {
			logoPath = altLogo
		}
	}

	if _, err := os.Stat(logoPath); err == nil {
		logoData, err := os.ReadFile(logoPath)
		if err == nil {
			logoHeader := make(textproto.MIMEHeader)
			logoHeader.Set("Content-Type", "image/png")
			logoHeader.Set("Content-Transfer-Encoding", "base64")
			logoHeader.Set("Content-ID", "<logo>")
			logoHeader.Set("Content-Disposition", "inline; filename=\"logo.png\"")

			logoWriter, err := writer.CreatePart(logoHeader)
			if err == nil {
				encoded := base64.StdEncoding.EncodeToString(logoData)
				logoWriter.Write([]byte(chunkBase64(encoded)))
			}
		}
	}

	// 3. Attachment Part
	if attachmentPath != "" {
		file, err := os.Open(attachmentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open attachment: %w", err)
		}
		defer file.Close()

		fileName := filepath.Base(attachmentPath)
		attachmentHeader := make(textproto.MIMEHeader)
		attachmentHeader.Set("Content-Type", mime.TypeByExtension(filepath.Ext(fileName)))
		attachmentHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
		attachmentHeader.Set("Content-Transfer-Encoding", "base64")

		attachmentWriter, err := writer.CreatePart(attachmentHeader)
		if err != nil {
			return nil, err
		}

		attachmentData, err := os.ReadFile(attachmentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read attachment: %w", err)
		}

		encoded := base64.StdEncoding.EncodeToString(attachmentData)
		attachmentWriter.Write([]byte(chunkBase64(encoded)))
	}

	writer.Close()
	return buf.Bytes(), nil
}

func chunkBase64(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); i += 76 {
		end := i + 76
		if end > len(s) {
			end = len(s)
		}
		buf.WriteString(s[i:end])
		buf.WriteString("\r\n")
	}
	return buf.String()
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
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		serverMsg := strings.ToLower(string(fromServer))
		if strings.Contains(serverMsg, "user") || strings.Contains(serverMsg, "username") {
			return []byte(a.username), nil
		}
		if strings.Contains(serverMsg, "pass") || strings.Contains(serverMsg, "password") {
			return []byte(a.password), nil
		}
		// Fallback to step-based if server message is unclear
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

func (s *MailService) SendEmailWithTemplate(to []string, subject string, templateName string, data interface{}, attachmentPath *string) error {
	if data == nil {
		// Default dummy data if none provided
		data = map[string]interface{}{
			"Timestamp":    "Just Now",
			"Date":         "Today",
			"SummaryItems": []string{"Item 1", "Item 2"},
		}
	}

	// Internal dereference for SendEmail
	pathStr := ""
	if attachmentPath != nil {
		pathStr = *attachmentPath
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

	return s.SendEmail(to, subject, body.String(), pathStr)
}
