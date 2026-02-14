package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"teralux_app/domain/common/utils"
)

type OllamaRepository struct {
	baseURL string
}

func tryDial(host, port string) error {
	addr := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

func NewOllamaRepository() *OllamaRepository {
	// Default to local Ollama if no specific base URL is provided in the future
	base := "http://localhost:11434"

	// If configured host is localhost/127.0.0.1 but not reachable from container,
	// try host.docker.internal:PORT as fallback (useful when running inside Docker).
	if u, err := url.Parse(base); err == nil {
		host := u.Hostname()
		port := u.Port()
		if port == "" {
			if u.Scheme == "http" {
				port = "80"
			} else if u.Scheme == "https" {
				port = "443"
			}
		}
		if (host == "localhost" || host == "127.0.0.1") && tryDial(host, port) != nil {
			alt := "host.docker.internal"
			if tryDial(alt, port) == nil {
				u.Host = alt + ":" + port
				base = u.String()
				utils.LogInfo("OLLAMA host %s not reachable from container; switching to %s", host, alt)
			}
		}
	}

	return &OllamaRepository{baseURL: base}
}

func (r *OllamaRepository) CallModel(prompt string, model string) (string, error) {
	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}
	b, _ := json.Marshal(reqBody)
	resp, err := http.Post(r.baseURL+"/api/generate", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ollama error: %s", string(body))
	}

	// Try to parse Ollama response to extract the generated text
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err == nil {
		// Ollama /api/generate non-stream returns the response in the "response" field
		if responseText, ok := m["response"].(string); ok {
			return responseText, nil
		}
	}

	// Fallback: return raw body as string
	return string(body), nil
}
