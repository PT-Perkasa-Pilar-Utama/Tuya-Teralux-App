package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

// WhisperProxyProcessor is an abstraction for Whisper proxy operations implemented by the usecase.
// This allows unit tests to provide a fake implementation.
type WhisperProxyProcessor interface {
	ProxyTranscribe(filePath string, fileName string) (string, error)
	GetStatus(taskID string) (*dtos.WhisperProxyStatusDTO, error)
}

type WhisperProxyController struct {
	usecase *usecases.WhisperProxyUsecase
	config  *utils.Config
}

func NewWhisperProxyController(usecase *usecases.WhisperProxyUsecase, cfg *utils.Config) *WhisperProxyController {
	return &WhisperProxyController{
		usecase: usecase,
		config:  cfg,
	}
}

// ProxyTranscribeAudio godoc
// @Summary Transcribe audio file (Proxy to Outsystems)
// @Description Submit audio file for transcription via Outsystems proxy. Processing is asynchronous.
// @Tags 04. Speech
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)"
// @Success 202 {object} dtos.StandardResponse{data=dtos.WhisperProxyProcessResponseDTO}
// @Failure 400 {object} dtos.StandardResponse
// @Failure 413 {object} dtos.StandardResponse
// @Failure 415 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/speech/transcribe/ppu [post]
func (c *WhisperProxyController) HandleProxyTranscribe(ctx *gin.Context) {
	log.Println("[DEBUG] HandleProxyTranscribe: Request received")
	log.Printf("[DEBUG] HandleProxyTranscribe: Content-Type: %s", ctx.GetHeader("Content-Type"))

	file, err := ctx.FormFile("audio")
	if err != nil {
		log.Printf("[DEBUG] HandleProxyTranscribe: FormFile error: %v", err)
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to get audio file",
			Details: err.Error(),
		})
		return
	}
	log.Printf("[DEBUG] HandleProxyTranscribe: File received: %s, Size: %d", file.Filename, file.Size)

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

	taskID, err := c.usecase.ProxyTranscribe(inputPath, file.Filename)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to submit transcription task",
			Details: err.Error(),
		})
		return
	}

	// Try to fetch cached status via DTO to include TTL info in response (optional)
	status, _ := c.usecase.GetStatus(taskID)
	// Use DTOs directly in the response, avoid hardcoding TTL values here
	if status != nil {
		ctx.JSON(http.StatusAccepted, dtos.StandardResponse{
			Status:  true,
			Message: "Task submitted",
			Data: map[string]interface{}{
				"task_id": taskID,
				"status":  status,
			},
		})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{
		Status:  true,
		Message: "Task submitted",
		Data: map[string]string{
			"task_id": taskID,
		},
	})
}

// GetProxyTranscribeStatus godoc
// @Summary [DEPRECATED] Get whisper proxy transcription task status
// @Description [DEPRECATED] Use /api/speech/transcribe/{transcribe_id} instead.
func (c *WhisperProxyController) GetProxyTranscribeStatus(ctx *gin.Context) {
	id := ctx.Param("transcribe_id")
	status, err := c.usecase.GetStatus(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.StandardResponse{
			Status:  false,
			Message: "Not found",
			Details: err.Error(),
		})
		return
	}

	httpStatus := c.determineHTTPStatus(status)
	message := "OK"
	if httpStatus >= 400 {
		message = http.StatusText(httpStatus)
	}
	ctx.JSON(httpStatus, dtos.StandardResponse{
		Status:  httpStatus < 400,
		Message: message,
		Data:    status,
	})
}

func (c *WhisperProxyController) determineHTTPStatus(status *dtos.WhisperProxyStatusDTO) int {
	if status.Status == "error" {
		return http.StatusInternalServerError
	}
	if status.Status == "pending" {
		return http.StatusAccepted
	}
	if status.Status == "completed" {
		return http.StatusOK
	}
	return http.StatusOK
}
