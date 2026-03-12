package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sensio/domain/common/utils"
)

// PythonRAGService handles communication with Python RAG service.
type PythonRAGService struct {
	baseURL    string
	httpClient *http.Client
}

// RAGRequest represents the request to Python RAG service.
type RAGRequest struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"`
	Mac      string `json:"mac_address,omitempty"`
}

// RAGResponse represents the response from Python RAG service.
type RAGResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// NewPythonRAGService creates a new Python RAG service.
func NewPythonRAGService(cfg *utils.Config) *PythonRAGService {
	return &PythonRAGService{
		baseURL:    cfg.PythonRAGServiceURL,
		httpClient: &http.Client{},
	}
}

func (s *PythonRAGService) Translate(req RAGRequest) (*RAGResponse, error) {
	return s.sendRequest("/translate", req)
}

func (s *PythonRAGService) Summary(req RAGRequest) (*RAGResponse, error) {
	return s.sendRequest("/summary", req)
}

func (s *PythonRAGService) Chat(req RAGRequest) (*RAGResponse, error) {
	return s.sendRequest("/chat", req)
}

func (s *PythonRAGService) Control(req RAGRequest) (*RAGResponse, error) {
	return s.sendRequest("/control", req)
}

func (s *PythonRAGService) GetStatus(taskID string) (*RAGResponse, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/%s", s.baseURL, taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result RAGResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

func (s *PythonRAGService) sendRequest(path string, req RAGRequest) (*RAGResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.httpClient.Post(s.baseURL+path, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result RAGResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}
