package middlewares

import (
	"net/http"
	"strings"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware processes the Authorization header to extract and validate the BE-generated Bearer token.
// After validation, it automatically fetches a valid Tuya access token and stores it in the context.
//
// @param tuyaAuthUC The TuyaAuthUseCase used to get/refresh Tuya tokens.
// @return gin.HandlerFunc The Gin middleware handler.
// @throws 401 If the Authorization header is missing, malformed, or the token is invalid.
func AuthMiddleware(tuyaAuthUC *usecases.TuyaAuthUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.LogDebug("AuthMiddleware: processing request")
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.LogWarn("AuthMiddleware: missing Authorization Header")
			c.JSON(http.StatusUnauthorized, dtos.StandardResponse{
				Status:  false,
				Message: "Authorization header is required",
				Data:    nil,
			})
			c.Abort()
			return
		}

		fields := strings.Fields(authHeader)
		var beToken string
		if len(fields) == 2 && strings.EqualFold(fields[0], "Bearer") {
			beToken = fields[1]
		} else if len(fields) == 1 {
			beToken = fields[0]
		} else {
			utils.LogWarn("AuthMiddleware: invalid Authorization Header format: %q", authHeader)
			c.JSON(http.StatusUnauthorized, dtos.StandardResponse{
				Status:  false,
				Message: "Invalid Authorization header format. Expected 'Bearer <token>'",
				Data:    nil,
			})
			c.Abort()
			return
		}
		beToken = strings.TrimSpace(beToken)

		// Validate BE Token
		uid, err := utils.ValidateToken(beToken)
		if err != nil {
			utils.LogWarn("AuthMiddleware: token validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, dtos.StandardResponse{
				Status:  false,
				Message: "Invalid or expired token",
				Data:    nil,
			})
			c.Abort()
			return
		}

		// Store BE token and UID in context
		c.Set("be_access_token", beToken)
		c.Set("uid", uid)

		// Auto-fetch Tuya Access Token
		tuyaToken, err := tuyaAuthUC.GetTuyaAccessToken()
		if err != nil {
			utils.LogError("AuthMiddleware: failed to auto-fetch Tuya token: %v", err)
			c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
				Status:  false,
				Message: "Failed to authenticate with Tuya cloud",
				Data:    nil,
			})
			c.Abort()
			return
		}

		// Store Tuya Access Token in context for downstream usecases
		c.Set("tuya_access_token", tuyaToken)
		c.Set("access_token", tuyaToken) // For compatibility with existing controllers
		utils.LogDebug("AuthMiddleware: token validated and Tuya token acquired for UID: %s", uid)

		tuyaUID := c.GetHeader("X-TUYA-UID")
		if tuyaUID != "" {
			c.Set("tuya_uid", tuyaUID)
		}

		c.Next()
	}
}
