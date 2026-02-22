package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"teralux_app/domain/common/tasks"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/mail/dtos"
	"teralux_app/domain/mail/services"
	"time"
)

// MailSendUseCase defines the interface for sending emails.
type MailSendUseCase interface {
	SendMail(req *dtos.MailSendRequestDTO) (string, error)
}

type mailSendUseCase struct {
	mailService *services.MailService
	store       *tasks.StatusStore[dtos.MailStatusDTO]
	cache       *tasks.BadgerTaskCache
}

// NewMailSendUseCase initializes a new mailSendUseCase.
func NewMailSendUseCase(mailService *services.MailService, store *tasks.StatusStore[dtos.MailStatusDTO], cache *tasks.BadgerTaskCache) MailSendUseCase {
	return &mailSendUseCase{
		mailService: mailService,
		store:       store,
		cache:       cache,
	}
}

func (uc *mailSendUseCase) SendMail(req *dtos.MailSendRequestDTO) (string, error) {
	// Validation
	if len(req.To) == 0 {
		return "", fmt.Errorf("to field is required")
	}
	if req.Subject == "" {
		return "", fmt.Errorf("subject field is required")
	}

	taskID := utils.GenerateUUID()
	status := &dtos.MailStatusDTO{
		Status:    "pending",
		StartedAt: time.Now().Format(time.RFC3339),
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	// Mark as pending
	uc.store.Set(taskID, status)
	_ = uc.cache.Set(taskID, status)

	utils.LogInfo("MailSendUseCase: Started task %s for recipients %v", taskID, req.To)

	go uc.processAsync(taskID, req)

	return taskID, nil
}

func (uc *mailSendUseCase) processAsync(taskID string, req *dtos.MailSendRequestDTO) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogError("Mail Task %s: Panic recovered: %v", taskID, r)
			uc.updateStatus(taskID, "failed", fmt.Errorf("internal panic: %v", r), "")
		}
	}()

	templateName := req.Template
	if templateName == "" {
		templateName = "test"
	}

	attachmentPath := req.AttachmentPath
	if attachmentPath != "" && strings.HasPrefix(attachmentPath, "/uploads") {
		// Resolve relative path to local disk path
		wd, _ := os.Getwd()
		baseDir := wd
		if !strings.HasSuffix(wd, "backend") {
			if _, err := os.Stat("backend"); err == nil {
				baseDir = filepath.Join(wd, "backend")
			}
		}

		relPath := strings.TrimPrefix(attachmentPath, "/")
		fullPath := filepath.Join(baseDir, relPath)

		if _, err := os.Stat(fullPath); err == nil {
			attachmentPath = fullPath
			utils.LogDebug("MailSendUseCase: Resolved attachment path to %s", attachmentPath)
		} else {
			utils.LogWarn("MailSendUseCase: Attachment file not found at %s", fullPath)
			attachmentPath = ""
		}
	}

	err := uc.mailService.SendEmailWithTemplate(req.To, req.Subject, templateName, nil, attachmentPath)
	if err != nil {
		utils.LogError("Mail Task %s: Failed to send email: %v", taskID, err)
		uc.updateStatus(taskID, "failed", err, "")
		return
	}

	utils.LogInfo("Mail Task %s: Email sent successfully", taskID)
	uc.updateStatus(taskID, "completed", nil, fmt.Sprintf("Email sent to %s", strings.Join(req.To, ", ")))
}

func (uc *mailSendUseCase) updateStatus(taskID string, statusStr string, err error, result string) {
	var existing dtos.MailStatusDTO
	_, _, _ = uc.cache.GetWithTTL(taskID, &existing)

	status := &dtos.MailStatusDTO{
		Status:    statusStr,
		Result:    result,
		StartedAt: existing.StartedAt,
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	if err != nil {
		status.Error = err.Error()
		status.HTTPStatusCode = utils.GetErrorStatusCode(err)
	} else if statusStr == "completed" {
		status.HTTPStatusCode = 200
	}

	// Calculate duration
	if statusStr == "completed" || statusStr == "failed" {
		if existing.StartedAt != "" {
			startTime, _ := time.Parse(time.RFC3339, existing.StartedAt)
			status.DurationSeconds = time.Since(startTime).Seconds()
		}
	}

	uc.store.Set(taskID, status)
	_ = uc.cache.SetPreserveTTL(taskID, status)
}
