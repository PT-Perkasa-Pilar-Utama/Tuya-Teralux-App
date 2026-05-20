package usecases

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sensio/domain/common/utils"
)

type resolvedAttachment struct {
	path        *string
	downloadURL string
	cleanup     func()
}

func isReportDownloadURL(path string) bool {
	return strings.Contains(path, "/api/reports/")
}

func resolveMailAttachment(rawPath *string) resolvedAttachment {
	if rawPath == nil {
		return resolvedAttachment{}
	}

	trimmed := strings.TrimSpace(*rawPath)
	if trimmed == "" {
		return resolvedAttachment{}
	}

	if isReportDownloadURL(trimmed) {
		utils.LogDebug("Mail attachment resolved to report download URL %s", trimmed)
		return resolvedAttachment{downloadURL: trimmed}
	}

	if localPath, ok := resolveUploadAttachmentPath(trimmed); ok {
		utils.LogDebug("Mail attachment resolved to local path %s", localPath)
		return resolvedAttachment{path: &localPath}
	}

	parsedURL, err := url.Parse(trimmed)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		utils.LogWarn("Mail attachment unavailable: unsupported path %q", trimmed)
		return resolvedAttachment{}
	}

	utils.LogDebug("Mail attachment resolved to remote URL %s", trimmed)
	return resolvedAttachment{downloadURL: trimmed}
}

func resolveUploadAttachmentPath(rawPath string) (string, bool) {
	relPath := extractUploadsRelativePath(rawPath)
	if relPath == "" {
		return "", false
	}

	baseDir, err := resolveBackendBaseDir()
	if err != nil {
		utils.LogWarn("Mail attachment base dir resolve failed: %v", err)
		return "", false
	}

	fullPath := filepath.Join(baseDir, relPath)
	if _, err := os.Stat(fullPath); err != nil {
		return "", false
	}

	return fullPath, true
}

func extractUploadsRelativePath(rawPath string) string {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "/uploads/") {
		return strings.TrimPrefix(trimmed, "/")
	}

	parsedURL, err := url.Parse(trimmed)
	if err != nil {
		return ""
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ""
	}

	if idx := strings.Index(parsedURL.Path, "/uploads/"); idx != -1 {
		return strings.TrimPrefix(parsedURL.Path[idx:], "/")
	}

	return ""
}

func resolveBackendBaseDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	baseDir := wd
	if !strings.HasSuffix(wd, "backend") {
		if _, err := os.Stat(filepath.Join(wd, "backend")); err == nil {
			baseDir = filepath.Join(wd, "backend")
		}
	}

	return baseDir, nil
}

func downloadAttachmentToTemp(rawURL string) (string, func(), error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	fileExt := filepath.Ext(resp.Request.URL.Path)
	if fileExt == "" {
		fileExt = filepath.Ext(rawURL)
	}

	tempFile, err := os.CreateTemp("", "mail-attachment-*"+fileExt)
	if err != nil {
		return "", nil, err
	}

	tempPath := tempFile.Name()
	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return "", nil, err
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", nil, err
	}

	cleanup := func() {
		if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
			utils.LogWarn("Mail attachment temp cleanup failed for %s: %v", tempPath, err)
		}
	}

	return tempPath, cleanup, nil
}
