package usecases

import (
	"testing"

	"teralux_app/domain/common/infrastructure"
	speechRepos "teralux_app/domain/speech/repositories"
	"teralux_app/domain/common/utils"
)

func TestRAGUsecase_ProcessAndGetStatus(t *testing.T) {
	vectorSvc := infrastructure.NewVectorService()
	ollama := speechRepos.NewOllamaRepository()
	cfg := &utils.Config{}
	u := NewRAGUsecase(vectorSvc, ollama, cfg)

	task, err := u.Process("some text")
	if err != nil {
		t.Fatalf("expected no error from Process, got %v", err)
	}
	if task == "" {
		t.Fatalf("expected non-empty task id")
	}

	status, err := u.GetStatus(task)
	if err != nil {
		t.Fatalf("expected no error from GetStatus, got %v", err)
	}
	if status == nil || status.Status == "" {
		t.Fatalf("expected valid status result, got %+v", status)
	}
}