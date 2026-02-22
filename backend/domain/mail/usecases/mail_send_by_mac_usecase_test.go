package usecases

import (
	"teralux_app/domain/common/utils"
	"teralux_app/domain/mail/dtos"
	"teralux_app/domain/mail/services"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMailSendByMacUseCase_SendMailByMac_Validation(t *testing.T) {
	cfg := &utils.Config{
		SMTPHost: "localhost",
		SMTPPort: "1025",
	}
	svc := services.NewMailService(cfg)
	extSvc := services.NewMailExternalService()
	uc := NewMailSendByMacUseCase(svc, extSvc)

	t.Run("Empty MAC", func(t *testing.T) {
		email, err := uc.SendMailByMac("", &dtos.SendMailByMacRequestDTO{Subject: "Test"})
		assert.Error(t, err)
		assert.Empty(t, email)
		assert.Contains(t, err.Error(), "mac_address is required")
	})

	t.Run("Empty Subject", func(t *testing.T) {
		email, err := uc.SendMailByMac("AA:BB:CC:DD:EE:FF", &dtos.SendMailByMacRequestDTO{Subject: ""})
		assert.Error(t, err)
		assert.Empty(t, email)
		assert.Contains(t, err.Error(), "subject is required")
	})

	t.Run("Valid UUID Format", func(t *testing.T) {
		// This should pass validation. 
		// If external API is reachable, it might try to send email.
		email, err := uc.SendMailByMac("db329671-96bb-368b-95d3-53a3a3712563", &dtos.SendMailByMacRequestDTO{Subject: "Test"})
		
		// If it reaches SMTP call, it will fail due to no server, which is fine
		if err != nil {
			assert.NotContains(t, err.Error(), "invalid mac address format")
			assert.Empty(t, email)
		} else {
			assert.NotEmpty(t, email)
		}
	})
}
