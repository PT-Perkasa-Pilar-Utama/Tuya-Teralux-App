package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerateTuyaStringToSign(t *testing.T) {
	method := "GET"
	contentHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // empty sha256
	headers := ""
	url := "/v1.0/devices"

	expected := "GET\ne3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855\n\n/v1.0/devices"
	result := GenerateTuyaStringToSign(method, contentHash, headers, url)

	if result != expected {
		t.Errorf("GenerateTuyaStringToSign() = %q; want %q", result, expected)
	}
}

func TestGenerateTuyaSignature(t *testing.T) {
	clientID := "client_id"
	clientSecret := "client_secret"
	accessToken := "access_token"
	timestamp := "1600000000000"
	stringToSign := "STRING_TO_SIGN"

	// Expected calculation
	message := clientID + accessToken + timestamp + stringToSign
	h := hmac.New(sha256.New, []byte(clientSecret))
	h.Write([]byte(message))
	expected := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	result := GenerateTuyaSignature(clientID, clientSecret, accessToken, timestamp, stringToSign)

	if result != expected {
		t.Errorf("GenerateTuyaSignature() = %q; want %q", result, expected)
	}
}
