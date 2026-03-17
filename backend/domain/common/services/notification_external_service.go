package services

import (
	"encoding/json"
	"fmt"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	"sensio/domain/terminal/terminal/repositories"
	"time"
)

// NotificationExternalService handles computing and publishing notifications to terminals
type NotificationExternalService struct {
	terminalRepo repositories.ITerminalRepository
	mqttSvc      infrastructure.IMqttService
}

// NewNotificationExternalService creates a new instance of NotificationExternalService
func NewNotificationExternalService(terminalRepo repositories.ITerminalRepository, mqttSvc infrastructure.IMqttService) *NotificationExternalService {
	return &NotificationExternalService{
		terminalRepo: terminalRepo,
		mqttSvc:      mqttSvc,
	}
}

// PublishNotificationToRoom computes the publish time and sends MQTT messages to all terminals in a room
func (s *NotificationExternalService) PublishNotificationToRoom(req terminal_dtos.NotificationPublishRequest) (*terminal_dtos.NotificationPublishResponse, error) {
	// 1. Parse DateTimeEnd
	dateTimeEnd, err := time.Parse(time.RFC3339, req.DateTimeEnd)
	if err != nil {
		utils.LogError("NotificationExternalService: Failed to parse DateTimeEnd: %v", err)
		return nil, utils.NewAPIError(400, "Invalid datetime_end format. Must be RFC3339.")
	}

	// 2. Compute PublishAt
	publishAt := dateTimeEnd.Add(time.Duration(-req.IntervalTime) * time.Minute)
	publishAtStr := publishAt.Format(time.RFC3339)

	// 3. Lookup terminals in the room
	terminals, err := s.terminalRepo.GetByRoomID(req.RoomID)
	if err != nil {
		utils.LogError("NotificationExternalService: Failed to query terminals for RoomID %s: %v", req.RoomID, err)
		return nil, fmt.Errorf("failed to lookup terminals: %w", err)
	}

	if len(terminals) == 0 {
		utils.LogDebug("NotificationExternalService: No terminals found for RoomID %s", req.RoomID)
		return nil, utils.NewAPIError(404, fmt.Sprintf("No terminals found for RoomID %s", req.RoomID))
	}

	// 4. Prepare MQTT payload
	payload := terminal_dtos.NotificationMQTTPayload{
		PublishAt: publishAtStr,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MQTT payload: %w", err)
	}

	// 5. Fan out to each terminal
	publishedTopics := make([]string, 0, len(terminals))
	for _, t := range terminals {
		topic := fmt.Sprintf("users/%s/notification", t.MacAddress)
		
		err := s.mqttSvc.Publish(topic, 0, false, payloadBytes)
		if err != nil {
			utils.LogError("NotificationExternalService: Failed to publish to %s: %v", topic, err)
			// According to the plan, we treat any failure as request failure
			return nil, fmt.Errorf("failed to publish to topic %s: %w", topic, err)
		}
		
		publishedTopics = append(publishedTopics, topic)
	}

	// 6. Return response
	return &terminal_dtos.NotificationPublishResponse{
		RoomID:          req.RoomID,
		PublishAt:       publishAt,
		PublishedCount:  len(publishedTopics),
		PublishedTopics: publishedTopics,
	}, nil
}
