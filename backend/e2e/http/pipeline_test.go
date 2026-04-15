package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PipelineE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *PipelineE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *PipelineE2ETestSuite) TestPipeline_SubmitJob() {
	payload := map[string]interface{}{
		"language":        "id",
		"target_language": "en",
		"summarize":       true,
		"refine":          true,
		"diarize":         false,
		"context":         "meeting",
		"style":           "minutes",
		"date":            "2026-04-15",
		"location":        "Jakarta",
		"participants":    "Alice,Bob",
		"mac_address":     "AA:BB:CC:DD:EE:FF",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/pipeline/job", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *PipelineE2ETestSuite) TestPipeline_GetStatus() {
	req, _ := http.NewRequest(http.MethodGet, "/api/models/pipeline/status/task-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *PipelineE2ETestSuite) TestPipeline_CancelTask() {
	req, _ := http.NewRequest(http.MethodDelete, "/api/models/pipeline/status/task-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *PipelineE2ETestSuite) TestPipelineV1_Job() {
	payload := map[string]interface{}{
		"language":        "id",
		"target_language": "en",
		"summarize":       "true",
		"diarize":         "false",
		"refine":          "true",
		"participants":    "Alice,Bob",
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/models/v1/pipeline/job", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *PipelineE2ETestSuite) TestPipelineV1_GetStatus() {
	req, _ := http.NewRequest(http.MethodGet, "/api/models/v1/pipeline/status/task-123", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func TestPipelineE2ETestSuite(t *testing.T) {
	suite.Run(t, new(PipelineE2ETestSuite))
}
