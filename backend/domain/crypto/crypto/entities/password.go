package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	minimumPasswordLength = 32
)

// GenerateStrongPassword generates a high-entropy password using the minimum required length.
func GenerateStrongPassword() (string, error) {
	return GeneratePassword(minimumPasswordLength)
}

var (
	lowerCharset   = []byte("abcdefghijklmnopqrstuvwxyz")
	upperCharset   = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	digitCharset   = []byte("0123456789")
	specialCharset = []byte("!@#$%^&*()-_=+[]{}<>?,.:;")
	fullCharset    = append(append(append(lowerCharset, upperCharset...), digitCharset...), specialCharset...)
)

// GeneratePassword generates a high-entropy random password with minimum 32 chars.
// It guarantees at least one lowercase, uppercase, digit, and special character.
func GeneratePassword(length int) (string, error) {
	if length < minimumPasswordLength {
		return "", fmt.Errorf("password length must be at least %d", minimumPasswordLength)
	}

	password := make([]byte, 0, length)

	mandatoryCharsets := [][]byte{lowerCharset, upperCharset, digitCharset, specialCharset}
	for _, charset := range mandatoryCharsets {
		ch, err := randomChar(charset)
		if err != nil {
			return "", fmt.Errorf("generate mandatory character: %w", err)
		}
		password = append(password, ch)
	}

	remaining := length - len(password)
	for i := 0; i < remaining; i++ {
		ch, err := randomChar(fullCharset)
		if err != nil {
			return "", fmt.Errorf("generate random character: %w", err)
		}
		password = append(password, ch)
	}

	if err := secureShuffle(password); err != nil {
		return "", fmt.Errorf("shuffle password: %w", err)
	}

	return string(password), nil
}

func randomChar(charset []byte) (byte, error) {
	if len(charset) == 0 {
		return 0, fmt.Errorf("empty charset")
	}

	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}

	return charset[idx.Int64()], nil
}

func secureShuffle(buf []byte) error {
	for i := len(buf) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		x := int(j.Int64())
		buf[i], buf[x] = buf[x], buf[i]
	}

	return nil
}
