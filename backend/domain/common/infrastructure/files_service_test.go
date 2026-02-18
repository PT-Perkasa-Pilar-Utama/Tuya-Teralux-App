package infrastructure

import (
	"bytes"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileService_EnsureDir(t *testing.T) {
	service := NewFileService()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "test_subdir")

	// Test creating new directory
	err := service.EnsureDir(targetDir)
	assert.NoError(t, err)
	info, err := os.Stat(targetDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Test existing directory
	err = service.EnsureDir(targetDir)
	assert.NoError(t, err)
}

func TestFileService_SaveUploadedFile(t *testing.T) {
	service := NewFileService()
	tmpDir := t.TempDir()

	// 1. Create content
	content := []byte("hello world")

	// 2. Construct multipart.FileHeader
    body := new(bytes.Buffer)
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.txt")
    assert.NoError(t, err)
    _, err = part.Write(content)
    assert.NoError(t, err)
    writer.Close()
    
    // Parse the multipart body to get a valid FileHeader
    req := multipart.NewReader(body, writer.Boundary())
    form, err := req.ReadForm(1024)
    assert.NoError(t, err)
    defer form.RemoveAll()
    
    fileHeader := form.File["file"][0]

	// 3. Define destination
	dst := filepath.Join(tmpDir, "saved", "test.txt")

	// 4. Test Save
	err = service.SaveUploadedFile(fileHeader, dst)
	assert.NoError(t, err)

	// 5. Verify Content
	savedContent, err := os.ReadFile(dst)
	assert.NoError(t, err)
	assert.Equal(t, content, savedContent)
}
