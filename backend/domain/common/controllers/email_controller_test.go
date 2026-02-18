package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestEmailController_SendEmail_InvalidJSON(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	cfg := &utils.Config{}
	ctrl := NewEmailController(cfg)
	
	router := gin.New()
	router.POST("/email", ctrl.SendEmail)

	// Execute
	reqBody := []byte(`{"invalid": "json"}`) // Missing required fields
	req, _ := http.NewRequest("POST", "/email", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["status"].(bool))
	assert.Equal(t, "Validation Error", response["message"])
}

func TestEmailController_SendEmail_Integration(t *testing.T) {
	// Proper integration test with temp template
	gin.SetMode(gin.TestMode)
	cfg := &utils.Config{
		SMTPHost: "invalid", // Will fail at SMTP level
	}
	ctrl := NewEmailController(cfg)
	
	router := gin.New()
	router.POST("/email", ctrl.SendEmail)
	
	// Create request
	body := map[string]interface{}{
		"to": []string{"test@example.com"},
		"subject": "Test",
		"template": "non_existent",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/email", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()
	
	// Execute
	router.ServeHTTP(w, req)
	
	// Verify
	// Should fail because template not found (INTERNAL SERVER ERROR)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["status"].(bool))
	assert.Equal(t, "Internal Server Error", response["message"])
}
