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

// MailSendByMacUseCase defines the interface for sending emails by MAC address.
type MailSendByMacUseCase interface {
	SendMailByMac(macAddress string, req *dtos.SendMailByMacRequestDTO) (string, error)
}

type mailSendByMacUseCase struct {
	mailService         *services.MailService
	mailExternalService *services.MailExternalService
	store               *tasks.StatusStore[dtos.MailStatusDTO]
	cache               *tasks.BadgerTaskCache
}

// NewMailSendByMacUseCase initializes a new mailSendByMacUseCase.
func NewMailSendByMacUseCase(
	mailService *services.MailService,
	mailExternalService *services.MailExternalService,
	store *tasks.StatusStore[dtos.MailStatusDTO],
	cache *tasks.BadgerTaskCache,
) MailSendByMacUseCase {
	return &mailSendByMacUseCase{
		mailService:         mailService,
		mailExternalService: mailExternalService,
		store:               store,
		cache:               cache,
	}
}

func (uc *mailSendByMacUseCase) SendMailByMac(macAddress string, req *dtos.SendMailByMacRequestDTO) (string, error) {
	// Normalization
	macAddress = strings.ToUpper(strings.TrimSpace(macAddress))

	// Validation
	if macAddress == "" {
		return "", utils.NewValidationError("Validation Error", []utils.ValidationErrorDetail{
			{Field: "mac_address", Message: "mac_address is required"},
		})
	}

	if req.Subject == "" {
		return "", utils.NewValidationError("Validation Error", []utils.ValidationErrorDetail{
			{Field: "subject", Message: "subject is required"},
		})
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

	utils.LogInfo("MailSendByMacUseCase: Started task %s for MAC %s", taskID, macAddress)

	go uc.processAsync(taskID, macAddress, req)

	return taskID, nil
}

func (uc *mailSendByMacUseCase) processAsync(taskID string, macAddress string, req *dtos.SendMailByMacRequestDTO) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogError("Mail Task %s (MAC): Panic recovered: %v", taskID, r)
			uc.updateStatus(taskID, "failed", fmt.Errorf("internal panic: %v", r), "")
		}
	}()

	// Fetch device/customer info
	utils.LogDebug("MailSendByMacUseCase: Fetching device info for MAC %s", macAddress)
	info, err := uc.mailExternalService.GetDeviceInfoByMac(macAddress)
	if err != nil {
		utils.LogError("Mail Task %s (MAC): Failed to fetch device info: %v", taskID, err)
		uc.updateStatus(taskID, "failed", err, "")
		return
	}

	// Extract email
	recipientEmail, ok := info["SDTGetRoomTeraluxItemCustomerEmail"].(string)
	if !ok || strings.TrimSpace(recipientEmail) == "" {
		utils.LogError("Mail Task %s (MAC): Customer email not found", taskID)
		uc.updateStatus(taskID, "failed", fmt.Errorf("customer email not found for this device"), "")
		return
	}

	templateName := req.Template
	if templateName == "" {
		templateName = "test"
	}

	// Map external data to template variables
	templateData := map[string]interface{}{
		"CustomerName": info["SDTGetRoomTeraluxCustomerName"],
		"RoomName":     info["SDTGetRoomTeraluxRoomName"],
		"RoomPassword": info["SDTGetRoomTeraluxItemRoomPassword"],
		"BookingTime":  info["SDTGetRoomTeraluxBookingtimeChar"],
		"Date":         info["SDTGetRoomTeraluxItemDate"],
		"Status":       info["SDTGetRoomTeraluxItemStatus"],
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
			utils.LogDebug("MailSendByMacUseCase: Resolved attachment path to %s", attachmentPath)
		} else {
			utils.LogWarn("MailSendByMacUseCase: Attachment file not found at %s", fullPath)
			attachmentPath = ""
		}
	}

	utils.LogDebug("MailSendByMacUseCase: Sending email to %s for MAC %s", recipientEmail, macAddress)
	err = uc.mailService.SendEmailWithTemplate([]string{recipientEmail}, req.Subject, templateName, templateData, attachmentPath)
	if err != nil {
		utils.LogError("Mail Task %s (MAC): Failed to send email: %v", taskID, err)
		uc.updateStatus(taskID, "failed", err, "")
		return
	}

	utils.LogInfo("Mail Task %s (MAC): Email sent successfully", taskID)
	uc.updateStatus(taskID, "completed", nil, fmt.Sprintf("Email sent to %s", recipientEmail))
}

func (uc *mailSendByMacUseCase) updateStatus(taskID string, statusStr string, err error, result string) {
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
