package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"teralux_app/domain/rag/dtos"
	"github.com/gin-gonic/gin"
)

// fake usecase implementation for tests
type fakeUsecase struct{
	taskID string
}

func (f *fakeUsecase) Process(text string) (string, error) {
	return f.taskID, nil
}

func (f *fakeUsecase) GetStatus(taskID string) (*dtos.RAGStatusDTO, error) {
	return &dtos.RAGStatusDTO{Status: "pending", Result: ""}, nil
}

func TestProcessTextReturns202(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &fakeUsecase{taskID: "test-uuid"}
	controller := NewRAGController(fake)

	r := gin.New()
	r.POST("/api/rag", controller.ProcessText)

	body := map[string]string{"text": "turn on the lamp"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/rag", bytes.NewBuffer(b))
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
}
