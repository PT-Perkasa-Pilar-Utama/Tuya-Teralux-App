package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"teralux_app/domain/common/utils"
	recordingUsecases "teralux_app/domain/recordings/usecases"
	"teralux_app/domain/speech/dtos"
	"teralux_app/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

// SpeechTranscribeWhisperCppController handles POST /api/speech/transcribe/whisper/cpp.
type SpeechTranscribeWhisperCppController struct {
	transcribeUC    usecases.TranscribeWhisperCppUseCase
	saveRecordingUC recordingUsecases.SaveRecordingUseCase
	config          *utils.Config
}

func NewSpeechTranscribeWhisperCppController(
	transcribeUC usecases.TranscribeWhisperCppUseCase,
	saveRecordingUC recordingUsecases.SaveRecordingUseCase,
	cfg *utils.Config,
) *SpeechTranscribeWhisperCppController {
	return &SpeechTranscribeWhisperCppController{
		transcribeUC:    transcribeUC,
		saveRecordingUC: saveRecordingUC,
		config:          cfg,
	}
}

// TranscribeWhisperCpp handles POST /api/speech/transcribe/whisper/cpp
// @Summary Transcribe audio file (Whisper.cpp)
// @Description Start transcription of audio file using Whisper.cpp. Asynchronous processing with background execution. Pure Whisper.cpp, no Orion. Supports: .mp3, .wav, .m4a, .aac, .ogg, .flac.
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
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/speech/transcribe/whisper/cpp [post]
func (c *SpeechTranscribeWhisperCppController) TranscribeWhisperCpp(ctx *gin.Context) {
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

	lang := ctx.PostForm("language")
	if lang == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "language", Message: "Language code is required (e.g. 'id', 'en')"},
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

	tempDir := "./tmp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		if err := os.Mkdir(tempDir, 0755); err != nil {
			utils.LogError("TranscribeWhisperCpp: Failed to create temp directory: %v", err)
		}
	}

	inputPath := filepath.Join(tempDir, "whisper_cpp_"+file.Filename)
	if err := ctx.SaveUploadedFile(file, inputPath); err != nil {
		utils.LogError("TranscribeWhisperCpp.SaveUploadedFile: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	recording, err := c.saveRecordingUC.SaveRecording(file)
	if err != nil {
		utils.LogError("TranscribeWhisperCpp.SaveRecording: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	finalInputPath := filepath.Join("uploads", "audio", recording.Filename)
	taskID, err := c.transcribeUC.TranscribeWhisperCpp(finalInputPath, file.Filename, lang)
	if err != nil {
		utils.LogError("TranscribeWhisperCpp.TranscribeWhisperCpp: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
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
