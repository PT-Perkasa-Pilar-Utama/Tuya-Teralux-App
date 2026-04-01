// +build ignore

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	TuyaClientID     string
	TuyaClientSecret string
	TuyaBaseURL      string
	TuyaUserID       string
}

type TuyaAuthResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Result  struct {
		AccessToken string `json:"access_token"`
		ExpireTime  int    `json:"expire_time"`
		UID         string `json:"uid"`
	} `json:"result"`
}

type TuyaDevicesResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Result  []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Category string `json:"category"`
		Online   bool   `json:"online"`
		GatewayID string `json:"gateway_id"`
	} `json:"result"`
}

type TuyaBatchStatusResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Result  []struct {
		ID       string `json:"id"`
		IsOnline bool   `json:"is_online"`
	} `json:"result"`
}

func loadConfig() *Config {
	return &Config{
		TuyaClientID:     os.Getenv("TUYA_CLIENT_ID"),
		TuyaClientSecret: os.Getenv("TUYA_ACCESS_SECRET"),
		TuyaBaseURL:      os.Getenv("TUYA_BASE_URL"),
		TuyaUserID:       os.Getenv("TUYA_USER_ID"),
	}
}

func generateSignature(clientID, secret, accessToken, timestamp, stringToSign string) string {
	// Simplified HMAC-SHA256 signature
	h := sha256.New()
	h.Write([]byte(secret + accessToken + timestamp + stringToSign))
	return hex.EncodeToString(h.Sum(nil))
}

func getToken(config *Config) (string, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	urlPath := "/v1.0/token?grant_type=1"
	fullURL := config.TuyaBaseURL + urlPath

	stringToSign := fmt.Sprintf("GET\n%s\n\n%s", "", urlPath)
	signature := generateSignature(config.TuyaClientID, config.TuyaClientSecret, "", timestamp, stringToSign)

	req, _ := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("client_id", config.TuyaClientID)
	req.Header.Set("sign", signature)
	req.Header.Set("t", timestamp)
	req.Header.Set("sign_method", "HMAC-SHA256")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var authResp TuyaAuthResponse
	json.Unmarshal(body, &authResp)

	if !authResp.Success {
		return "", fmt.Errorf("auth failed: %s (code: %d)", authResp.Msg, authResp.Code)
	}

	fmt.Printf("✅ Auth Success | UID: %s | Token expires in: %d seconds\n", authResp.Result.UID, authResp.Result.ExpireTime)
	return authResp.Result.AccessToken, nil
}

func getDevices(config *Config, accessToken string) error {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	urlPath := fmt.Sprintf("/v1.0/users/%s/devices", config.TuyaUserID)
	fullURL := config.TuyaBaseURL + urlPath

	stringToSign := fmt.Sprintf("GET\n%s\n\n%s", "", urlPath)
	signature := generateSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, stringToSign)

	req, _ := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("client_id", config.TuyaClientID)
	req.Header.Set("sign", signature)
	req.Header.Set("t", timestamp)
	req.Header.Set("sign_method", "HMAC-SHA256")
	req.Header.Set("access_token", accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var devicesResp TuyaDevicesResponse
	json.Unmarshal(body, &devicesResp)

	fmt.Printf("\n📱 DEVICES LIST (from /v1.0/users/{uid}/devices):\n")
	fmt.Printf("   Success: %v | Code: %d | Msg: %s\n", devicesResp.Success, devicesResp.Code, devicesResp.Msg)
	fmt.Printf("   Total devices: %d\n\n", len(devicesResp.Result))

	for i, dev := range devicesResp.Result {
		status := "🟢 ONLINE"
		if !dev.Online {
			status = "🔴 OFFLINE"
		}
		fmt.Printf("   [%d] %s | ID: %s | Category: %s | Gateway: %s\n", 
			i+1, status, dev.ID, dev.Category, dev.GatewayID)
	}

	// Fetch batch status
	deviceIDs := make([]string, 0, len(devicesResp.Result))
	for _, dev := range devicesResp.Result {
		deviceIDs = append(deviceIDs, dev.ID)
	}

	if len(deviceIDs) > 0 {
		statusURLPath := "/v1.0/iot-03/devices/status"
		statusFullURL := config.TuyaBaseURL + statusURLPath + "?device_ids=" + joinStrings(deviceIDs, ",")

		statusTimestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		statusStringToSign := fmt.Sprintf("GET\n%s\n\n%s", "", statusURLPath)
		statusSignature := generateSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, statusTimestamp, statusStringToSign)

		statusReq, _ := http.NewRequest("GET", statusFullURL, nil)
		statusReq.Header.Set("client_id", config.TuyaClientID)
		statusReq.Header.Set("sign", statusSignature)
		statusReq.Header.Set("t", statusTimestamp)
		statusReq.Header.Set("sign_method", "HMAC-SHA256")
		statusReq.Header.Set("access_token", accessToken)

		statusResp, err := client.Do(statusReq)
		if err != nil {
			fmt.Printf("\n⚠️  Batch status fetch failed: %v\n", err)
			return nil
		}
		defer statusResp.Body.Close()

		statusBody, _ := io.ReadAll(statusResp.Body)
		var batchStatusResp TuyaBatchStatusResponse
		json.Unmarshal(statusBody, &batchStatusResp)

		fmt.Printf("\n📊 BATCH STATUS (from /v1.0/iot-03/devices/status):\n")
		fmt.Printf("   Success: %v | Code: %d | Msg: %s\n", batchStatusResp.Success, batchStatusResp.Code, batchStatusResp.Msg)
		fmt.Printf("   Devices in response: %d\n\n", len(batchStatusResp.Result))

		for i, s := range batchStatusResp.Result {
			status := "🟢 ONLINE"
			if !s.IsOnline {
				status = "🔴 OFFLINE"
			}
			fmt.Printf("   [%d] %s | ID: %s\n", i+1, status, s.ID)
		}
	}

	return nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

func main() {
	fmt.Println("🔍 Tuya Device Status Debugger")
	fmt.Println("==============================\n")

	config := loadConfig()
	
	if config.TuyaClientID == "" {
		fmt.Println("❌ Error: TUYA_CLIENT_ID not set")
		os.Exit(1)
	}

	fmt.Printf("📋 Configuration:\n")
	fmt.Printf("   Client ID: %s\n", config.TuyaClientID)
	fmt.Printf("   Base URL: %s\n", config.TuyaBaseURL)
	fmt.Printf("   User ID: %s\n\n", config.TuyaUserID)

	accessToken, err := getToken(config)
	if err != nil {
		fmt.Printf("❌ Token error: %v\n", err)
		os.Exit(1)
	}

	err = getDevices(config, accessToken)
	if err != nil {
		fmt.Printf("❌ Devices error: %v\n", err)
		os.Exit(1)
	}
}
