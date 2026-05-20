package services

import (
	"os"
	"strings"
	"testing"

	"sensio/domain/common/utils"
)

func TestBuildMultipartMessage_IncludesInlineLogo(t *testing.T) {
	os.Setenv("ASSETS_DIR", "/home/farismnrr/Documents/shared/teralux_app/backend/assets")
	defer os.Unsetenv("ASSETS_DIR")

	cfg := &utils.Config{
		SMTPFrom: "test@sensio.app",
	}
	svc := NewMailService(cfg)
	svc.TemplateDir = "/home/farismnrr/Documents/shared/teralux_app/backend/assets/templates/mail"

	body := `<html><body><img src="cid:logo" alt="Logo"></body></html>`

	msg, err := svc.buildMultipartMessage([]string{"recipient@example.com"}, "Test Subject", body, "")
	if err != nil {
		t.Fatalf("buildMultipartMessage failed: %v", err)
	}

	msgStr := string(msg)

	if !strings.Contains(msgStr, "multipart/mixed") {
		t.Error("MIME message missing multipart/mixed Content-Type")
	}

	if !strings.Contains(msgStr, "multipart/related") {
		t.Error("MIME message missing multipart/related for inline images")
	}

	if !strings.Contains(msgStr, "text/html") {
		t.Error("MIME message missing text/html part")
	}

	if !strings.Contains(msgStr, "cid:logo") {
		t.Error("HTML body missing cid:logo reference")
	}

	if !strings.Contains(strings.ToLower(msgStr), "content-id: <logo>") {
		t.Error("MIME message missing Content-ID: <logo> for inline image")
	}

	if !strings.Contains(msgStr, "image/png") {
		t.Error("MIME message missing image/png Content-Type for logo")
	}

	if !strings.Contains(msgStr, "Content-Transfer-Encoding: base64") {
		t.Error("MIME message missing base64 encoding for inline image")
	}
}

func TestBuildMultipartMessage_AttachmentAfterInlineImages(t *testing.T) {
	os.Setenv("ASSETS_DIR", "/home/farismnrr/Documents/shared/teralux_app/backend/assets")
	defer os.Unsetenv("ASSETS_DIR")

	cfg := &utils.Config{
		SMTPFrom: "test@sensio.app",
	}
	svc := NewMailService(cfg)
	svc.TemplateDir = "/home/farismnrr/Documents/shared/teralux_app/backend/assets/templates/mail"

	body := `<html><body><img src="cid:logo"></body></html>`

	msg, err := svc.buildMultipartMessage([]string{"a@b.com"}, "Subject", body, "/nonexistent/file.pdf")
	if err == nil {
		t.Error("Expected error for nonexistent attachment")
	}
	if msg != nil {
		t.Error("Expected nil message when attachment fails")
	}
}
