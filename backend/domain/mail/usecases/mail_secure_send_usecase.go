package usecases

import (
	"context"
	"fmt"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/download_token"
	"sensio/domain/mail/dtos"
	"sensio/domain/mail/entities"
	"sensio/domain/mail/repositories"
	"sensio/domain/mail/services"
	"time"
)

type MetadataStatus int

const (
	MetadataStatusUnknown MetadataStatus = iota
	MetadataStatusPending
	MetadataStatusCompleted
	MetadataStatusFailed
)

type MetadataChecker interface {
	GetS3ZipStatus(ctx context.Context, objectKey string) (MetadataStatus, error)
	GetS3PdfStatus(ctx context.Context, objectKey string) (MetadataStatus, error)
}

type SecureMailUseCase interface {
	SendSecureLinkWithPassword(ctx context.Context, recipient, objectKey, purpose, subject string) (string, error)
	SetMetadataChecker(checker MetadataChecker)
}

type secureMailUseCase struct {
	mailService     *services.MailService
	tokenService    *download_token.DownloadTokenService
	store           *tasks.StatusStore[dtos.MailStatusDTO]
	cache           *tasks.BadgerTaskCache
	metadataChecker MetadataChecker
	outboxRepo      repositories.MailOutboxRepository
}

func NewSecureMailUseCase(
	mailService *services.MailService,
	tokenService *download_token.DownloadTokenService,
	store *tasks.StatusStore[dtos.MailStatusDTO],
	cache *tasks.BadgerTaskCache,
) SecureMailUseCase {
	return &secureMailUseCase{
		mailService:  mailService,
		tokenService: tokenService,
		store:        store,
		cache:        cache,
		outboxRepo:   repositories.NewMailOutboxRepository(),
	}
}

func (uc *secureMailUseCase) SetMetadataChecker(checker MetadataChecker) {
	uc.metadataChecker = checker
}

func (uc *secureMailUseCase) SendSecureLinkWithPassword(ctx context.Context, recipient, objectKey, purpose, subject string) (string, error) {
	if recipient == "" {
		return "", fmt.Errorf("recipient is required")
	}
	if objectKey == "" {
		return "", fmt.Errorf("object key is required")
	}

	var existingStatus dtos.MailStatusDTO
	if _, found, _ := uc.cache.GetIdempotencyTask(recipient, objectKey, &existingStatus); found {
		if existingStatus.Status == "completed" {
			return "", fmt.Errorf("email already sent for this recipient and object")
		}
		if existingStatus.Status == "pending" {
			return "", fmt.Errorf("email send already in progress for this recipient and object")
		}
	}

	taskID := utils.GenerateUUID()
	status := &dtos.MailStatusDTO{
		Status:    "pending",
		StartedAt: time.Now().Format(time.RFC3339),
		ExpiresAt: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	}

	uc.store.Set(taskID, status)
	_ = uc.cache.Set(taskID, status)
	_ = uc.cache.SetIdempotencyTask(recipient, objectKey, taskID, status)

	uc.outboxRepo.Create(&entities.MailOutbox{ //nolint:errcheck
		TaskID:    taskID,
		Recipient: recipient,
		ObjectKey: objectKey,
		Purpose:   purpose,
		Subject:   subject,
		Status:    entities.MailOutboxStatusPending,
		CreatedAt: time.Now(),
	})

	go uc.processAsync(ctx, taskID, recipient, objectKey, purpose, subject)

	return taskID, nil
}

func (uc *secureMailUseCase) processAsync(_ context.Context, taskID, recipient, objectKey, purpose, subject string) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogError("SecureMail Task %s: Panic recovered: %v", taskID, r)
			uc.updateStatus(taskID, "failed", fmt.Errorf("internal panic"), "")
		}
	}()

	token, err := uc.tokenService.CreateToken(recipient, objectKey, purpose)
	if err != nil {
		utils.LogError("SecureMail Task %s: Failed to create token: %v", taskID, err)
		uc.updateStatus(taskID, "failed", err, "")
		return
	}

	linkData := map[string]interface{}{
		"download_link": fmt.Sprintf("/api/download/resolve?state=%s&purpose=%s", token, purpose),
	}

	utils.LogInfo("SecureMail Task %s: Sending link email to %s", taskID, recipient)
	err = uc.mailService.SendEmailWithTemplate(
		[]string{recipient},
		subject+" - Download Link",
		"secure_link",
		linkData,
		nil,
	)
	if err != nil {
		utils.LogError("SecureMail Task %s: Failed to send link email: %v", taskID, err)
		uc.updateStatus(taskID, "failed", err, "")
		return
	}

	utils.LogInfo("SecureMail Task %s: Link email sent successfully", taskID)
	uc.updateStatus(taskID, "completed", nil, fmt.Sprintf("Secure email link sent to %s", recipient))
}

func (uc *secureMailUseCase) updateStatus(taskID string, statusStr string, err error, result string) {
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
