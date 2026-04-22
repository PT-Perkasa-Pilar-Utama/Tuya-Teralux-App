package utils

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DefaultTokenExpiration is the default JWT expiration duration (7 days)
const DefaultTokenExpiration = 7 * 24 * time.Hour

// GenerateToken generates a JWT token for a given Tuya UID with expiration.
// The expiration duration is read from JWT_EXPIRATION_HOURS environment variable.
// If not set, defaults to 7 days.
func GenerateToken(uid string) (string, error) {
	config := GetConfig()
	if config.JWTSecret == "" {
		return "", errors.New("JWT_SECRET is not set in configuration")
	}

	// Read expiration duration from environment variable (in hours)
	expirationHours := DefaultTokenExpiration
	if expStr := os.Getenv("JWT_EXPIRATION_HOURS"); expStr != "" {
		if hours, err := strconv.Atoi(expStr); err == nil && hours > 0 {
			expirationHours = time.Duration(hours) * time.Hour
		} else {
			// Invalid value, use default
			expirationHours = DefaultTokenExpiration
		}
	}

	// Create claims with expiration
	claims := jwt.MapClaims{
		"uid": uid,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(expirationHours).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateLoginToken generates a JWT token for Tuya login with the given access token and expiry.
func GenerateLoginToken(uid string, tuyaAccessToken string, expiry time.Time) (string, error) {
	config := GetConfig()
	if config.JWTSecret == "" {
		return "", errors.New("JWT_SECRET is not set in configuration")
	}

	claims := jwt.MapClaims{
		"uid":          uid,
		"access_token": tuyaAccessToken,
		"iat":          time.Now().Unix(),
		"exp":          expiry.Unix(),
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
// The validation includes checking token expiration.
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
