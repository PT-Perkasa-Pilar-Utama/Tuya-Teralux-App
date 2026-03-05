package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"sensio/domain/common/utils"
)

type BIGRoomAudioUpdateService interface {
	UpdateRoomOccupiedAudio(macAddress, audioPath string) error
}

type bigRoomAudioUpdateService struct {
	client *http.Client
}

func NewBIGRoomAudioUpdateService() BIGRoomAudioUpdateService {
	return &bigRoomAudioUpdateService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type updateRoomOccupiedAudioRequest struct {
	MacAddress string `json:"MacAddress"`
	AudioPath  string `json:"audioPath"`
}

type updateRoomOccupiedAudioResponse struct {
	ResponseJSON string `json:"responsejson"`
}

func (s *bigRoomAudioUpdateService) UpdateRoomOccupiedAudio(macAddress, audioPath string) error {
	url := "https://aplikasi-big.com/IOTAN5JavaDasboard/rest/ProcUpdateRoomOccupiedAudio"

	reqBody := updateRoomOccupiedAudioRequest{
		MacAddress: macAddress,
		AudioPath:  audioPath,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	utils.LogInfo("BIGRoomAudioUpdateService: Sending request to %s with MacAddress: %s, AudioPath: %s", url, macAddress, audioPath)

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("external api returned status: %s", resp.Status)
	}

	var result updateRoomOccupiedAudioResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if result.ResponseJSON != "success" {
		return fmt.Errorf("external api returned unexpected response: %s", result.ResponseJSON)
	}

	utils.LogInfo("BIGRoomAudioUpdateService: Successfully updated external API for MacAddress: %s", macAddress)
	return nil
}
