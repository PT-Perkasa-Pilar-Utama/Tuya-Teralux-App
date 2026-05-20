package usecases

import (
	"fmt"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/mail/dtos"
	"sensio/domain/mail/services"
	"strings"
	"time"
)

type MailSendUseCase interface {
	SendMail(req *dtos.MailSendRequestDTO) (string, error)
}

type mailSendUseCase struct {
	mailService *services.MailService
	store       *tasks.StatusStore[dtos.MailStatusDTO]
	cache       *tasks.BadgerTaskCache
}

func NewMailSendUseCase(mailService *services.MailService, store *tasks.StatusStore[dtos.MailStatusDTO], cache *tasks.BadgerTaskCache) MailSendUseCase {
	return &mailSendUseCase{
		mailService: mailService,
		store:       store,
		cache:       cache,
	}
}

func (uc *mailSendUseCase) toPublicDownloadURL(path string) string {
	if strings.HasPrefix(path, "/api/reports/") && !strings.Contains(path, "://") {
		base := utils.GetConfig().BackendPublicBaseURL
		if base != "" {
			return base + path
		}
	}
	return path
}

func (uc *mailSendUseCase) SendMail(req *dtos.MailSendRequestDTO) (string, error) {
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

	attachment := resolveMailAttachment(req.AttachmentPath)
	defer func() {
		if attachment.cleanup != nil {
			attachment.cleanup()
		}
	}()

	if req.Data == nil {
		req.Data = make(map[string]interface{})
	}
	req.Data["download_url"] = uc.toPublicDownloadURL(attachment.downloadURL)
	req.Data["has_attachment"] = attachment.path != nil && attachment.downloadURL == ""
	if req.AudioURL != nil && strings.TrimSpace(*req.AudioURL) != "" {
		req.Data["audio_url"] = strings.TrimSpace(*req.AudioURL)
	}

	err := uc.mailService.SendEmailWithTemplate(req.To, req.Subject, templateName, req.Data, attachment.path)
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

	if statusStr == "completed" || statusStr == "failed" {
		if existing.StartedAt != "" {
			startTime, _ := time.Parse(time.RFC3339, existing.StartedAt)
			status.DurationSeconds = time.Since(startTime).Seconds()
		}
	}

	uc.store.Set(taskID, status)
	_ = uc.cache.SetPreserveTTL(taskID, status)
}