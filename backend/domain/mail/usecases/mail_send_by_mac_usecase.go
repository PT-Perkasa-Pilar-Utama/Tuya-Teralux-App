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

	// Extract email from external API
	recipientEmail, ok := info["SDTGetRoomTeraluxItemCustomerEmail"].(string)

	// Override email if provided in req.Data
	if overrideEmail, hasOverride := req.Data["email"].(string); hasOverride && strings.TrimSpace(overrideEmail) != "" {
		recipientEmail = strings.TrimSpace(overrideEmail)
		utils.LogDebug("Mail Task %s (MAC): Overriding recipient email with %s", taskID, recipientEmail)
	} else if !ok || strings.TrimSpace(recipientEmail) == "" {
		utils.LogError("Mail Task %s (MAC): Customer email not found in external API and no override provided", taskID)
		uc.updateStatus(taskID, "failed", fmt.Errorf("customer email not found for this device"), "")
		return
	}

	templateName := req.Template
	if templateName == "" {
		templateName = "test"
	}

	// Extract brand name from config or default
	brandName := "Sensio"

	// Parsing booking time start/stop if possible
	bookingTime := fmt.Sprintf("%v", info["SDTGetRoomTeraluxBookingtimeChar"])
	timeStart := ""
	timeStop := ""
	if strings.Contains(bookingTime, "–") {
		parts := strings.Split(bookingTime, "–")
		if len(parts) == 2 {
			timeStart = strings.TrimSpace(parts[0])
			timeStop = strings.TrimSpace(parts[1])
		}
	} else if strings.Contains(bookingTime, "-") {
		parts := strings.Split(bookingTime, "-")
		if len(parts) == 2 {
			timeStart = strings.TrimSpace(parts[0])
			timeStop = strings.TrimSpace(parts[1])
		}
	}

	if timeStart == "" {
		timeStart = bookingTime // Fallback
	}

	// Map external data to template variables
	templateData := map[string]interface{}{
		"brand_name":         brandName,
		"customer_name":      info["SDTGetRoomTeraluxCustomerName"],
		"customer_company":   info["SDTGetRoomTeraluxItemCompanyName"],
		"booking_date":       info["SDTGetRoomTeraluxtimeendDate"],
		"booking_time_start": timeStart,
		"booking_time_stop":  timeStop,
		"booking_place":      info["SDTGetRoomTeraluxRoomName"], // Mapping room name to place as building name is missing in new API
		"booking_room":       info["SDTGetRoomTeraluxRoomName"],
		"has_attachment":     req.AttachmentPath != nil && *req.AttachmentPath != "",
	}

	// Merge with custom data from request (request data takes precedence)
	for k, v := range req.Data {
		templateData[k] = v
	}

	attachmentPath := req.AttachmentPath
	if attachmentPath != nil && *attachmentPath != "" && strings.HasPrefix(*attachmentPath, "/uploads") {
		// Resolve relative path to local disk path
		wd, _ := os.Getwd()
		baseDir := wd
		if !strings.HasSuffix(wd, "backend") {
			if _, err := os.Stat("backend"); err == nil {
				baseDir = filepath.Join(wd, "backend")
			}
		}

		relPath := strings.TrimPrefix(*attachmentPath, "/")
		fullPath := filepath.Join(baseDir, relPath)

		if _, err := os.Stat(fullPath); err == nil {
			attachmentPath = &fullPath
			utils.LogDebug("MailSendByMacUseCase: Resolved attachment path to %s", *attachmentPath)
		} else {
			utils.LogWarn("MailSendByMacUseCase: Attachment file not found at %s", fullPath)
			attachmentPath = nil
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
