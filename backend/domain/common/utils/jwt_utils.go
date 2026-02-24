package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken generates a JWT token for a given Tuya UID with no expiration.
// It uses the JWT_SECRET from the application configuration.
func GenerateToken(uid string) (string, error) {
	config := GetConfig()
	if config.JWTSecret == "" {
		return "", errors.New("JWT_SECRET is not set in configuration")
	}

	// Create claims without 'exp' (expiration) to make it "no-expired"
	claims := jwt.MapClaims{
		"uid": uid,
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken parses and validates a JWT token string.
// It returns the UID extracted from the token if successful.
func ValidateToken(tokenString string) (string, error) {
	config := GetConfig()
	if config.JWTSecret == "" {
		return "", errors.New("JWT_SECRET is not set in configuration")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid, ok := claims["uid"].(string)
		if !ok {
			return "", errors.New("token does not contain a valid UID")
		}
		return uid, nil
	}

	return "", errors.New("invalid token")
}
