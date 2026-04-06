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

// Transcribe handles POST /api/models/v1/whisper/transcribe
// @Summary      Transcribe audio (v1)
// @Description  Transcribe an audio file by providing its local path
// @Tags         05. Models-v1
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        audio_path  formData  string  true   "Local path to audio file"
// @Param        language    formData  string  false  "Audio language"
// @Param        diarize     formData  bool    false  "Enable speaker diarization"
// @Success      200  {object}  commonDtos.StandardResponse
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router       /api/models/v1/whisper/transcribe [post]
// @Security     BearerAuth
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

// GetStatus handles GET /api/models/v1/whisper/status/:transcribe_id
// @Summary      Get transcription status (v1)
// @Description  Get the status of a transcription task by ID
// @Tags         05. Models-v1
// @Produce      json
// @Param        transcribe_id  path      string  true  "Transcription Task ID"
// @Success      200  {object}  commonDtos.StandardResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router       /api/models/v1/whisper/status/{transcribe_id} [get]
// @Security     BearerAuth
func (c *WhisperController) GetStatus(ctx *gin.Context) {
	taskID := ctx.Param("transcribe_id")
	resp, err := c.grpcSvc.GetStatus(taskID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{Status: false, Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{Status: true, Data: resp})
}
