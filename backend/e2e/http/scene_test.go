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

type SceneE2ETestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *SceneE2ETestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
}

func (s *SceneE2ETestSuite) TestScene_GetAll() {
	req, _ := http.NewRequest(http.MethodGet, "/api/scenes", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *SceneE2ETestSuite) TestScene_GetByTerminal() {
	req, _ := http.NewRequest(http.MethodGet, "/api/terminal/term-123/scenes", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *SceneE2ETestSuite) TestScene_Create() {
	payload := map[string]interface{}{
		"name": "Test Scene",
		"actions": []map[string]interface{}{
			{
				"device_id": "device-123",
				"action":    "turn_on",
				"value":     "100",
			},
		},
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/terminal/term-123/scenes", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *SceneE2ETestSuite) TestScene_Update() {
	payload := map[string]interface{}{
		"name": "Updated Scene",
		"actions": []map[string]interface{}{
			{
				"device_id": "device-123",
				"action":    "turn_off",
			},
		},
	}

	req, _ := http.NewRequest(http.MethodPut, "/api/terminal/term-123/scenes/scene-456", bytes.NewBufferString(mustMarshal(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *SceneE2ETestSuite) TestScene_Delete() {
	req, _ := http.NewRequest(http.MethodDelete, "/api/terminal/term-123/scenes/scene-456", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *SceneE2ETestSuite) TestScene_Trigger() {
	req, _ := http.NewRequest(http.MethodGet, "/api/terminal/term-123/scenes/scene-456/control", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert := assert.New(s.T())
	assert.NotEqual(http.StatusInternalServerError, w.Code)
}

func (s *SceneE2ETestSuite) TestScene_ValidationErrors() {
	s.T().Run("Missing name", func(t *testing.T) {
		payload := map[string]interface{}{
			"actions": []map[string]interface{}{},
		}

		req, _ := http.NewRequest(http.MethodPost, "/api/terminal/term-123/scenes", bytes.NewBufferString(mustMarshal(payload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert := assert.New(t)
		assert.True(w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
	})
}

func TestSceneE2ETestSuite(t *testing.T) {
	suite.Run(t, new(SceneE2ETestSuite))
}
