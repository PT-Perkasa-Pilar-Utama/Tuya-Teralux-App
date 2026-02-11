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


// HandleProxyTranscribe godoc
// @Summary Transcribe audio file (Whisper)
// @Description Start transcription of audio file using local Whisper STT. Asynchronous processing. Supports: .mp3, .wav, .m4a, .aac, .ogg, .flac.
// @Tags 04. Speech
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)"
// @Success 202 {object} dtos.StandardResponse{data=dtos.TranscriptionTaskResponseDTO}
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
		if err := os.Mkdir(tempDir, 0755); err != nil {
			utils.LogError("HandleProxyTranscribe: Failed to create temp directory: %v", err)
		}
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
	// Use ProxyTranscribeAudio which uses local Whisper
	taskID, err := c.usecase.ProxyTranscribeAudio(inputPath, file.Filename)
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
		Data: dtos.TranscriptionTaskResponseDTO{
			TaskID:     taskID,
			TaskStatus: "pending",
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
// @Success 200 {object} dtos.StandardResponse{data=dtos.AsyncTranscriptionProcessStatusResponseDTO}
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
			Data: dtos.AsyncTranscriptionProcessStatusResponseDTO{
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
			Data: dtos.AsyncTranscriptionProcessStatusResponseDTO{
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
			Data: dtos.WhisperProxyProcessStatusResponseDTO{
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

// HandleWhisperCppTranscribe godoc
// @Summary Transcribe audio file (Whisper.cpp)
// @Description Start transcription of audio file using Whisper.cpp. Asynchronous processing with background execution. Pure Whisper.cpp, no PPU. Supports: .mp3, .wav, .m4a, .aac, .ogg, .flac.
// @Tags 04. Speech
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)"
// @Param language formData string true "Language code (e.g. id, en)"
// @Success 202 {object} dtos.StandardResponse{data=dtos.TranscriptionTaskResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 413 {object} dtos.StandardResponse
// @Failure 415 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/speech/transcribe/whisper/cpp [post]
func (c *TranscriptionController) HandleWhisperCppTranscribe(ctx *gin.Context) {
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
		if err := os.Mkdir(tempDir, 0755); err != nil {
			utils.LogError("HandleWhisperCppTranscribe: Failed to create temp directory: %v", err)
		}
	}

	inputPath := filepath.Join(tempDir, "whisper_cpp_"+file.Filename)
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
		Data: dtos.TranscriptionTaskResponseDTO{
			TaskID:     taskID,
			TaskStatus: "pending",
		},
	})
}
