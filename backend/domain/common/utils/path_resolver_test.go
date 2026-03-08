package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetAssetPath(t *testing.T) {
	// Save original env and working directory
	origAssetsDir := os.Getenv("ASSETS_DIR")
	defer os.Setenv("ASSETS_DIR", origAssetsDir)

	wd, _ := os.Getwd()

	tests := []struct {
		name      string
		assetsDir string
		subPath   string
		want      string
	}{
		{
			name:      "Use ASSETS_DIR env",
			assetsDir: "/tmp/custom_assets",
			subPath:   "images/logo.png",
			want:      "/tmp/custom_assets/images/logo.png",
		},
		{
			name:      "Fallback to ./assets (default)",
			assetsDir: "",
			subPath:   "images/logo.png",
			want:      filepath.Join(wd, "assets", "images", "logo.png"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.assetsDir != "" {
				os.Setenv("ASSETS_DIR", tt.assetsDir)
			} else {
				os.Unsetenv("ASSETS_DIR")
			}

			// We skip testing the /app/assets fallback because it's environment dependent
			// (unless we are running in the container)
			if tt.assetsDir == "" {
				// Ensure we don't accidentally match /app/assets if it exists on the host
				if _, err := os.Stat("/app/assets"); err == nil {
					t.Skip("Skipping fallback test because /app/assets exists on host")
				}
			}

			got := GetAssetPath(tt.subPath)
			if got != tt.want {
				t.Errorf("GetAssetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
