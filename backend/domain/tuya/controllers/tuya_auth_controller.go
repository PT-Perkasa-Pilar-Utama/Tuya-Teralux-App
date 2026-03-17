package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	tuya_dtos "sensio/domain/tuya/dtos"
	"sensio/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// TuyaAuthController handles authentication requests for Tuya
type TuyaAuthController struct {
	useCase usecases.TuyaAuthUseCase
}

// NewTuyaAuthController creates a new TuyaAuthController instance
func NewTuyaAuthController(useCase usecases.TuyaAuthUseCase) *TuyaAuthController {
	return &TuyaAuthController{
		useCase: useCase,
	}
}

var _ = tuya_dtos.TuyaAuthResponseDTO{}

// Authenticate handles GET /api/tuya/auth endpoint
// @Summary      Authenticate with Tuya
// @Description  Authenticates the user and retrieves a Tuya access token
// @Tags         01. Tuya
// @Accept       json
// @Produce      json
// @Success      200  {object}  dtos.StandardResponse{data=tuya_dtos.TuyaAuthResponseDTO}
// @Failure      500  {object}  dtos.ErrorResponse
// @Security     ApiKeyAuth
// @Router       /api/tuya/auth [get]
func (c *TuyaAuthController) Authenticate(ctx *gin.Context) {
	utils.LogDebug("Authenticate request received")
	token, err := c.useCase.Authenticate()
	if err != nil {
		utils.LogError("TuyaAuthController.Authenticate: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	utils.LogDebug("Authentication successful")
	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Authentication successful",
		Data:    token,
	})
}
