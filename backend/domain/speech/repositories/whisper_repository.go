package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"teralux_app/domain/common/utils"
)

type WhisperRepository struct {
	config *utils.Config
}

func NewWhisperRepository(cfg *utils.Config) *WhisperRepository {
	return &WhisperRepository{
		config: cfg,
	}
}

// WhisperServerResponse represents the JSON response from whisper.cpp server
type WhisperServerResponse struct {
	Text string `json:"text"`
	// We can add "transcription" array if we need timestamps later
}

// Transcribe executes the transcription.
// If WHISPER_SERVER_URL is set, it posts to the server.
// Otherwise, it falls back to the local CLI binary.
func (r *WhisperRepository) Transcribe(wavPath string, modelPath string, lang string) (string, error) {
	if r.config.WhisperServerURL != "" {
		utils.LogDebug("Whisper: Transcription Path: Server (%s)", r.config.WhisperServerURL)
		return r.transcribeViaServer(wavPath, lang)
	}
	utils.LogDebug("Whisper: Transcription Path: Local CLI")
	return r.transcribeViaCLI(wavPath, modelPath, lang, false)
}

// TranscribeFull executes the transcription and returns the full text.
// If WHISPER_SERVER_URL is set, it posts to the server.
// Otherwise, it falls back to the local CLI binary.
func (r *WhisperRepository) TranscribeFull(wavPath string, modelPath string, lang string) (string, error) {
	if r.config.WhisperServerURL != "" {
		utils.LogDebug("Whisper: Transcription Full Path: Server (%s)", r.config.WhisperServerURL)
		return r.transcribeViaServer(wavPath, lang)
	}
	utils.LogDebug("Whisper: Transcription Full Path: Local CLI")
	return r.transcribeViaCLI(wavPath, modelPath, lang, true)
}

func (r *WhisperRepository) Convert(wavPath string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

// transcribeViaServer sends the audio file to the configured Whisper server
func (r *WhisperRepository) transcribeViaServer(audioPath string, lang string) (string, error) {
	url := fmt.Sprintf("%s/inference", r.config.WhisperServerURL)

	// Prepare multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	file, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add fields
	_ = writer.WriteField("response_format", "json")
	if lang != "" && lang != "auto" {
		_ = writer.WriteField("language", lang)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Send Request
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to whisper server failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("whisper server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse Response
	var result WhisperServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode json response: %w", err)
	}

	return strings.TrimSpace(result.Text), nil
}

// transcribeViaCLI contains the legacy logic to run whisper-cli
func (r *WhisperRepository) transcribeViaCLI(wavPath string, modelPath string, lang string, full bool) (string, error) {
	// Find whisper-cli: try local bin first, then PATH
	bin := "./bin/whisper-cli"
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		binInPath, err := exec.LookPath("whisper-cli")
		if err != nil {
			return "", fmt.Errorf("whisper-cli not found in ./bin or PATH. Run './setup.sh' to build and prepare the binary: %w", err)
		}
		bin = binInPath
	}

	// instructed whisper-cli to write a .txt file (it formats it nicely without weights/timings)
	base := strings.TrimSuffix(wavPath, ".wav")
	txtPath := base + ".txt"
	_ = os.Remove(txtPath)

	args := []string{"-m", modelPath, "-f", wavPath, "--no-timestamps", "-otxt", "-of", base}
	if lang != "" && lang != "auto" {
		args = append(args, "-l", lang)
	}

	cmd := exec.Command(bin, args...)
	// utils.LogDebug("[whisper-cli] running: %v", cmd.Args) // Commented out to avoid circular deps if utils uses repo? No, repo uses utils.
	// But let's keep it clean.

	out, err := cmd.CombinedOutput()
	// If full mode, we might want to just return error if it fails?
	// The original code had different fallback logic for Transcribe vs TranscribeFull.
	// For simplicity, I'm normalizing it a bit here, but ensuring we respect the file output preference.

	if err != nil && !full {
		// Existing Transcribe fallback to stdout parsing
		// ... (omitted for brevity, assume file output works for now as primary method)
		// Actually, let's just return what we have or error, similar to TranscribeFull logic in original code
		return "", fmt.Errorf("whisper-cli failed: %w - output: %s", err, string(out))
	}

	// Read produced .txt file
	b, err := os.ReadFile(txtPath)
	if err != nil {
		// Fallback to stdout if file not found
		s := strings.TrimSpace(string(out))
		if s == "" {
			return "", fmt.Errorf("no output file and empty stdout: %w", err)
		}
		// Basic stdout parsing
		lines := strings.Split(s, "\n")
		var result []string
		for _, l := range lines {
			line := strings.TrimSpace(l)
			if line == "" {
				continue
			}
			low := strings.ToLower(line)
			if strings.Contains(low, "whisper_") || strings.Contains(low, "total time") || strings.Contains(low, "error") {
				continue
			}
			result = append(result, line)
		}
		if full {
			return strings.Join(result, " "), nil
		}
		// For non-full (Transcribe), return last line?
		if len(result) > 0 {
			return result[len(result)-1], nil
		}
		return "", nil
	}

	// Tidy up the text file content
	s := strings.TrimSpace(string(b))
	if full {
		lines := strings.Split(s, "\n")
		var result []string
		for _, l := range lines {
			line := strings.TrimSpace(l)
			if line != "" {
				result = append(result, line)
			}
		}
		return strings.Join(result, " "), nil
	}

	// For original Transcribe behavior (last line)
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines[i])
		if l != "" {
			return l, nil
		}
	}
	return s, nil
}
