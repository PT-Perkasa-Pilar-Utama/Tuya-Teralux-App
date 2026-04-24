package crypto

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yeka/zip"
)

type ZIPEncryptor struct{}

func NewZIPEncryptor() *ZIPEncryptor {
	return &ZIPEncryptor{}
}

// EncryptFiles creates an AES-256 encrypted zip archive containing all provided input files.
func (e *ZIPEncryptor) EncryptFiles(outputZipPath string, inputFiles []string, password string) error {
	if outputZipPath == "" {
		return fmt.Errorf("output zip path is required")
	}
	if len(inputFiles) == 0 {
		return fmt.Errorf("at least one input file is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	if err := os.MkdirAll(filepath.Dir(outputZipPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	outFile, err := os.Create(outputZipPath)
	if err != nil {
		return fmt.Errorf("create output zip: %w", err)
	}

	zw := zip.NewWriter(outFile)

	for _, inputPath := range inputFiles {
		if err := e.addEncryptedFile(zw, inputPath, password); err != nil {
			_ = outFile.Close()
			return err
		}
	}

	if err := zw.Close(); err != nil {
		_ = outFile.Close()
		return fmt.Errorf("close zip writer: %w", err)
	}

	if err := outFile.Close(); err != nil {
		return fmt.Errorf("close output zip file: %w", err)
	}

	return nil
}

func (e *ZIPEncryptor) addEncryptedFile(zw *zip.Writer, inputPath, password string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input file %s: %w", inputPath, err)
	}
	defer func() { _ = file.Close() }()

	archiveName := filepath.Base(inputPath)
	entryWriter, err := zw.Encrypt(archiveName, password, zip.AES256Encryption)
	if err != nil {
		return fmt.Errorf("create encrypted entry for %s: %w", inputPath, err)
	}

	if _, err := io.Copy(entryWriter, file); err != nil {
		return fmt.Errorf("write encrypted file content for %s: %w", inputPath, err)
	}

	return nil
}
