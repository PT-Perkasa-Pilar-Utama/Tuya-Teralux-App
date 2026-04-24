package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"sensio/domain/common/dtos"
	"sensio/domain/common/interfaces"
	"sensio/domain/common/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoginController handles terminal login requests
type LoginController struct {
	terminalRepo interfaces.ITerminalRepository
	authUseCase  interfaces.AuthUseCase
}

// NewLoginController creates a new LoginController instance
func NewLoginController(terminalRepo interfaces.ITerminalRepository, authUseCase interfaces.AuthUseCase) *LoginController {
	return &LoginController{
		terminalRepo: terminalRepo,
		authUseCase:  authUseCase,
	}
}

// Login handles POST /api/login endpoint
// @Summary      Terminal login
// @Description  Authenticates a terminal with Tuya and returns JWT tokens
// @Tags         00. Auth
// @Accept       json
// @Produce      json
// @Param        request body dtos.LoginRequestDTO true "Login request"
// @Success      200 {object} dtos.StandardResponse{data=dtos.LoginResponseDTO}
// @Failure      400 {object} dtos.StandardResponse
// @Failure      404 {object} dtos.StandardResponse
// @Failure      500 {object} dtos.StandardResponse
// @Security     ApiKeyAuth
// @Router       /api/common/login [post]
func (c *LoginController) Login(ctx *gin.Context) {
	var req dtos.LoginRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation failed",
		})
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(req.TerminalID); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid terminal ID format",
		})
		return
	}

	// Check if terminal exists
	terminal, err := c.terminalRepo.GetByID(ctx.Request.Context(), req.TerminalID)
	if err != nil || terminal == nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Terminal not found",
			Data:    map[string]string{"redirect": "register"},
		})
		return
	}

	// Short UUID for logging
	shortID := req.TerminalID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	// Check Authorization header for existing JWT
	authHeader := ctx.GetHeader("Authorization")
	var tokenValid bool
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseTokenWithoutValidation(tokenString)
		if err == nil {
			if exp, ok := claims["exp"].(float64); ok {
				tokenValid = time.Unix(int64(exp), 0).After(time.Now())
			}
		}
	}

	if tokenValid {
		fmt.Printf("[TOKEN_CHECK] terminal_id=%s status=valid skip_tuya=true\n", shortID)
		ctx.JSON(http.StatusOK, dtos.StandardResponse{
			Status:  true,
			Message: "Token still valid",
			Data: dtos.LoginResponseDTO{
				TerminalID: req.TerminalID,
				Status:     "valid",
			},
		})
		return
	}

	fmt.Printf("[TOKEN_CHECK] terminal_id=%s status=expired skip_tuya=false calling_tuya=true\n", shortID)

	// Call Tuya auth
	tuyaResult, err := c.authUseCase.Authenticate()
	if err != nil {
		fmt.Printf("[TOKEN_CHECK] terminal_id=%s status=renewed_error error=tuya_timeout\n", shortID)
		ctx.JSON(http.StatusServiceUnavailable, dtos.StandardResponse{
			Status:  false,
			Message: "Authentication service unavailable",
		})
		return
	}

	// Convert ExpireTime (Unix timestamp) to time.Time
	expiry := time.Unix(int64(tuyaResult.ExpireTime), 0)

	// Generate JWT access token with Tuya payload
	accessToken, err := utils.GenerateLoginToken(tuyaResult.UID, tuyaResult.AccessToken, expiry)
	if err != nil {
		utils.LogError("LoginController.Login: Failed to generate access token: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to generate token",
		})
		return
	}

	// Generate refresh token (stateless JWT, 1 month expiry)
	refreshToken, err := utils.GenerateToken(tuyaResult.UID)
	if err != nil {
		utils.LogError("LoginController.Login: Failed to generate refresh token: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to generate token",
		})
		return
	}

	// Set cookies
	// access_token: regular cookie, Secure=false for development
	ctx.SetCookie("access_token", accessToken, int(time.Until(expiry).Seconds()), "/", "", false, false)
	// refresh_token: http-only cookie, 1 month expiry (30*24*3600 seconds)
	ctx.SetCookie("refresh_token", refreshToken, 30*24*3600, "/", "", true, false)

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Login successful",
		Data: dtos.LoginResponseDTO{
			TerminalID:  req.TerminalID,
			AccessToken: accessToken,
			Message:     "Login successful",
			Status:      "renewed",
		},
	})
}
