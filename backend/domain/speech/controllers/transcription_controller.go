package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

type TranscriptionController struct {
	usecase *usecases.TranscriptionUsecase
	config  *utils.Config
}

func NewTranscriptionController(usecase *usecases.TranscriptionUsecase, cfg *utils.Config) *TranscriptionController {
	return &TranscriptionController{
		usecase: usecase,
		config:  cfg,
	}
}

// TranscribeAudio godoc
// @Summary Transcribe audio file
// @Description Transcribe audio to text using Whisper. Supports: .mp3, .wav, .m4a, .aac, .ogg, .flac.
// @Tags 08. Speech
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)"
// @Success 200 {object} dtos.StandardResponse{data=dtos.TranscriptionResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 413 {object} dtos.StandardResponse
// @Failure 415 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/speech/transcribe [post]
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

	if file.Size > c.config.MaxFileSize {
		ctx.JSON(http.StatusRequestEntityTooLarge, dtos.StandardResponse{
			Status:  false,
			Message: "File too large",
			Details: fmt.Sprintf("Max file size allowed is %dMB", c.config.MaxFileSize/(1024*1024)),
		})
		return
	}

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

	translated, _ := c.usecase.TranslateToEnglish(text)

	// If it looks like a command, process via RAG
	if translated != "" {
		go c.usecase.HandleCommand(translated)
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Transcription and processing successful",
		Data: dtos.TranscriptionResponseDTO{
			Text:           text,
			TranslatedText: translated,
		},
	})
}

// TranscribeLongAudio godoc
// @Summary Transcribe long audio file
// @Description Transcribe long audio to text using Whisper. Supports: .mp3, .wav, .m4a, .aac, .ogg, .flac. No translation, no RAG.
// @Tags 08. Speech
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)"
// @Param language formData string true "Language code (e.g. id, en)"
// @Success 200 {object} dtos.StandardResponse{data=dtos.TranscriptionLongResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 413 {object} dtos.StandardResponse
// @Failure 415 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/speech/transcribe/long [post]
func (c *TranscriptionController) HandleTranscribeLong(ctx *gin.Context) {
	file, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to get audio file",
			Details: err.Error(),
		})
		return
	}

	lang := ctx.PostForm("language")
	if lang == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Language is required",
			Details: "Please provide a language code (e.g. 'id', 'en')",
		})
		return
	}

	if file.Size > c.config.MaxFileSize {
		ctx.JSON(http.StatusRequestEntityTooLarge, dtos.StandardResponse{
			Status:  false,
			Message: "File too large",
			Details: fmt.Sprintf("Max file size allowed is %dMB", c.config.MaxFileSize/(1024*1024)),
		})
		return
	}

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

	tempDir := "./tmp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		os.Mkdir(tempDir, 0755)
	}

	inputPath := filepath.Join(tempDir, "long_"+file.Filename)
	if err := ctx.SaveUploadedFile(file, inputPath); err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to save uploaded file",
			Details: err.Error(),
		})
		return
	}
	defer os.Remove(inputPath)

	text, err := c.usecase.TranscribeLongAudio(inputPath, lang)
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
		Data: dtos.TranscriptionLongResponseDTO{
			Text:             text,
			DetectedLanguage: lang,
		},
	})
}

// HandlePublishMqtt godoc
// @Summary Publish message to MQTT
// @Description Publish a message to the configured MQTT topic
// @Tags 08. Speech
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.MqttPublishRequest true "Message to publish"
// @Success 200 {object} dtos.StandardResponse
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/speech/mqtt/publish [post]
func (c *TranscriptionController) HandlePublishMqtt(ctx *gin.Context) {
	var req dtos.MqttPublishRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if err := c.usecase.PublishToWhisper(req.Message); err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to publish message",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Message published successfully",
	})
}
