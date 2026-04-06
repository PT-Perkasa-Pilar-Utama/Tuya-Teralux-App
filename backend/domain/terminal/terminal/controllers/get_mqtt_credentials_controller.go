package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/services"

	"github.com/gin-gonic/gin"
)

// Force usage for Swagger
var _ = terminal_dtos.MQTTCredentialsResponseDTO{}

// GetMQTTCredentialsController handles fetching MQTT credentials.
type GetMQTTCredentialsController struct {
	mqttClient *services.MqttAuthClient
}

// NewGetMQTTCredentialsController creates a new instance.
func NewGetMQTTCredentialsController(mqttClient *services.MqttAuthClient) *GetMQTTCredentialsController {
	return &GetMQTTCredentialsController{
		mqttClient: mqttClient,
	}
}

// GetMQTTCredentials fetches the MQTT password generated during terminal creation.
// @Summary      Get MQTT credentials
// @Description  Get MQTT credentials by username for device authentication
// @Tags         02. Terminal
// @Accept       json
// @Produce      json
// @Param        username  path  string  true  "MQTT Username"
// @Success      200  {object}  dtos.StandardResponse{data=terminal_dtos.MQTTCredentialsResponseDTO}
// @Failure      404  {object}  dtos.ErrorResponse
// @Failure      422  {object}  dtos.ValidationErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Router       /api/mqtt/users/{username} [get]
// @Security     BearerAuth
func (c *GetMQTTCredentialsController) GetMQTTCredentials(ctx *gin.Context) {
	username := ctx.Param("username")

	if username == "" {
		ctx.JSON(http.StatusUnprocessableEntity, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "username", Message: "username is required"},
			},
		})
		return
	}

	creds, err := c.mqttClient.GetMQTTCredentials(username)
	if err != nil {
		utils.LogError("Failed to fetch MQTT credentials for user %s: %v", username, err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to fetch MQTT credentials",
			Data:    nil,
		})
		return
	}

	if creds == nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "MQTT credentials not found",
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Credentials retrieved successfully",
		Data:    creds,
	})
}
