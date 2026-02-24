package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"teralux_app/domain/common/utils"
	"time"
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

	args := []string{
		"-m", s.modelPath,
		"-p", prompt,
		"-n", "64", // Moderate length
		"--single-turn",
		"--simple-io",
		"--log-disable",
		"--no-display-prompt",
		"--color", "off",
		"--log-colors", "off",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second) // Increased timeout for loading
	defer cancel()

	utils.LogDebug("LlamaLocal: Running %s (single-turn)", bin)

	cmd := exec.CommandContext(ctx, bin, args...)
	
	// Force non-interactive environment
	cmd.Env = append(os.Environ(), "TERM=dumb")
	
	// Capture BOTH stdout and stderr to ensure absolutely nothing leaks to the terminal
	// and we can clean the whole stream.
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			utils.LogError("LlamaLocal: Execution timed out")
			return "", fmt.Errorf("llama-cli timed out after 90s")
		}
		// If it failed, don't show the noisy output unless in debug
		utils.LogDebug("LlamaLocal: raw failure output: %s", string(out))
		return "", fmt.Errorf("llama-cli failed: %w", err)
	}

	rawOutput := string(out)
	
	// Robust parsing:
	// The output looks like: [Junk] \n > [Prompt] \n [Result] \n [Junk/Metrics]
	// We look for the last "> " followed by the prompt (if echoed) or just the last "> ".
	
	result := rawOutput
	
	// 1. Find the content after the last "> " prompt marker
	lastPromptIdx := strings.LastIndex(result, "> ")
	if lastPromptIdx != -1 {
		result = result[lastPromptIdx+2:]
	}

	// 2. Remove any prompt echo if it exists immediately after the marker
	trimmedPrompt := strings.TrimSpace(prompt)
	if strings.HasPrefix(strings.TrimSpace(result), trimmedPrompt) {
		result = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(result), trimmedPrompt))
	}

	// 3. Cut off at metrics/logs that start with "llama_" or bracketed timings
	if idx := strings.Index(result, "llama_"); idx != -1 {
		result = result[:idx]
	}
	if idx := strings.Index(result, "[ Prompt:"); idx != -1 {
		result = result[:idx]
	}
	if idx := strings.Index(result, "Exiting..."); idx != -1 {
		result = result[:idx]
	}

	result = strings.TrimSpace(result)

	utils.LogDebug("LlamaLocal: Processed result (length: %d)", len(result))
	return result, nil
}
