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
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
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
