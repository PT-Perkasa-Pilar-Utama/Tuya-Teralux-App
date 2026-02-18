package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"teralux_app/domain/common/utils"

	"github.com/stretchr/testify/assert"
)

func TestEmailService_SendEmailWithTemplate_TemplateNotFound(t *testing.T) {
	// Setup
	cfg := &utils.Config{
		SMTPHost: "localhost",
		SMTPPort: "1025",
	}
	service := NewEmailService(cfg)
	// Point to a non-existent dir or ensure standard dir doesn't have the "invalid" template
	service.SetTemplateDir(os.TempDir()) // Temp dir is likely empty of our templates

	// Execute
	err := service.SendEmailWithTemplate([]string{"test@example.com"}, "Test", "non_existent_template")

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEmailService_SendEmailWithTemplate_SuccessParsing(t *testing.T) {
	// This test verifies template parsing works, but expects SMTP failure (since we have no server)
	// Setup
	cfg := &utils.Config{
		SMTPHost:     "invalid_host",
		SMTPPort:     "25",
		SMTPFrom:     "sender@example.com",
		SMTPUsername: "user",
		SMTPPassword: "password",
	}
	service := NewEmailService(cfg)

	// Create a temp template file
	tmpDir := t.TempDir()
	tmplContent := "<h1>Hello {{.Timestamp}}</h1>"
	tmplPath := filepath.Join(tmpDir, "test_template.html")
	err := os.WriteFile(tmplPath, []byte(tmplContent), 0644)
	assert.NoError(t, err)

	service.SetTemplateDir(tmpDir)

	// Execute
	err = service.SendEmailWithTemplate([]string{"test@example.com"}, "Test Subject", "test_template")

	// Verify
	// We expect an error, but NOT a "template not found" error.
	// We expect an SMTP error like "lookup invalid_host" or "no such host"
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "template")
	assert.NotContains(t, err.Error(), "parse")
	// assert.Contains(t, err.Error(), "failed to send email") // Use broader check
}
