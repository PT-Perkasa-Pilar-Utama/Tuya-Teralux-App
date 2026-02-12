package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"

	"github.com/gin-gonic/gin"
)

// fake usecase implementation for tests
type fakeUsecase struct {
	taskID    string
	expiresIn int64
}

func (f *fakeUsecase) Control(text string, authToken string, onComplete func(string, *dtos.RAGStatusDTO)) (string, error) {
	return f.taskID, nil
}

func (f *fakeUsecase) GetStatus(taskID string) (*dtos.RAGStatusDTO, error) {
	return &dtos.RAGStatusDTO{Status: "pending", Result: "", ExpiresInSecond: f.expiresIn, ExpiresAt: time.Now().Add(time.Duration(f.expiresIn) * time.Second).UTC().Format(time.RFC3339)}, nil
}

func (f *fakeUsecase) Translate(text string, language string) (string, error) {
	return "Translated: " + text, nil
}

func (f *fakeUsecase) TranslateAsync(text string, language string) (string, error) {
	return f.taskID, nil
}

func (f *fakeUsecase) Summary(text string, language string, context string, style string) (*dtos.RAGSummaryResponseDTO, error) {
	return &dtos.RAGSummaryResponseDTO{Summary: "Summary: " + text, PDFUrl: "http://example.com/pdf"}, nil
}

func (f *fakeUsecase) SummaryAsync(text string, language string, context string, style string) (string, error) {
	return f.taskID, nil
}

func TestControlReturns202(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &fakeUsecase{taskID: "test-uuid", expiresIn: 3600}
	cfg := &utils.Config{}
	controller := NewRAGController(fake, cfg)

	r := gin.New()
	r.POST("/api/rag/control", controller.Control)

	body := map[string]string{"text": "turn on the lamp"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/rag/control", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", w.Code)
	}

	var resp dtos.StandardResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected response Data to be map, got %+v", resp.Data)
	}
	if data["task_id"] != "test-uuid" {
		t.Fatalf("expected task_id test-uuid, got %v", data["task_id"])
	}
	// Check status DTO presence and TTL via DTO (no hardcoded values in controller)
	statusRaw, ok := data["task_status"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected status to be present in response Data, got %+v", data["status"])
	}
	if int64(statusRaw["expires_in_seconds"].(float64)) != fake.expiresIn {
		t.Fatalf("expected expires_in_seconds %d, got %v", fake.expiresIn, statusRaw["expires_in_seconds"])
	}
}
