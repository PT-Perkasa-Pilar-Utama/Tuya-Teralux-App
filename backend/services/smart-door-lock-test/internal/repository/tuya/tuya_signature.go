package tuya

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// SignatureGenerator handles Tuya API signature generation
type SignatureGenerator struct {
	clientID     string
	accessSecret string
}

// NewSignatureGenerator creates a new signature generator
func NewSignatureGenerator(clientID, accessSecret string) *SignatureGenerator {
	return &SignatureGenerator{
		clientID:     clientID,
		accessSecret: accessSecret,
	}
}

// Generate creates a HMAC-SHA256 signature for Tuya API requests
// Message format: clientID + accessToken + timestamp + stringToSign
func (g *SignatureGenerator) Generate(accessToken, timestamp, stringToSign string) string {
	message := g.clientID + accessToken + timestamp + stringToSign

	h := hmac.New(sha256.New, []byte(g.accessSecret))
	h.Write([]byte(message))

	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

// GenerateStringToSign creates the canonical string for signature calculation
// Format: HTTPMethod + "\n" + ContentHash + "\n" + Headers + "\n" + URL
func (g *SignatureGenerator) GenerateStringToSign(method, contentHash, headers, urlPath string) string {
	return method + "\n" + contentHash + "\n" + headers + "\n" + urlPath
}

// GenerateContentHash generates SHA256 hash of content (empty string for GET requests)
func (g *SignatureGenerator) GenerateContentHash(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

// TokenCache holds cached access token with expiry
type TokenCache struct {
	token     string
	expiresAt time.Time
}

// TokenCacheManager manages token caching
type TokenCacheManager struct {
	cache *TokenCache
}

// NewTokenCacheManager creates a new token cache manager
func NewTokenCacheManager() *TokenCacheManager {
	return &TokenCacheManager{}
}

// Get returns cached token if valid
func (m *TokenCacheManager) Get() (string, bool) {
	if m.cache == nil {
		return "", false
	}

	if time.Now().Before(m.cache.expiresAt) {
		return m.cache.token, true
	}

	return "", false
}

// Set caches a token with expiry (subtracts buffer for safety)
func (m *TokenCacheManager) Set(token string, expireSeconds int64) {
	buffer := int64(60) // 60 second buffer
	if expireSeconds > buffer {
		expireSeconds -= buffer
	}

	m.cache = &TokenCache{
		token:     token,
		expiresAt: time.Now().Add(time.Duration(expireSeconds) * time.Second),
	}
}
