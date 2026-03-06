package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
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

// HashFile generates the SHA256 hash of a file's content.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return HashReader(f)
}

// HashReader generates the SHA256 hash of an io.Reader's content.
func HashReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// GenerateUUID generates a random UUID string using google/uuid.
func GenerateUUID() string {
	return uuid.New().String()
}
