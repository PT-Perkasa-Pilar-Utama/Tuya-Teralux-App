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
	Token string `json:"token"`
}

type ResolveTokenResponse struct {
	URL string `json:"url"`
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
			Token: token,
		},
	})
}

func (h *Handler) ResolveToken(ctx *gin.Context) {
	tokenString := ctx.Query("state")
	client := ctx.Query("client")
	purpose := ctx.Query("purpose")

	if tokenString == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "missing required query param: state",
		})
		return
	}

	signedURL, err := h.service.ResolveToken(tokenString, client, purpose)
	if err != nil {
		switch {
		case errors.Is(err, entities.ErrTokenNotFound):
			ctx.JSON(http.StatusUnauthorized, dtos.StandardResponse{Status: false, Message: err.Error()})
		case errors.Is(err, entities.ErrTokenExpired):
			ctx.JSON(http.StatusGone, dtos.StandardResponse{Status: false, Message: err.Error()})
		default:
			ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{Status: false, Message: err.Error()})
		}
		return
	}

	ctx.Redirect(http.StatusFound, signedURL)
}
