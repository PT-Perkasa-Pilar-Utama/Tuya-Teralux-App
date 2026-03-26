package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sensio/domain/common/utils"
)

// PythonPipelineService handles communication with Python Pipeline service.
type PythonPipelineService struct {
	baseURL    string
	httpClient *http.Client
}

// PipelineRequest represents the request to Python Pipeline service.
type PipelineRequest struct {
	AudioPath      string   `json:"audio_path"`
	Language       string   `json:"language"`
	TargetLanguage string   `json:"target_language"`
	Context        string   `json:"context,omitempty"`
	Style          string   `json:"style,omitempty"`
	Date           string   `json:"date,omitempty"`
	Location       string   `json:"location,omitempty"`
	Participants   []string `json:"participants,omitempty"`
	Diarize        bool     `json:"diarize"`
	Refine         *bool    `json:"refine,omitempty"`
	Summarize      bool     `json:"summarize"`
}

// PipelineResponse represents the response from Python Pipeline service.
type PipelineResponse struct {
	TaskID      string `json:"task_id"`
	Status      string `json:"status"`
	Transcript  string `json:"transcript,omitempty"`
	RefinedText string `json:"refined_text,omitempty"`
	Translated  string `json:"translated,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Error       string `json:"error,omitempty"`
}

// NewPythonPipelineService creates a new Python Pipeline service.
func NewPythonPipelineService(cfg *utils.Config) *PythonPipelineService {
	return &PythonPipelineService{
		baseURL:    cfg.PythonPipelineServiceURL,
		httpClient: &http.Client{},
	}
}

// ExecuteJob sends audio file to Python service for pipeline processing.
func (s *PythonPipelineService) ExecuteJob(req PipelineRequest) (*PipelineResponse, error) {
	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Open audio file if path is provided
	if req.AudioPath != "" {
		file, err := os.Open(req.AudioPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open audio file: %w", err)
		}
		defer file.Close()

		part, err := writer.CreateFormFile("audio", req.AudioPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %w", err)
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, fmt.Errorf("failed to copy file: %w", err)
		}
	}

	// Add other fields
	addField := func(key, value string) error {
		if value != "" {
			return writer.WriteField(key, value)
		}
		return nil
	}

	_ = addField("language", req.Language)
	_ = addField("target_language", req.TargetLanguage)
	_ = addField("context", req.Context)
	_ = addField("style", req.Style)
	_ = addField("date", req.Date)
	_ = addField("location", req.Location)
	_ = addField("diarize", fmt.Sprintf("%v", req.Diarize))
	_ = addField("summarize", fmt.Sprintf("%v", req.Summarize))

	if req.Refine != nil {
		_ = writer.WriteField("refine", fmt.Sprintf("%v", *req.Refine))
	}

	if len(req.Participants) > 0 {
		for _, p := range req.Participants {
			_ = writer.WriteField("participants", p)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create HTTP request
	reqHTTP, err := http.NewRequest("POST", s.baseURL+"/job", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := s.httpClient.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result PipelineResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetStatus gets pipeline task status from Python service.
func (s *PythonPipelineService) GetStatus(taskID string) (*PipelineResponse, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("%s/status/%s", s.baseURL, taskID))
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result PipelineResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}
