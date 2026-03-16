package controllers

import (
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/models/whisper/dtos"
	"sensio/domain/models/whisper/usecases"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UploadSessionController struct {
	useCase      usecases.UploadSessionUseCase
	transcribeUC usecases.TranscribeUseCase
}

func NewUploadSessionController(useCase usecases.UploadSessionUseCase, transcribeUC usecases.TranscribeUseCase) *UploadSessionController {
	return &UploadSessionController{
		useCase:      useCase,
		transcribeUC: transcribeUC,
	}
}

func (c *UploadSessionController) CreateSession(ctx *gin.Context) {
	var req dtos.CreateUploadSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Invalid request payload",
			Details: err.Error(),
		})
		return
	}

	// Phase 3: Ownership enforcement - passing UID from context
	uid, _ := ctx.Get("uid")
	if uidStr, ok := uid.(string); ok {
		req.OwnerUID = uidStr
	}

	resp, err := c.useCase.CreateSession(req)
	if err != nil {
		ctx.JSON(mapUploadErrToStatus(err), commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to create upload session",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, commonDtos.StandardResponse{
		Status:  true,
		Message: "Upload session created successfully",
		Data:    resp,
	})
}

func (c *UploadSessionController) UploadChunk(ctx *gin.Context) {
	sessionID := ctx.Param("id")
	indexStr := ctx.Param("index")

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Invalid chunk index",
		})
		return
	}

	// Log request metadata for observability
	contentLength := ctx.Request.ContentLength
	utils.LogInfo("UploadChunk: request received for session %s chunk %d (Content-Length: %d bytes, client: %s)",
		sessionID, index, contentLength, ctx.ClientIP())

	// Phase 3: Ownership check - passing UID from context
	uid, _ := ctx.Get("uid")
	uidStr, _ := uid.(string)

	resp, err := c.useCase.UploadChunk(sessionID, index, uidStr, ctx.Request.Body)
	if err != nil {
		ctx.JSON(mapUploadErrToStatus(err), commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to upload chunk",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Chunk uploaded successfully",
		Data:    resp,
	})
}

func (c *UploadSessionController) GetSessionStatus(ctx *gin.Context) {
	sessionID := ctx.Param("id")

	// Phase 3: Ownership check
	uid, _ := ctx.Get("uid")
	uidStr, _ := uid.(string)

	resp, err := c.useCase.GetSessionStatus(sessionID, uidStr)
	if err != nil {
		ctx.JSON(mapUploadErrToStatus(err), commonDtos.StandardResponse{
			Status:  false,
			Message: "Failed to get session status",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
		Status:  true,
		Message: "Session status retrieved successfully",
		Data:    resp,
	})
}

func mapUploadErrToStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	msg := err.Error()

	// 408 Request Timeout - interrupted upload body
	// This indicates the upload body was interrupted during transmission
	if strings.HasPrefix(msg, "upload interrupted") {
		return http.StatusRequestTimeout
	}

	// 404 Not Found - session does not exist
	if msg == "session not found" {
		return http.StatusNotFound
	}

	// 403 Forbidden - ownership/auth violation
	// This is a non-recoverable auth error that client should surface
	if msg == "unauthorized session access" {
		return http.StatusForbidden
	}

	// 409 Conflict - session state conflicts with requested operation
	// Includes: wrong state, session consumed, session already expired,
	// upload session invalidated, chunk size mismatch, or session not ready
	// (session not ready means chunks are incomplete - a conflict state)
	if msg == "session not in uploading state" ||
		msg == "session conflict" ||
		msg == "session consumed" ||
		msg == "session already expired" ||
		strings.HasPrefix(msg, "session not ready") ||
		strings.HasPrefix(msg, "upload session invalidated due to incomplete chunk data") ||
		strings.HasPrefix(msg, "chunk size mismatch:") {
		return http.StatusConflict
	}

	// 400 Bad Request - validation errors
	if msg == "file_name is required" ||
		msg == "total_size_bytes must be greater than 0" ||
		msg == "invalid chunk index" ||
		strings.HasPrefix(msg, "invalid chunk index") ||
		strings.HasPrefix(msg, "file size exceeds maximum allowed") ||
		strings.HasPrefix(msg, "chunk size exceeds") {
		return http.StatusBadRequest
	}

	// Default 500 Internal Server Error for unexpected errors
	return http.StatusInternalServerError
}
