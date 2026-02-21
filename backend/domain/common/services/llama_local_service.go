package services

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"teralux_app/domain/common/utils"
)

type LlamaLocalService struct {
	modelPath string
}

func NewLlamaLocalService(cfg *utils.Config) *LlamaLocalService {
	return &LlamaLocalService{
		modelPath: cfg.LlamaLocalModel,
	}
}

func (s *LlamaLocalService) HealthCheck() bool {
	if s.modelPath == "" {
		return false
	}
	if _, err := os.Stat(s.modelPath); os.IsNotExist(err) {
		return false
	}

	// Check if binary exists
	bin := "./bin/llama-cli"
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		_, err := exec.LookPath("llama-cli")
		return err == nil
	}
	return true
}

func (s *LlamaLocalService) CallModel(prompt string, model string) (string, error) {
	if s.modelPath == "" {
		return "", fmt.Errorf("LLAMA_LOCAL_MODEL is not configured")
	}

	// Find llama-cli: try local bin first, then PATH
	bin := "./bin/llama-cli"
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		binInPath, err := exec.LookPath("llama-cli")
		if err != nil {
			return "", fmt.Errorf("llama-cli not found in ./bin or PATH: %w", err)
		}
		bin = binInPath
	}

	// For local service, we use the same model regardless of 'high'/'low'
	// unless we implement multiple local models later.

	args := []string{
		"-m", s.modelPath,
		"-p", prompt,
		"-n", "128", // Limit response length for speed in dev
		"--log-disable", // Reduce noise
		"--simple-io",   // Better compatibility in subprocesses
	}

	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("llama-cli failed: %w - output: %s", err, string(out))
	}

	result := strings.TrimSpace(string(out))

	// llama-cli often repeats the prompt or adds garbage, we might need basic cleaning
	// However, llama-cli usually just outputs the completion after the prompt if not in interactive mode.
	// We'll trust the output for now as a simple local implementation.

	utils.LogDebug("LlamaLocal: Response received (length: %d)", len(result))
	return result, nil
}
