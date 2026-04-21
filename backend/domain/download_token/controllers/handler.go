package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"sensio/domain/common/dtos"
	"sensio/domain/download_token/entities"
	"sensio/domain/download_token/services"
)

type Handler struct {
	service *services.DownloadTokenService
}

type CreateTokenRequest struct {
	Recipient string `json:"recipient" binding:"required"`
	ObjectKey string `json:"object_key" binding:"required"`
	Purpose   string `json:"purpose" binding:"required"`
}

type CreateTokenResponse struct {
	TokenID string `json:"token_id"`
}

func NewHandler(service *services.DownloadTokenService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateToken(ctx *gin.Context) {
	var req CreateTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: err.Error(),
		})
		return
	}

	token, err := h.service.CreateToken(req.Recipient, req.ObjectKey, req.Purpose)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dtos.StandardResponse{
		Status:  true,
		Message: "Download token created",
		Data: CreateTokenResponse{
			TokenID: token.TokenID,
		},
	})
}

func (h *Handler) ResolveToken(ctx *gin.Context) {
	tokenID := ctx.Param("token")
	signedURL, err := h.service.ResolveToken(tokenID)
	if err != nil {
		switch {
		case errors.Is(err, entities.ErrTokenNotFound):
			ctx.JSON(http.StatusNotFound, dtos.StandardResponse{Status: false, Message: err.Error()})
		case errors.Is(err, entities.ErrTokenExpired), errors.Is(err, entities.ErrTokenConsumed), errors.Is(err, entities.ErrTokenRevoked):
			ctx.JSON(http.StatusGone, dtos.StandardResponse{Status: false, Message: err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{Status: false, Message: "Internal Server Error"})
		}
		return
	}

	ctx.Redirect(http.StatusFound, signedURL)
}
