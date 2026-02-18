package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
)

type EmailController struct {
	emailService *infrastructure.EmailService
}

func NewEmailController(cfg *utils.Config) *EmailController {
	service := infrastructure.NewEmailService(cfg)
	return &EmailController{emailService: service}
}

type SendEmailRequest struct {
	To       []string `json:"to" binding:"required"`
	Subject  string   `json:"subject" binding:"required"`
	Template string   `json:"template" binding:"omitempty"`
	Body     string   `json:"body" binding:"omitempty"`
}

// SendEmail godoc
// @Summary Send an email
// @Description Send an email using a template or raw body
// @Tags 07. Common
// @Accept json
// @Produce json
// @Param request body SendEmailRequest true "Email Request"
// @Security BearerAuth
// @Success 200 {object} dtos.StandardResponse
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/email/send [post]
func (c *EmailController) SendEmail(ctx *gin.Context) {
	var req SendEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.LogError("EmailController.SendEmail: Invalid request body: %v", err)
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "payload", Message: "Invalid request body: " + err.Error()},
			},
		})
		return
	}

	// Basic Validation
	if len(req.To) == 0 {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "to", Message: "Recipients list is required"},
			},
		})
		return
	}

	if req.Subject == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "subject", Message: "Subject is required"},
			},
		})
		return
	}

	if req.Template == "" && req.Body == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "content", Message: "Either 'template' or 'body' is required"},
			},
		})
		return
	}

	// Logic execution
	var err error
	if req.Body != "" {
		// Send raw body (could be HTML or Text)
		err = c.emailService.SendEmail(req.To, req.Subject, req.Body)
	} else {
		// Send via template
		err = c.emailService.SendEmailWithTemplate(req.To, req.Subject, req.Template)
	}

	if err != nil {
		utils.LogError("EmailController.SendEmail: %v", err)
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
