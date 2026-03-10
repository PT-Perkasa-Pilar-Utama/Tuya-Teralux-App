package utils

import (
	"os"
	"path/filepath"
)

// GetAssetPath resolves the absolute path to an asset file.
// Prioritas sumber path:
// 1. env ASSETS_DIR (jika diset),
// 2. fallback /app/assets (runtime container),
// 3. fallback relative project (./assets).
func GetAssetPath(subPath string) string {
	var baseDir string

	if envDir := os.Getenv("ASSETS_DIR"); envDir != "" {
		baseDir = envDir
	} else if _, err := os.Stat("/app/assets"); err == nil {
		baseDir = "/app/assets"
	} else {
		wd, _ := os.Getwd()
		baseDir = filepath.Join(wd, "assets")
	}

	fullPath := filepath.Join(baseDir, subPath)

	// Log warning if the file does not exist to help debugging,
	// but return the resolved path anyway.
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		LogWarn("Asset not found at resolved path: %s (subPath: %s)", fullPath, subPath)
	}

	return fullPath
}
