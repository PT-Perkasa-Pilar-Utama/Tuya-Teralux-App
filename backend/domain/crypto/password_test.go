package crypto

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/yeka/zip"
)

func TestGeneratePassword(t *testing.T) {
	password, err := GeneratePassword(32)
	if err != nil {
		t.Fatalf("GeneratePassword returned error: %v", err)
	}

	if len(password) < 32 {
		t.Fatalf("password length = %d, want at least 32", len(password))
	}

	if !containsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		t.Fatal("password does not contain lowercase characters")
	}
	if !containsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		t.Fatal("password does not contain uppercase characters")
	}
	if !containsAny(password, "0123456789") {
		t.Fatal("password does not contain numeric characters")
	}
	if !containsAny(password, "!@#$%^&*()-_=+[]{}<>?,.:;") {
		t.Fatal("password does not contain special characters")
	}
}

func TestZIPEncryptionRejectsWrongPassword(t *testing.T) {
	tempDir := t.TempDir()

	inputFilePath := filepath.Join(tempDir, "hello.txt")
	if err := os.WriteFile(inputFilePath, []byte("top-secret-content"), 0o600); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	zipPath := filepath.Join(tempDir, "encrypted.zip")
	password := "StrongP@ssw0rd!With32+Chars__123456"

	encryptor := NewZIPEncryptor()
	if err := encryptor.EncryptFiles(zipPath, []string{inputFilePath}, password); err != nil {
		t.Fatalf("EncryptFiles returned error: %v", err)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open encrypted zip: %v", err)
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		t.Fatal("encrypted zip does not contain files")
	}

	wrongFile := reader.File[0]
	wrongFile.SetPassword("incorrect-password")
	rcWrong, err := wrongFile.Open()
	if err == nil {
		_, readErr := io.ReadAll(rcWrong)
		rcWrong.Close()
		if readErr == nil {
			t.Fatal("expected wrong password to fail when reading encrypted zip")
		}
	}

	correctFile := reader.File[0]
	correctFile.SetPassword(password)
	rc, err := correctFile.Open()
	if err != nil {
		t.Fatalf("open encrypted entry with correct password: %v", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read encrypted entry with correct password: %v", err)
	}

	if string(content) != "top-secret-content" {
		t.Fatalf("decrypted content mismatch: got %q", string(content))
	}
}

func TestPDFProtectionRejectsWrongPassword(t *testing.T) {
	sourcePDF := filepath.Clean(filepath.Join("..", "..", "services", "smart-door-lock-test", "Panduandaring_Smart_Door_Lock_MJ1S.pdf"))
	if _, err := os.Stat(sourcePDF); err != nil {
		t.Fatalf("source PDF fixture missing: %v", err)
	}

	tempDir := t.TempDir()
	protectedPath := filepath.Join(tempDir, "protected.pdf")
	decryptedPath := filepath.Join(tempDir, "decrypted.pdf")

	password := "An0ther$tr0ngP@ssw0rd!WithLen32++"
	protector := NewPDFProtector()
	if err := protector.Protect(sourcePDF, protectedPath, password); err != nil {
		t.Fatalf("Protect returned error: %v", err)
	}

	wrongConf := model.NewAESConfiguration("wrong-password", "wrong-password", 256)
	if err := api.DecryptFile(protectedPath, decryptedPath, wrongConf); err == nil {
		t.Fatal("expected wrong password to fail decrypting protected PDF")
	}

	correctConf := model.NewAESConfiguration(password, password, 256)
	if err := api.DecryptFile(protectedPath, decryptedPath, correctConf); err != nil {
		t.Fatalf("decrypt protected PDF with correct password failed: %v", err)
	}
}

func containsAny(source, charset string) bool {
	for _, ch := range charset {
		if strings.ContainsRune(source, ch) {
			return true
		}
	}

	return false
}
