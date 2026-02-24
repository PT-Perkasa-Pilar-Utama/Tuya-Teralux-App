package services

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"mime"
	"mime/multipart"
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

	return s.sendWithTimeout(addr, auth, s.config.SMTPFrom, to, msg)
}

// sendWithTimeout dials SMTP manually so we can set TCP-level deadlines.
// smtp.SendMail has no timeout; Hostinger resets connections on long transfers.
func (s *MailService) sendWithTimeout(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// 1. TCP dial with timeout - Force IPv4 (tcp4) to avoid instability
	conn, err := net.DialTimeout("tcp4", addr, smtpDialTimeout)
	if err != nil {
		return fmt.Errorf("smtp dial failed (IPv4): %w", err)
	}
	// Set overall deadline for the entire exchange (dial + auth + DATA)
	if err := conn.SetDeadline(time.Now().Add(smtpDialTimeout)); err != nil {
		conn.Close()
		return fmt.Errorf("failed to set SMTP deadline: %w", err)
	}

	host, _, _ := net.SplitHostPort(addr)

	// 2. Upgrade to TLS (STARTTLS) - required by Hostinger port 587
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp NewClient failed: %w", err)
	}
	defer c.Close()

	tlsConfig := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}
	if err := c.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("smtp StartTLS failed: %w", err)
	}

	// 3. Authenticate
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth failed: %w", err)
	}

	// 4. Set sender and recipients
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("smtp MAIL FROM failed: %w", err)
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp RCPT TO %s failed: %w", rcpt, err)
		}
	}

	// 5. Write message body
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA failed: %w", err)
	}
	if n, err := io.Copy(wc, bytes.NewReader(msg)); err != nil {
		wc.Close()
		return fmt.Errorf("smtp write body failed (sent %d bytes): %w", n, err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp body close failed: %w", err)
	}

	return c.Quit()
}

func (s *MailService) buildMultipartMessage(to []string, subject string, body string, attachmentPath string) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	// Headers
	fmt.Fprintf(buf, "Subject: %s\r\n", subject)
	fmt.Fprintf(buf, "To: %s\r\n", strings.Join(to, ","))
	fmt.Fprintf(buf, "MIME-Version: 1.0\r\n")

	// If we have an attachment, we use multipart/mixed.
	// If we ONLY have inline images, we could use multipart/related.
	// For simplicity and compatibility, multipart/mixed with a nested multipart/related is best,
	// but here we'll just use multipart/mixed and put the logo in it.
	fmt.Fprintf(buf, "Content-Type: multipart/mixed; boundary=%s\r\n", writer.Boundary())
	fmt.Fprintf(buf, "\r\n")

	// 1. Text/HTML Body Part
	bodyPartHeader := make(textproto.MIMEHeader)
	bodyPartHeader.Set("Content-Type", "text/html; charset=\"UTF-8\"")
	bodyPartWriter, err := writer.CreatePart(bodyPartHeader)
	if err != nil {
		return nil, err
	}
	bodyPartWriter.Write([]byte(body))

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
				encoder := base64.NewEncoder(base64.StdEncoding, logoWriter)
				encoder.Write(logoData)
				encoder.Close()
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

		encoder := base64.NewEncoder(base64.StdEncoding, attachmentWriter)
		if _, err := io.Copy(encoder, file); err != nil {
			return nil, fmt.Errorf("failed to copy attachment content: %w", err)
		}
		encoder.Close()
	}

	writer.Close()
	return buf.Bytes(), nil
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
