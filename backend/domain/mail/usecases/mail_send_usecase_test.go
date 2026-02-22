package usecases

import (
	"os"
	"path/filepath"
	"testing"

	"teralux_app/domain/common/utils"
	"teralux_app/domain/mail/dtos"
	"teralux_app/domain/mail/services"

	"github.com/stretchr/testify/assert"
)

func createTestTemplate(t *testing.T, svc *services.MailService, name string) func() {
	tmpDir, _ := os.MkdirTemp("", "templates")
	tmplPath := filepath.Join(tmpDir, name+".html")
	os.WriteFile(tmplPath, []byte("<h1>Test</h1>"), 0644)
	svc.SetTemplateDir(tmpDir)
	return func() { os.RemoveAll(tmpDir) }
}

func TestMailSendUseCase_SendMail_DefaultTemplate(t *testing.T) {
	cfg := &utils.Config{
		SMTPHost: "localhost",
		SMTPPort: "1025",
	}
	svc := services.NewMailService(cfg)
	cleanup := createTestTemplate(t, svc, "test")
	defer cleanup()

	uc := NewMailSendUseCase(svc)

	req := &dtos.MailSendRequestDTO{
		To:      []string{"user@example.com"},
		Subject: "Default Template Test",
	}
	
	err := uc.SendMail(req)
	// Expect SMTP error since no server running, but logic should reach SMTP call
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email")
}

func TestMailSendUseCase_SendMail_SpecificTemplate(t *testing.T) {
	cfg := &utils.Config{
		SMTPHost: "localhost",
		SMTPPort: "1025",
	}
	svc := services.NewMailService(cfg)
	cleanup := createTestTemplate(t, svc, "summary")
	defer cleanup()

	uc := NewMailSendUseCase(svc)

	req := &dtos.MailSendRequestDTO{
		To:       []string{"user@example.com"},
		Subject:  "Specific Template Test",
		Template: "summary",
	}
	
	err := uc.SendMail(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email")
}

func TestMailSendUseCase_SendMail_ValidationError_To(t *testing.T) {
	uc := NewMailSendUseCase(nil)
	req := &dtos.MailSendRequestDTO{
		To:      []string{},
		Subject: "Test",
	}
	
	err := uc.SendMail(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to field is required")
}

func TestMailSendUseCase_SendMail_ValidationError_Subject(t *testing.T) {
	uc := NewMailSendUseCase(nil)
	req := &dtos.MailSendRequestDTO{
		To:      []string{"user@example.com"},
		Subject: "",
	}
	
	err := uc.SendMail(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subject field is required")
}

func TestMailSendUseCase_SendMail_TemplateNotFound(t *testing.T) {
	svc := services.NewMailService(&utils.Config{})
	svc.SetTemplateDir("/tmp/ghost")
	uc := NewMailSendUseCase(svc)

	req := &dtos.MailSendRequestDTO{
		To:       []string{"user@example.com"},
		Subject:  "Test",
		Template: "ghost",
	}
	
	err := uc.SendMail(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ghost not found")
}
