package tuya

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Client handles HTTP communication with Tuya API
type Client struct {
	baseURL            string
	clientID           string
	accessSecret       string
	httpClient         *http.Client
	signatureGenerator *SignatureGenerator
	tokenManager       *TokenCacheManager
}

// NewClient creates a new Tuya API client
func NewClient(baseURL, clientID, accessSecret string) *Client {
	return &Client{
		baseURL:            baseURL,
		clientID:           clientID,
		accessSecret:       accessSecret,
		httpClient:         &http.Client{Timeout: 30 * time.Second},
		signatureGenerator: NewSignatureGenerator(clientID, accessSecret),
		tokenManager:       NewTokenCacheManager(),
	}
}

// GetAccessToken returns a valid access token (cached or fresh)
func (c *Client) GetAccessToken() (string, error) {
	// Check cache first
	if token, ok := c.tokenManager.Get(); ok {
		return token, nil
	}

	// Fetch new token
	token, expireTime, err := c.fetchNewToken()
	if err != nil {
		return "", err
	}

	// Cache the token
	c.tokenManager.Set(token, expireTime)

	return token, nil
}

// fetchNewToken requests a new access token from Tuya API
func (c *Client) fetchNewToken() (string, int64, error) {
	timestamp := c.getTimestamp()
	urlPath := "/v1.0/token?grant_type=1"

	// For GET requests, content hash is SHA256 of empty string
	contentHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	stringToSign := c.signatureGenerator.GenerateStringToSign("GET", contentHash, "", urlPath)
	signature := c.signatureGenerator.Generate("", timestamp, stringToSign)

	url := c.baseURL + urlPath
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeaders(req, timestamp, signature)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %w", err)
	}

	var respData map[string]interface{}
	if err := json.Unmarshal(body, &respData); err != nil {
		return "", 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if success, ok := respData["success"].(bool); !ok || !success {
		code := int(respData["code"].(float64))
		msg := respData["msg"].(string)
		return "", 0, fmt.Errorf("%s (code: %d)", msg, code)
	}

	result := respData["result"].(map[string]interface{})
	accessToken := result["access_token"].(string)
	expireTime := int64(result["expire_time"].(float64))

	return accessToken, expireTime, nil
}

// ExecuteRequest executes an authenticated HTTP request to Tuya API
func (c *Client) ExecuteRequest(method, urlPath string, body interface{}) ([]byte, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	timestamp := c.getTimestamp()

	// Prepare request body
	var jsonBody []byte
	var contentHash string

	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		contentHash = c.signatureGenerator.GenerateContentHash(jsonBody)
	} else {
		// For GET requests, content hash is SHA256 of empty string
		contentHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	}

	stringToSign := c.signatureGenerator.GenerateStringToSign(method, contentHash, "", urlPath)
	signature := c.signatureGenerator.Generate(token, timestamp, stringToSign)

	url := c.baseURL + urlPath
	var req *http.Request

	if jsonBody != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeaders(req, timestamp, signature)
	req.Header.Set("access_token", token)

	if jsonBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return respBody, nil
}

// setAuthHeaders sets the required authentication headers
func (c *Client) setAuthHeaders(req *http.Request, timestamp, signature string) {
	req.Header.Set("client_id", c.clientID)
	req.Header.Set("sign", signature)
	req.Header.Set("t", timestamp)
	req.Header.Set("sign_method", "HMAC-SHA256")
}

// getTimestamp returns current timestamp in milliseconds
func (c *Client) getTimestamp() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

// API Response helpers

// APIResponse represents a standard Tuya API response
type APIResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Result  interface{} `json:"result,omitempty"`
	T       int64       `json:"t"`
	Tid     string      `json:"tid,omitempty"`
}

// ParseResponse parses raw response bytes into APIResponse
func ParseResponse(data []byte) (*APIResponse, error) {
	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}

// CheckError checks if API response contains an error
func (r *APIResponse) CheckError() error {
	if !r.Success {
		return fmt.Errorf("%s (code: %d)", r.Msg, r.Code)
	}
	return nil
}
