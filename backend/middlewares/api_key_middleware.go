package middlewares

import (
	"net/http"
	"teralux_app/dtos"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

// ApiKeyMiddleware checks for X-API-KEY header
func ApiKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		// Use config
		config := utils.GetConfig()
		validApiKey := config.ApiKey

		if validApiKey == "" {
			// If API_KEY is not configured on server, we might want to fail open or closed.
			// Falsing closed (500) is safer to alert admin.
			c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
				Status:  false,
				Message: "Server misconfiguration: API_KEY not set",
				Data:    nil,
			})
			c.Abort()
			return
		}

		if apiKey != validApiKey {
			c.JSON(http.StatusUnauthorized, dtos.StandardResponse{
				Status:  false,
				Message: "Invalid API Key",
				Data:    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
