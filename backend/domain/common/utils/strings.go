package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/google/uuid"
)

// JoinStrings concatenates a slice of strings into a single string with a given separator.
// This is a wrapper around the standard strings.Join function.
//
// param elems The slice of strings to join.
// param sep The separator string.
// return string The joined string.
func JoinStrings(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

// HashString generates the SHA256 hash of a given string.
// It returns the hash as a hexadecimal encoded string.
//
// param s The input string to hash.
// return string The SHA256 hash in hex format.
func HashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateUUID generates a random UUID string using google/uuid.
func GenerateUUID() string {
	return uuid.New().String()
}
