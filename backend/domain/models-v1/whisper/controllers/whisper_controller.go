package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/models-v1/whisper/services"

	"github.com/gin-gonic/gin"
)

type WhisperController struct {
	grpcSvc *services.GrpcWhisperService
}

func NewWhisperController(grpcSvc *services.GrpcWhisperService) *WhisperController {
	return &WhisperController{
		grpcSvc: grpcSvc,
	}
}

func (c *WhisperController) Transcribe(ctx *gin.Context) {
	audioPath := ctx.PostForm("audio_path")
	language := ctx.PostForm("language")
	diarize := ctx.PostForm("diarize") == "true"

	if audioPath == "" {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{Status: false, Message: "audio_path is required"})
		return
	}

	resp, err := c.grpcSvc.Transcribe(audioPath, language, diarize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{Status: true, Data: resp})
}

func (c *WhisperController) GetStatus(ctx *gin.Context) {
	taskID := ctx.Param("transcribe_id")
	resp, err := c.grpcSvc.GetStatus(taskID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{Status: true, Data: resp})
}
