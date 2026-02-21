package controllers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"teralux_app/domain/common/utils"
	recordingUsecases "teralux_app/domain/recordings/usecases"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

// SpeechTranscribeOrionController handles POST /api/speech/transcribe/orion.
type SpeechTranscribeOrionController struct {
	proxyUC       usecases.WhisperProxyUsecase
	saveRecording recordingUsecases.SaveRecordingUseCase
	config        *utils.Config
}

func NewSpeechTranscribeOrionController(
	proxyUC usecases.WhisperProxyUsecase,
	saveRecording recordingUsecases.SaveRecordingUseCase,
	cfg *utils.Config,
) *SpeechTranscribeOrionController {
	return &SpeechTranscribeOrionController{
		proxyUC:       proxyUC,
		saveRecording: saveRecording,
		config:        cfg,
	}
}

// TranscribeOrion handles POST /api/speech/transcribe/orion
// @Summary Transcribe audio file (Proxy to Outsystems)
// @Description Submit audio file for transcription via Outsystems proxy. Processing is asynchronous.
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
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/speech/transcribe/orion [post]
func (c *SpeechTranscribeOrionController) TranscribeOrion(ctx *gin.Context) {
	log.Println("[DEBUG] TranscribeOrion: Request received")
	log.Printf("[DEBUG] TranscribeOrion: Content-Type: %s", ctx.GetHeader("Content-Type"))

	file, err := ctx.FormFile("audio")
	if err != nil {
		log.Printf("[DEBUG] TranscribeOrion: FormFile error: %v", err)
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "audio", Message: "Audio file is required: " + err.Error()},
			},
		})
		return
	}
	log.Printf("[DEBUG] TranscribeOrion: File received: %s, Size: %d", file.Filename, file.Size)

	if file.Size > c.config.MaxFileSize {
		ctx.JSON(http.StatusRequestEntityTooLarge, dtos.StandardResponse{
			Status:  false,
			Message: "File too large",
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
			Message: "Unsupported Media Type",
		})
		return
	}

	tempDir := "./tmp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		if err := os.Mkdir(tempDir, 0755); err != nil {
			utils.LogError("TranscribeOrion: Failed to create temp directory: %v", err)
		}
	}

	inputPath := filepath.Join(tempDir, file.Filename)
	if err := ctx.SaveUploadedFile(file, inputPath); err != nil {
		utils.LogError("TranscribeOrion.SaveUploadedFile: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	recording, err := c.saveRecording.SaveRecording(file)
	if err != nil {
		utils.LogError("TranscribeOrion.SaveRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	finalPath := filepath.Join("uploads", "audio", recording.Filename)

	language := ctx.PostForm("language")
	if language == "" {
		language = "id"
	}

	taskID, err := c.proxyUC.ProxyTranscribe(finalPath, file.Filename, language)
	if err != nil {
		utils.LogError("TranscribeOrion.ProxyTranscribe: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{
		Status:  true,
		Message: "Task submitted",
		Data: dtos.TranscriptionTaskResponseDTO{
			TaskID:      taskID,
			TaskStatus:  "pending",
			RecordingID: recording.ID,
		},
	})
}
