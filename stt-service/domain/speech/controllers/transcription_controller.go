package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"stt-service/domain/common/config"
	"stt-service/domain/speech/dtos"
	"stt-service/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

type TranscriptionController struct {
	usecase *usecases.TranscriptionUsecase
	config  *config.Config
}

func NewTranscriptionController(usecase *usecases.TranscriptionUsecase, cfg *config.Config) *TranscriptionController {
	return &TranscriptionController{
		usecase: usecase,
		config:  cfg,
	}
}

// HandleTranscribe godoc
// @Summary Transcribe audio file
// @Description Uploads an audio file (MP3/WAV) and returns the transcribed text
// @Tags transcription
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file to transcribe"
// @Success 200 {object} dtos.StandardResponse{data=dtos.TranscriptionResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 413 {object} dtos.StandardResponse
// @Failure 415 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /transcribe [post]
func (c *TranscriptionController) HandleTranscribe(ctx *gin.Context) {
	file, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to get audio file",
			Details: err.Error(),
		})
		return
	}

	// 1. Check file size from config
	if file.Size > c.config.MaxFileSize {
		ctx.JSON(http.StatusRequestEntityTooLarge, dtos.StandardResponse{
			Status:  false,
			Message: "File too large",
			Details: fmt.Sprintf("Max file size allowed is %dMB", c.config.MaxFileSize/(1024*1024)),
		})
		return
	}

	// 2. Check file extension (Supported: .mp3, .wav, .m4a, .aac, .ogg)
	ext := filepath.Ext(file.Filename)
	supportedExts := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".m4a":  true,
		".aac":  true,
		".ogg":  true,
		".flac": true,
	}
	if !supportedExts[ext] {
		ctx.JSON(http.StatusUnsupportedMediaType, dtos.StandardResponse{
			Status:  false,
			Message: "Unsupported file type",
			Details: "Only .mp3, .wav, .m4a, .aac, .ogg, .flac are supported",
		})
		return
	}

	// Create temp directory if not exists
	tempDir := "./tmp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		os.Mkdir(tempDir, 0755)
	}

	inputPath := filepath.Join(tempDir, file.Filename)
	if err := ctx.SaveUploadedFile(file, inputPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to save uploaded file",
			Details: err.Error(),
		})
		return
	}
	defer os.Remove(inputPath)

	text, err := c.usecase.TranscribeAudio(inputPath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Transcription failed",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Transcription successful",
		Data: dtos.TranscriptionResponseDTO{
			Text: text,
		},
	})
}
