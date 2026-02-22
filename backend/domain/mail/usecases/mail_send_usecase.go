package usecases

import (
	"fmt"
	"teralux_app/domain/mail/dtos"
	"teralux_app/domain/mail/services"
)

// MailSendUseCase defines the interface for sending emails.
type MailSendUseCase interface {
	SendMail(req *dtos.MailSendRequestDTO) error
}

type mailSendUseCase struct {
	mailService *services.MailService
}

// NewMailSendUseCase initializes a new mailSendUseCase.
func NewMailSendUseCase(mailService *services.MailService) MailSendUseCase {
	return &mailSendUseCase{
		mailService: mailService,
	}
}

func (uc *mailSendUseCase) SendMail(req *dtos.MailSendRequestDTO) error {
	// Validation (Project standard: UseCase should validate business rules)
	if len(req.To) == 0 {
		return fmt.Errorf("to field is required")
	}
	if req.Subject == "" {
		return fmt.Errorf("subject field is required")
	}

	templateName := req.Template
	if templateName == "" {
		templateName = "test"
	}

	err := uc.mailService.SendEmailWithTemplate(req.To, req.Subject, templateName, nil)
	if err != nil {
		return fmt.Errorf("UseCase.SendMail failed: %w", err)
	}

	return nil
}
