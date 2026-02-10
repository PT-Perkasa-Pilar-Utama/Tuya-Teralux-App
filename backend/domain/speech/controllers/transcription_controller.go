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
	usecase          *usecases.TranscriptionUsecase
	whisperProxyCase *usecases.WhisperProxyUsecase
	config           *utils.Config
}

func NewTranscriptionController(usecase *usecases.TranscriptionUsecase, whisperProxyCase *usecases.WhisperProxyUsecase, cfg *utils.Config) *TranscriptionController {
	return &TranscriptionController{
		usecase:          usecase,
		whisperProxyCase: whisperProxyCase,
		config:           cfg,
	}
}

// HandlePublishMqtt godoc
// @Summary Publish message to MQTT
// @Description Publish a message to the configured MQTT topic
// @Tags 04. Speech
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

// HandleProxyTranscribe godoc
// @Summary Transcribe audio file (with PPU fallback)
// @Description Start transcription of audio file. Attempts PPU (Outsystems) first, falls back to local Whisper if PPU fails. Asynchronous processing. Supports: .mp3, .wav, .m4a, .aac, .ogg, .flac.
// @Tags 04. Speech
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)"
// @Success 202 {object} dtos.StandardResponse{data=dtos.AsyncTranscriptionProcessResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 413 {object} dtos.StandardResponse
// @Failure 415 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/speech/transcribe [post]
func (c *TranscriptionController) HandleProxyTranscribe(ctx *gin.Context) {
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

	// Submit async task
	// Case 1: Try PPU (Outsystems)
	taskID, err := c.whisperProxyCase.ProxyTranscribe(inputPath, file.Filename)
	if err != nil {
		utils.LogWarn("Transcription Fallback: PPU failed, falling back to local Whisper: %v", err)
		// Case 2: Fallback to Long (Local Whisper)
		// Note: Using "id" as default language for fallback as previously requested for speech
		taskID, err = c.usecase.ProxyTranscribeLongAudio(inputPath, file.Filename, "id")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
				Status:  false,
				Message: "Failed to start transcription task (All methods failed)",
				Details: err.Error(),
			})
			return
		}
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{
		Status:  true,
		Message: "Transcription task submitted successfully",
		Data: dtos.AsyncTranscriptionProcessResponseDTO{
			TaskID: taskID,
		},
	})
}

// GetProxyTranscribeStatus godoc
// @Summary Get transcription status (Consolidated)
// @Description Get the status and result of any transcription task (Short, Long, or PPU).
// @Tags 04. Speech
// @Security BearerAuth
// @Produce json
// @Param transcribe_id path string true "Task ID"
// @Success 200 {object} dtos.StandardResponse{data=dtos.AsyncTranscriptionProcessResponseDTO}
// @Failure 404 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/speech/transcribe/{transcribe_id} [get]
func (c *TranscriptionController) GetProxyTranscribeStatus(ctx *gin.Context) {
	taskID := ctx.Param("transcribe_id")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Task ID is required",
		})
		return
	}

	// Try Long/Short Whisper first
	status, err := c.usecase.GetTranscriptionStatus(taskID)
	if err == nil {
		ctx.JSON(http.StatusOK, dtos.StandardResponse{
			Status:  true,
			Message: "Task status retrieved successfully",
			Data: dtos.AsyncTranscriptionProcessResponseDTO{
				TaskID:     taskID,
				TaskStatus: status,
			},
		})
		return
	}

	// Try Long Whisper status
	statusLong, err := c.usecase.GetTranscriptionLongStatus(taskID)
	if err == nil {
		ctx.JSON(http.StatusOK, dtos.StandardResponse{
			Status:  true,
			Message: "Task status retrieved successfully",
			Data: dtos.AsyncTranscriptionLongProcessResponseDTO{
				TaskID:     taskID,
				TaskStatus: statusLong,
			},
		})
		return
	}

	// Try PPU status
	proxyStatus, err := c.whisperProxyCase.GetStatus(taskID)
	if err == nil {
		ctx.JSON(http.StatusOK, dtos.StandardResponse{
			Status:  true,
			Message: "Task status retrieved successfully",
			Data: dtos.WhisperProxyProcessResponseDTO{
				TaskID:     taskID,
				TaskStatus: proxyStatus,
			},
		})
		return
	}

	ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
		Status:  false,
		Message: "Task not found in any service",
	})
}

// HandleProxyTranscribeLong godoc
// @Summary Transcribe long audio file
// @Description Start transcription of long audio file using Whisper. Asynchronous processing with background execution. No translation, no RAG. Supports: .mp3, .wav, .m4a, .aac, .ogg, .flac.
// @Tags 04. Speech
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)"
// @Param language formData string true "Language code (e.g. id, en)"
// @Success 202 {object} dtos.StandardResponse{data=dtos.AsyncTranscriptionLongProcessResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 413 {object} dtos.StandardResponse
// @Failure 415 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/speech/transcribe/long [post]
func (c *TranscriptionController) HandleProxyTranscribeLong(ctx *gin.Context) {
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

	// Submit async task
	taskID, err := c.usecase.ProxyTranscribeLongAudio(inputPath, file.Filename, lang)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to start transcription task",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{
		Status:  true,
		Message: "Transcription task submitted successfully",
		Data: dtos.AsyncTranscriptionLongProcessResponseDTO{
			TaskID: taskID,
		},
	})
}

// GetProxyTranscribeLongStatus godoc
// @Summary [DEPRECATED] Get long transcription task status
// @Description [DEPRECATED] Use /api/speech/transcribe/{transcribe_id} instead. Gets handle of a long transcription task.
func (c *TranscriptionController) GetProxyTranscribeLongStatus(ctx *gin.Context) {
	taskID := ctx.Param("transcribe_id")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Task ID is required",
		})
		return
	}

	status, err := c.usecase.GetTranscriptionLongStatus(taskID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Task not found",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Task status retrieved successfully",
		Data: dtos.AsyncTranscriptionLongProcessResponseDTO{
			TaskID:     taskID,
			TaskStatus: status,
		},
	})
}
