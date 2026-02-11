package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"teralux_app/domain/recordings/usecases"
)

type RecordingsController struct {
	getAllUseCase  usecases.GetAllRecordingsUseCase
	getByIDUseCase usecases.GetRecordingByIDUseCase
	deleteUseCase  usecases.DeleteRecordingUseCase
	saveUseCase    usecases.SaveRecordingUseCase
}

func NewRecordingsController(
	getAllUseCase usecases.GetAllRecordingsUseCase,
	getByIDUseCase usecases.GetRecordingByIDUseCase,
	deleteUseCase usecases.DeleteRecordingUseCase,
	saveUseCase usecases.SaveRecordingUseCase,
) *RecordingsController {
	return &RecordingsController{
		getAllUseCase:  getAllUseCase,
		getByIDUseCase: getByIDUseCase,
		deleteUseCase:  deleteUseCase,
		saveUseCase:    saveUseCase,
	}
}

// GetAllRecordings godoc
// @Summary      Get all recordings
// @Description  Get a paginated list of recordings
// @Tags         06. Recordings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        page   query     int  false  "Page number" default(1)
// @Param        limit  query     int  false  "Items per page" default(10)
// @Success      200    {object}  dtos.GetAllRecordingsResponseDto
// @Failure      400    {object}  dtos.RecordingStandardResponse
// @Failure      500    {object}  dtos.RecordingStandardResponse
// @Router       /api/recordings [get]
func (c *RecordingsController) GetAllRecordings(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 { page = 1 }
	if limit < 1 { limit = 10 }

	result, err := c.getAllUseCase.Execute(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// GetRecordingByID godoc
// @Summary      Get a recording by ID
// @Description  Get a single recording's metadata by its ID
// @Tags         06. Recordings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Recording ID"
// @Success      200  {object}  dtos.RecordingResponseDto
// @Failure      404  {object}  dtos.RecordingStandardResponse
// @Failure      500  {object}  dtos.RecordingStandardResponse
// @Router       /api/recordings/{id} [get]
func (c *RecordingsController) GetRecordingByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.getByIDUseCase.Execute(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Recording not found"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

// DeleteRecording godoc
// @Summary      Delete a recording
// @Description  Hard delete a recording's metadata and its physical file
// @Tags         06. Recordings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Recording ID"
// @Success      200  {object}  dtos.RecordingStandardResponse
// @Failure      500  {object}  dtos.RecordingStandardResponse
// @Router       /api/recordings/{id} [delete]
func (c *RecordingsController) DeleteRecording(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.deleteUseCase.Execute(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": true, "message": "Recording deleted successfully"})
}

// UploadRecording godoc
// @Summary      Upload a recording
// @Description  Upload an audio file and save its metadata
// @Tags         06. Recordings
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Audio file"
// @Success      201   {object}  dtos.RecordingResponseDto
// @Failure      400   {object}  dtos.RecordingStandardResponse
// @Failure      500   {object}  dtos.RecordingStandardResponse
// @Router       /api/recordings [post]
func (c *RecordingsController) UploadRecording(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	result, err := c.saveUseCase.Execute(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, result)
}
