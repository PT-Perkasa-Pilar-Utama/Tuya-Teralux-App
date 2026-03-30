package controllers

import (
	"net/http"
	"path/filepath"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/models/whisper/dtos"
	"sensio/domain/models/whisper/usecases"
	recordingUsecases "sensio/domain/recordings/usecases"

	"github.com/gin-gonic/gin"
)

type WhisperModelsOpenAIController struct {
	usecase       usecases.TranscribeOpenAIModelUseCase
	saveRecording recordingUsecases.SaveRecordingUseCase
	config        *utils.Config
}

func NewWhisperModelsOpenAIController(
	usecase usecases.TranscribeOpenAIModelUseCase,
	saveRecording recordingUsecases.SaveRecordingUseCase,
	cfg *utils.Config,
) *WhisperModelsOpenAIController {
	return &WhisperModelsOpenAIController{
		usecase:       usecase,
		saveRecording: saveRecording,
		config:        cfg,
	}
}

// Transcribe handles POST /api/whisper/models/openai
// @Summary Transcribe audio file (OpenAI)
// @Description Submit audio file for transcription via OpenAI. Processing is asynchronous.
// @Tags 04. Models
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (.mp3, .wav, .m4a, .aac, .ogg, .flac)"
// @Param language formData string false "Language code (e.g. id, en)"
// @Success 202 {object} commonDtos.StandardResponse{data=dtos.TranscriptionTaskResponseDTO}
// @Failure      400  {object}  commonDtos.ValidationErrorResponse
// @Failure 413 {object} commonDtos.StandardResponse
// @Failure 415 {object} commonDtos.StandardResponse
// @Failure      500  {object}  commonDtos.ErrorResponse
// @Router /api/whisper/models/openai [post]
func (c *WhisperModelsOpenAIController) Transcribe(ctx *gin.Context) {
	file, err := ctx.FormFile("audio")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "audio", Message: "Audio file is required: " + err.Error()},
			},
		})
		return
	}

	if file.Size > c.config.MaxFileSize {
		ctx.JSON(http.StatusRequestEntityTooLarge, commonDtos.StandardResponse{
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
		ctx.JSON(http.StatusUnsupportedMediaType, commonDtos.StandardResponse{
			Status:  false,
			Message: "Unsupported Media Type",
		})
		return
	}

	macAddress := ctx.PostForm("mac_address")
	baseURL := utils.GetBaseURL(ctx)
	recording, err := c.saveRecording.SaveRecording(file, macAddress, baseURL)
	if err != nil {
		utils.LogError("OpenAI.SaveRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	finalPath := filepath.Join("uploads", "audio", recording.Filename)
	language := ctx.PostForm("language")

	taskID, err := c.usecase.TranscribeAsync(finalPath, file.Filename, language, ctx.Request.URL.Path)
	if err != nil {
		utils.LogError("OpenAI.TranscribeAsync: %v", err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusAccepted, commonDtos.StandardResponse{
		Status:  true,
		Message: "OpenAI transcription task submitted",
		Data: dtos.TranscriptionTaskResponseDTO{
			TaskID:      taskID,
			TaskStatus:  "pending",
			RecordingID: recording.ID,
		},
	})
}
