package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	commonServices "sensio/domain/common/services"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	"sensio/domain/mail/dtos"
	"sensio/domain/mail/services"
	"strings"
	"time"
)

// MailSendByMacUseCase defines the interface for sending emails by MAC address.
type MailSendByMacUseCase interface {
	SendMailByMac(macAddress string, req *dtos.SendMailByMacRequestDTO) (string, error)
}

type mailSendByMacUseCase struct {
	mailService        *services.MailService
	bigExternalService *commonServices.DeviceInfoExternalService
	store              *tasks.StatusStore[dtos.MailStatusDTO]
	cache              *tasks.BadgerTaskCache
}

// NewMailSendByMacUseCase initializes a new mailSendByMacUseCase.
func NewMailSendByMacUseCase(
	mailService *services.MailService,
	bigExternalService *commonServices.DeviceInfoExternalService,
	store *tasks.StatusStore[dtos.MailStatusDTO],
	cache *tasks.BadgerTaskCache,
) MailSendByMacUseCase {
	return &mailSendByMacUseCase{
		mailService:        mailService,
		bigExternalService: bigExternalService,
		store:              store,
		cache:              cache,
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
	info, err := uc.bigExternalService.GetDeviceInfoByMac(macAddress)
	if err != nil {
		utils.LogError("Mail Task %s (MAC): Failed to fetch device info: %v", taskID, err)
		uc.updateStatus(taskID, "failed", err, "")
		return
	}

	// Extract email from external API
	rawRecipientEmail, ok := info["SDTGetRoomTeraluxItemCustomerEmail"].(string)

	// Override email if provided in req.Data
	var overriddenRecipients []string
	if overrideData, hasOverride := req.Data["email"]; hasOverride {
		switch v := overrideData.(type) {
		case string:
			trimmed := strings.TrimSpace(v)
			if trimmed != "" {
				if strings.Contains(trimmed, ",") {
					parts := strings.Split(trimmed, ",")
					for _, p := range parts {
						tp := strings.TrimSpace(p)
						if tp != "" {
							overriddenRecipients = append(overriddenRecipients, tp)
						}
					}
				} else {
					overriddenRecipients = append(overriddenRecipients, trimmed)
				}
			}
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok {
					trimmed := strings.TrimSpace(s)
					if trimmed != "" {
						overriddenRecipients = append(overriddenRecipients, trimmed)
					}
				}
			}
		case []string:
			for _, s := range v {
				trimmed := strings.TrimSpace(s)
				if trimmed != "" {
					overriddenRecipients = append(overriddenRecipients, trimmed)
				}
			}
		}
	}

	var recipients []string
	switch {
	case len(overriddenRecipients) > 0:
		recipients = overriddenRecipients
		utils.LogDebug("Mail Task %s (MAC): Overriding recipients with %v", taskID, recipients)
	case ok && strings.TrimSpace(rawRecipientEmail) != "":
		recipients = []string{strings.TrimSpace(rawRecipientEmail)}
	default:
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
		"brand_name":       brandName,
		"customer_name":    info["SDTGetRoomTeraluxCustomerName"],
		"customer_company": info["SDTGetRoomTeraluxItemCompanyName"],
		"booking_date": func() interface{} {
			date := info["SDTGetRoomTeraluxByendDate"]
			if date == nil || date == "" || date == "<nil>" {
				return info["SDTGetRoomTeraluxtimeendDate"]
			}
			return date
		}(),
		"booking_time_start": timeStart,
		"booking_time_stop":  timeStop,
		"booking_place": func() interface{} {
			building := info["SDTGetRoomTeraluxBuildingsName"]
			if building == nil || building == "" || building == "<nil>" {
				return info["SDTGetRoomTeraluxRoomName"]
			}
			return building
		}(),
		"booking_room":       info["SDTGetRoomTeraluxRoomName"],
		"agenda_context": func() interface{} {
			// 1. Try external API
			apiAgenda := info["SDTGetRoomTeraluxMeetingAgenda"]
			if apiAgenda != nil && apiAgenda != "" && apiAgenda != "<nil>" {
				return apiAgenda
			}

			// 2. Try cache (inferred from recent summary)
			cacheKey := fmt.Sprintf("cache:agenda_mac_%s", macAddress)
			var cachedAgenda string
			_, found, err := uc.cache.GetWithTTL(cacheKey, &cachedAgenda)
			if err == nil && found && cachedAgenda != "" {
				utils.LogDebug("MailSendByMacUseCase: Using cached inferred agenda for MAC %s", macAddress)
				return cachedAgenda
			}

			return ""
		}(),
		"has_attachment": req.AttachmentPath != nil && *req.AttachmentPath != "",
	}

	// Merge with custom data from request (request data takes precedence)
	for k, v := range req.Data {
		templateData[k] = v
	}

	attachmentPath := req.AttachmentPath
	if attachmentPath != nil && *attachmentPath != "" && (strings.HasPrefix(*attachmentPath, "/uploads") || strings.HasPrefix(*attachmentPath, "http")) {
		// Resolve to local disk path
		wd, _ := os.Getwd()
		baseDir := wd
		if !strings.HasSuffix(wd, "backend") {
			if _, err := os.Stat("backend"); err == nil {
				baseDir = filepath.Join(wd, "backend")
			}
		}

		var relPath string
		if strings.HasPrefix(*attachmentPath, "http") {
			// Full URL — extract relative path after /uploads
			if idx := strings.Index(*attachmentPath, "/uploads"); idx != -1 {
				relPath = (*attachmentPath)[idx+1:] // e.g. "uploads/reports/f.pdf"
			}
		} else {
			relPath = strings.TrimPrefix(*attachmentPath, "/") // e.g. "uploads/reports/f.pdf"
		}

		if relPath == "" {
			attachmentPath = nil
		} else {
			fullPath := filepath.Join(baseDir, relPath)
			if _, err := os.Stat(fullPath); err == nil {
				attachmentPath = &fullPath
				utils.LogDebug("MailSendByMacUseCase: Resolved attachment path to %s", *attachmentPath)
			} else {
				utils.LogWarn("MailSendByMacUseCase: Attachment file not found at %s", fullPath)
				attachmentPath = nil
			}
		}
	}

	// Dynamic Subject Generation if empty or placeholder
	finalSubject := req.Subject
	if finalSubject == "" || strings.EqualFold(finalSubject, "Auto-generated") || strings.EqualFold(finalSubject, "Meeting Summary") {
		roomName := fmt.Sprintf("%v", info["SDTGetRoomTeraluxRoomName"])
		if roomName == "" || roomName == "<nil>" {
			roomName = "Terminal"
		}

		bookingDateRaw := templateData["booking_date"]
		bookingDateStr := fmt.Sprintf("%v", bookingDateRaw)
		if bookingDateStr == "" || bookingDateStr == "<nil>" {
			bookingDateStr = time.Now().Format("02 Jan 2006")
		}

		finalSubject = fmt.Sprintf("%s Meeting Summary - %s", roomName, bookingDateStr)
	}

	utils.LogDebug("MailSendByMacUseCase: Sending email to %v for MAC %s with subject: %s", recipients, macAddress, finalSubject)
	err = uc.mailService.SendEmailWithTemplate(recipients, finalSubject, templateName, templateData, attachmentPath)
	if err != nil {
		utils.LogError("Mail Task %s (MAC): Failed to send email: %v", taskID, err)
		uc.updateStatus(taskID, "failed", err, "")
		return
	}

	utils.LogInfo("Mail Task %s (MAC): Email sent successfully", taskID)
	uc.updateStatus(taskID, "completed", nil, fmt.Sprintf("Email sent to %s", strings.Join(recipients, ", ")))
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
