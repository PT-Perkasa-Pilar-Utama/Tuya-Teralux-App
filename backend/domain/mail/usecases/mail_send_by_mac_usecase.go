package usecases

import (
	"fmt"
	"strings"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/mail/dtos"
	"teralux_app/domain/mail/services"
)

// MailSendByMacUseCase defines the interface for sending emails by MAC address.
type MailSendByMacUseCase interface {
	SendMailByMac(macAddress string, req *dtos.SendMailByMacRequestDTO) (string, error)
}

type mailSendByMacUseCase struct {
	mailService         *services.MailService
	mailExternalService *services.MailExternalService
}

// NewMailSendByMacUseCase initializes a new mailSendByMacUseCase.
func NewMailSendByMacUseCase(mailService *services.MailService, mailExternalService *services.MailExternalService) MailSendByMacUseCase {
	return &mailSendByMacUseCase{
		mailService:         mailService,
		mailExternalService: mailExternalService,
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

	// Fetch device/customer info
	utils.LogDebug("MailSendByMacUseCase: Fetching device info for MAC %s", macAddress)
	info, err := uc.mailExternalService.GetDeviceInfoByMac(macAddress)
	if err != nil {
		return "", fmt.Errorf("failed to fetch device info: %w", err)
	}

	// Extract email
	recipientEmail, ok := info["SDTGetRoomTeraluxItemCustomerEmail"].(string)
	if !ok || strings.TrimSpace(recipientEmail) == "" {
		return "", utils.NewAPIError(404, "Customer email not found for this device")
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

	utils.LogDebug("MailSendByMacUseCase: Sending email to %s for MAC %s", recipientEmail, macAddress)
	err = uc.mailService.SendEmailWithTemplate([]string{recipientEmail}, req.Subject, templateName, templateData)
	if err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	return recipientEmail, nil
}
