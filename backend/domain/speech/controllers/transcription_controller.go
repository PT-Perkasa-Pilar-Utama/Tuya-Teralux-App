package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"teralux_app/domain/common/utils"
	recordingUsecases "teralux_app/domain/recordings/usecases"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

type TranscriptionController struct {
	transcribeUC           usecases.TranscribeUseCase
	transcribeWhisperCppUC usecases.TranscribeWhisperCppUseCase
	getStatusUC            usecases.GetTranscriptionStatusUseCase
	saveRecordingUC        recordingUsecases.SaveRecordingUseCase
	config                 *utils.Config
}

func NewTranscriptionController(
	transcribeUC usecases.TranscribeUseCase,
	transcribeWhisperCppUC usecases.TranscribeWhisperCppUseCase,
	getStatusUC usecases.GetTranscriptionStatusUseCase,
	saveRecordingUC recordingUsecases.SaveRecordingUseCase,
	cfg *utils.Config,
) *TranscriptionController {
	return &TranscriptionController{
		transcribeUC:           transcribeUC,
		transcribeWhisperCppUC: transcribeWhisperCppUC,
		getStatusUC:            getStatusUC,
		saveRecordingUC:        saveRecordingUC,
		config:                 cfg,
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
// @Param language formData string false "Language code (e.g. id, en)"
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

	// 1. Save as recording first (Automatic metadata + physical save)
	recording, err := c.saveRecordingUC.Execute(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to save recording metadata",
			Details: err.Error(),
		})
		return
	}

	// 2. Submit async task using the file provided by SaveRecording (cleaner pathing)
	// construction of path matches SaveRecordingUseCase behavior
	finalInputPath := filepath.Join("uploads", "audio", recording.Filename)

	// Extract language (optional, default to "id")
	language := ctx.PostForm("language")
	if language == "" {
		language = "id"
	}

	taskID, err := c.transcribeUC.Execute(finalInputPath, file.Filename, language)
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
			TaskID:      taskID,
			TaskStatus:  "pending",
			RecordingID: recording.ID,
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

	status, err := c.getStatusUC.Execute(taskID)
	if err == nil {
		ctx.JSON(http.StatusOK, dtos.StandardResponse{
			Status:  true,
			Message: "Task status retrieved successfully",
			Data:    status,
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

	// 1. Save as recording first
	recording, err := c.saveRecordingUC.Execute(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to save recording metadata",
			Details: err.Error(),
		})
		return
	}

	// 2. Submit async task
	finalPath := filepath.Join("uploads", "audio", recording.Filename)
	taskID, err := c.transcribeWhisperCppUC.Execute(finalPath, file.Filename, lang)
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
			TaskID:      taskID,
			TaskStatus:  "pending",
			RecordingID: recording.ID,
		},
	})
}
