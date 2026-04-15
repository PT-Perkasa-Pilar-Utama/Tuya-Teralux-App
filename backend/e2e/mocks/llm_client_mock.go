package mocks

import (
	"context"
	"sensio/domain/models/rag/skills"
	"sensio/domain/models/whisper/dtos"
)

type MockWhisperClient struct {
	TranscribeFunc func(ctx context.Context, audioPath string, language string, diarize bool) (*dtos.WhisperResult, error)
}

func (m *MockWhisperClient) Transcribe(ctx context.Context, audioPath string, language string, diarize bool) (*dtos.WhisperResult, error) {
	if m.TranscribeFunc != nil {
		return m.TranscribeFunc(ctx, audioPath, language, diarize)
	}
	return &dtos.WhisperResult{
		Transcription:    "mocked transcription",
		DetectedLanguage: "id",
	}, nil
}

type MockLLMClient struct {
	CallModelFunc func(ctx context.Context, prompt string, model string) (string, error)
}

func (m *MockLLMClient) CallModel(ctx context.Context, prompt string, model string) (string, error) {
	if m.CallModelFunc != nil {
		return m.CallModelFunc(ctx, prompt, model)
	}
	return "mocked LLM response", nil
}

type MockLLMFactory struct {
	Clients map[string]skills.LLMClient
}

func (f *MockLLMFactory) GetClient(provider string) skills.LLMClient {
	if f.Clients == nil {
		f.Clients = make(map[string]skills.LLMClient)
	}
	if client, ok := f.Clients[provider]; ok {
		return client
	}
	return &MockLLMClient{}
}
