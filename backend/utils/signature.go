package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// GenerateTuyaSignature generates HMAC-SHA256 signature for Tuya API
// Formula: HMAC-SHA256(client_id + access_token + t + stringToSign, client_secret)
// If access_token is empty, it is omitted from the message.
func GenerateTuyaSignature(clientID, clientSecret, accessToken, timestamp, stringToSign string) string {
	// Concatenate: client_id + access_token + t + stringToSign
	message := clientID + accessToken + timestamp + stringToSign

	// Create HMAC-SHA256 hash
	h := hmac.New(sha256.New, []byte(clientSecret))
	h.Write([]byte(message))
	signature := h.Sum(nil)

	// Convert to uppercase hexadecimal
	return strings.ToUpper(hex.EncodeToString(signature))
}

// GenerateTuyaStringToSign creates the string to sign for GET requests
// Format: HTTPMethod + "\n" + ContentHash + "\n" + Headers + "\n" + URL
func GenerateTuyaStringToSign(httpMethod, contentHash, headers, url string) string {
	return httpMethod + "\n" + contentHash + "\n" + headers + "\n" + url
}
