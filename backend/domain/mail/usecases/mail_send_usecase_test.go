package usecases

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"teralux_app/domain/common/tasks"
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

	store := tasks.NewStatusStore[dtos.MailStatusDTO]()
	uc := NewMailSendUseCase(svc, store, nil)

	req := &dtos.MailSendRequestDTO{
		To:      []string{"user@example.com"},
		Subject: "Default Template Test",
	}
	
	taskID, err := uc.SendMail(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)
}

func TestMailSendUseCase_SendMail_SpecificTemplate(t *testing.T) {
	cfg := &utils.Config{
		SMTPHost: "localhost",
		SMTPPort: "1025",
	}
	svc := services.NewMailService(cfg)
	cleanup := createTestTemplate(t, svc, "summary")
	defer cleanup()

	store := tasks.NewStatusStore[dtos.MailStatusDTO]()
	uc := NewMailSendUseCase(svc, store, nil)

	req := &dtos.MailSendRequestDTO{
		To:       []string{"user@example.com"},
		Subject:  "Specific Template Test",
		Template: "summary",
	}
	
	taskID, err := uc.SendMail(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)
}

func TestMailSendUseCase_SendMail_ValidationError_To(t *testing.T) {
	store := tasks.NewStatusStore[dtos.MailStatusDTO]()
	uc := NewMailSendUseCase(nil, store, nil)
	req := &dtos.MailSendRequestDTO{
		To:      []string{},
		Subject: "Test",
	}
	
	taskID, err := uc.SendMail(req)
	assert.Error(t, err)
	assert.Empty(t, taskID)
	assert.Contains(t, err.Error(), "to field is required")
}

func TestMailSendUseCase_SendMail_ValidationError_Subject(t *testing.T) {
	store := tasks.NewStatusStore[dtos.MailStatusDTO]()
	uc := NewMailSendUseCase(nil, store, nil)
	req := &dtos.MailSendRequestDTO{
		To:      []string{"user@example.com"},
		Subject: "",
	}
	
	taskID, err := uc.SendMail(req)
	assert.Error(t, err)
	assert.Empty(t, taskID)
	assert.Contains(t, err.Error(), "subject field is required")
}

func TestMailSendUseCase_SendMail_TemplateNotFound(t *testing.T) {
	svc := services.NewMailService(&utils.Config{})
	svc.SetTemplateDir("/tmp/ghost")
	store := tasks.NewStatusStore[dtos.MailStatusDTO]()
	uc := NewMailSendUseCase(svc, store, nil)

	req := &dtos.MailSendRequestDTO{
		To:       []string{"user@example.com"},
		Subject:  "Test",
		Template: "ghost",
	}

	// Async design: SendMail always returns taskID immediately (no upfront template check)
	taskID, err := uc.SendMail(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)

	// Wait briefly for background goroutine to process and mark task as failed
	time.Sleep(100 * time.Millisecond)

	status, ok := store.Get(taskID)
	assert.True(t, ok, "task should exist in store")
	if ok {
		assert.Equal(t, "failed", status.Status)
		assert.Contains(t, status.Error, "ghost not found")
	}
}
