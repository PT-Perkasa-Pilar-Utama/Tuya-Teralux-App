package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	mail_dtos "teralux_app/domain/mail/dtos"
	"teralux_app/domain/mail/usecases"
)

type MailSendController struct {
	useCase usecases.MailSendUseCase
}

func NewMailSendController(useCase usecases.MailSendUseCase) *MailSendController {
	return &MailSendController{
		useCase: useCase,
	}
}

// SendMail handles POST /api/mail/send
// @Summary Send an email using a template
// @Description Send an email using a server-side template and specified recipients
// @Tags 09. Mail
// @Accept json
// @Produce json
// @Param request body mail_dtos.MailSendRequestDTO true "Mail Request"
// @Security BearerAuth
// @Success 200 {object} dtos.StandardResponse
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/mail/send [post]
func (c *MailSendController) SendMail(ctx *gin.Context) {
	var req mail_dtos.MailSendRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.LogError("MailSendController.SendMail: Invalid request body: %v", err)

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

	err := c.useCase.SendMail(&req)
	if err != nil {
		utils.LogError("MailSendController.SendMail: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Email sent successfully",
	})
}
