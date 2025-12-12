package controllers

import (
	"net/http"
	"teralux_app/dtos"
	"teralux_app/usecases"

	"github.com/gin-gonic/gin"
)

// TuyaAuthController handles authentication requests for Tuya
type TuyaAuthController struct {
	useCase *usecases.TuyaAuthUseCase
}

// NewTuyaAuthController creates a new TuyaAuthController instance
func NewTuyaAuthController(useCase *usecases.TuyaAuthUseCase) *TuyaAuthController {
	return &TuyaAuthController{
		useCase: useCase,
	}
}

// Authenticate handles POST /api/tuya/auth endpoint
func (c *TuyaAuthController) Authenticate(ctx *gin.Context) {
	// Call use case
	token, err := c.useCase.Authenticate()																																																																									
	if err != nil {
		errorResponse := dtos.ErrorResponseDTO{
			Error:   "Authentication failed",
			Message: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// Return success response
	ctx.JSON(http.StatusOK, token)
}
