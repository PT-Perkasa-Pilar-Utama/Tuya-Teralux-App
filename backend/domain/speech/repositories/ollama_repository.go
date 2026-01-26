package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type OllamaRepository struct {
	baseURL string
}

func NewOllamaRepository() *OllamaRepository {
	return &OllamaRepository{baseURL: "http://localhost:11434"}
}

func (r *OllamaRepository) CallModel(prompt string, model string) (string, error) {
	reqBody := map[string]interface{}{"model": model, "prompt": prompt}
	b, _ := json.Marshal(reqBody)
	resp, err := http.Post(r.baseURL+"/api/generate", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ollama error: %s", string(body))
	}

	// Try to parse Ollama response to extract the generated text
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err == nil {
		if results, ok := m["results"].([]interface{}); ok && len(results) > 0 {
			var outText strings.Builder
			for _, r := range results {
				if rm, ok := r.(map[string]interface{}); ok {
					if contents, ok := rm["content"].([]interface{}); ok {
						for _, c := range contents {
							if cm, ok := c.(map[string]interface{}); ok {
								if txt, ok := cm["text"].(string); ok {
									outText.WriteString(txt)
								}
							}
						}
					}
				}
			}
			if outText.Len() > 0 {
				return outText.String(), nil
			}
		}
	}

	// Fallback: return raw body as string
	return string(body), nil
}
