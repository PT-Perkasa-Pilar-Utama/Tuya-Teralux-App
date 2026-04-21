package crypto

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PDFProtector struct{}

func NewPDFProtector() *PDFProtector {
	return &PDFProtector{}
}

// Protect applies AES-256 password protection to a PDF.
// The same password is used as both user and owner password.
func (p *PDFProtector) Protect(inputPDFPath, outputPDFPath, password string) error {
	if inputPDFPath == "" {
		return fmt.Errorf("input PDF path is required")
	}
	if outputPDFPath == "" {
		return fmt.Errorf("output PDF path is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	if err := os.MkdirAll(filepath.Dir(outputPDFPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	conf := model.NewAESConfiguration(password, password, 256)
	if err := api.EncryptFile(inputPDFPath, outputPDFPath, conf); err != nil {
		return fmt.Errorf("protect PDF with password: %w", err)
	}

	return nil
}
