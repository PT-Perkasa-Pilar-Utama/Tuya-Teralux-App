package controllers

import (
	"net/http"
	"path/filepath"
	"teralux_app/domain/common/utils"
	recordingUsecases "teralux_app/domain/recordings/usecases"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

type SpeechModelsGeminiController struct {
	usecase       usecases.TranscribeGeminiModelUseCase
	saveRecording recordingUsecases.SaveRecordingUseCase
	config        *utils.Config
}

func NewSpeechModelsGeminiController(
	usecase usecases.TranscribeGeminiModelUseCase,
	saveRecording recordingUsecases.SaveRecordingUseCase,
	cfg *utils.Config,
) *SpeechModelsGeminiController {
	return &SpeechModelsGeminiController{
		usecase:       usecase,
		saveRecording: saveRecording,
		config:        cfg,
	}
}

// Transcribe handles POST /api/speech/models/gemini
// @Summary Transcribe audio file (Gemini)
// @Description Submit audio file for transcription via Gemini. Processing is asynchronous.
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
// @Router /api/speech/models/gemini [post]
func (c *SpeechModelsGeminiController) Transcribe(ctx *gin.Context) {
	file, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "audio", Message: "Audio file is required: " + err.Error()},
			},
		})
		return
	}

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

	recording, err := c.saveRecording.SaveRecording(file)
	if err != nil {
		utils.LogError("Gemini.SaveRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	finalPath := filepath.Join("uploads", "audio", recording.Filename)
	language := ctx.PostForm("language")

	taskID, err := c.usecase.TranscribeAsync(finalPath, file.Filename, language, ctx.Request.URL.Path)
	if err != nil {
		utils.LogError("Gemini.TranscribeAsync: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusAccepted, dtos.StandardResponse{
		Status:  true,
		Message: "Gemini transcription task submitted",
		Data: dtos.TranscriptionTaskResponseDTO{
			TaskID:      taskID,
			TaskStatus:  "pending",
			RecordingID: recording.ID,
		},
	})
}
