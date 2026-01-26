package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	return string(body), nil
}
