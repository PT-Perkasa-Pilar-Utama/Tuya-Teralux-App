package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	mail_dtos "teralux_app/domain/mail/dtos"
	"teralux_app/domain/mail/usecases"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// MailSendByMacController handles requests for sending emails by MAC address
type MailSendByMacController struct {
	useCase usecases.MailSendByMacUseCase
}

// NewMailSendByMacController initializes a new MailSendByMacController
func NewMailSendByMacController(useCase usecases.MailSendByMacUseCase) *MailSendByMacController {
	return &MailSendByMacController{
		useCase: useCase,
	}
}

// SendMailByMac handles POST /api/mail/send/mac/:mac_address
// @Summary Send an email by Teralux MAC Address
// @Description Looks up customer email by MAC address and sends an email using a template
// @Tags 09. Mail
// @Accept json
// @Produce json
// @Param mac_address path string true "Teralux MAC Address"
// @Param request body mail_dtos.SendMailByMacRequestDTO true "Mail Request"
// @Security BearerAuth
// @Success 202 {object} dtos.StandardResponse{data=mail_dtos.MailTaskResponseDTO} "Email task submitted successfully"
// @Failure 400 {object} dtos.StandardResponse
// @Failure 404 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/mail/send/mac/{mac_address} [post]
func (c *MailSendByMacController) SendMailByMac(ctx *gin.Context) {
	macAddress := ctx.Param("mac_address")

	var req mail_dtos.SendMailByMacRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.LogError("MailSendByMacController.SendMailByMac: Invalid request body: %v", err)

		var details []utils.ValidationErrorDetail
		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				fieldName := strings.ToLower(fe.Field())
				details = append(details, utils.ValidationErrorDetail{
					Field:   fieldName,
					Message: fmt.Sprintf("%s field is %s", fieldName, fe.Tag()),
				})
			}
		} else {
			details = []utils.ValidationErrorDetail{
				{Field: "payload", Message: "Invalid JSON format or request body: " + err.Error()},
			}
		}

		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: details,
		})
		return
	}

	taskID, err := c.useCase.SendMailByMac(macAddress, &req)
	if err != nil {
		utils.LogError("MailSendByMacController.SendMailByMac: %v", err)
		statusCode := utils.GetErrorStatusCode(err)
		message := "Internal Server Error"

		if statusCode != http.StatusInternalServerError {
			message = err.Error()
		}

		var details interface{}
		if valErr, ok := err.(*utils.ValidationError); ok {
			details = valErr.Details
			message = valErr.Message
		}

		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: message,
			Details: details,
		})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{
		Status:  true,
		Message: "Email task submitted successfully",
		Data: mail_dtos.MailTaskResponseDTO{
			TaskID:     taskID,
			TaskStatus: "pending",
		},
	})
}
