package infrastructure

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// FileService defines operations for file handling
type FileService interface {
	SaveUploadedFile(file *multipart.FileHeader, dst string) error
	SaveFile(data []byte, dst string) error
	MoveFile(src, dst string) error
	DeleteFile(path string) error
	EnsureDir(dirName string) error
}

type fileService struct{}

func NewFileService() FileService {
	return &fileService{}
}

// SaveUploadedFile saves a multipart file to the specified destination path.
func (s *fileService) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, src)
	return err
}

// SaveFile saves raw bytes to the specified destination path.
func (s *fileService) SaveFile(data []byte, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// MoveFile moves a file from src to dst.
func (s *fileService) MoveFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	// Try atomic rename first
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Fallback for cross-device move (EXDEV)
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source for move: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination for move: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy during move: %v", err)
	}

	// Explicit fsync for security
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination: %v", err)
	}

	// Success, we can remove the source
	sourceFile.Close() // Close before removal
	return os.Remove(src)
}

// DeleteFile deletes a file.
func (s *fileService) DeleteFile(path string) error {
	return os.Remove(path)
}

// EnsureDir ensures that a directory exists, creating it if necessary.
func (s *fileService) EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0755)
	if err == nil || os.IsExist(err) {
		return nil
	}
	return fmt.Errorf("failed to create directory %s: %v", dirName, err)
}

// Helper for direct usage if needed, but discouraged in favor of DI
var DefaultFileService = NewFileService()

func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	return DefaultFileService.SaveUploadedFile(file, dst)
}
