package services

import (
	"encoding/json"
	"fmt"
	"sensio/domain/common/entities"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/repositories"
	"sensio/domain/common/utils"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	terminal_entities "sensio/domain/terminal/terminal/entities"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"
	"strings"
	"time"

	"github.com/google/uuid"
)

type NotificationExternalService struct {
	terminalRepo  terminal_repositories.ITerminalRepository
	scheduledRepo repositories.IScheduledNotificationRepository
	deviceInfoSvc *DeviceInfoExternalService
	mqttSvc       infrastructure.IMqttService
}

func NewNotificationExternalService(
	terminalRepo terminal_repositories.ITerminalRepository,
	mqttSvc infrastructure.IMqttService,
) *NotificationExternalService {
	return &NotificationExternalService{
		terminalRepo: terminalRepo,
		mqttSvc:      mqttSvc,
	}
}

func NewNotificationExternalServiceWithWA(
	terminalRepo terminal_repositories.ITerminalRepository,
	scheduledRepo repositories.IScheduledNotificationRepository,
	deviceInfoSvc *DeviceInfoExternalService,
	mqttSvc infrastructure.IMqttService,
) *NotificationExternalService {
	return &NotificationExternalService{
		terminalRepo:  terminalRepo,
		scheduledRepo: scheduledRepo,
		deviceInfoSvc: deviceInfoSvc,
		mqttSvc:       mqttSvc,
	}
}

func (s *NotificationExternalService) PublishNotificationToRoom(req terminal_dtos.NotificationPublishRequest) (*terminal_dtos.NotificationPublishResponse, error) {
	loc, err := time.LoadLocation(req.Timezone)
	if err != nil {
		utils.LogError("NotificationExternalService: Failed to load timezone %s: %v", req.Timezone, err)
		return nil, utils.NewAPIError(400, "Invalid timezone. Must be a valid IANA timezone.")
	}

	dateOnly, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		utils.LogError("NotificationExternalService: Failed to parse date: %v", err)
		return nil, utils.NewAPIError(400, "Invalid date format. Must be YYYY-MM-DD.")
	}

	timeOnly, err := time.Parse("15:04:05", req.Time)
	if err != nil {
		utils.LogError("NotificationExternalService: Failed to parse time: %v", err)
		return nil, utils.NewAPIError(400, "Invalid time format. Must be HH:MM:SS.")
	}

	dateTimeEnd := time.Date(dateOnly.Year(), dateOnly.Month(), dateOnly.Day(), timeOnly.Hour(), timeOnly.Minute(), timeOnly.Second(), 0, loc)

	publishAtStr := dateTimeEnd.Format(time.RFC3339)

	terminals, err := s.terminalRepo.GetByRoomID(req.RoomID)
	if err != nil {
		utils.LogError("NotificationExternalService: Failed to query terminals for RoomID %s: %v", req.RoomID, err)
		return nil, fmt.Errorf("failed to lookup terminals: %w", err)
	}

	if len(terminals) == 0 {
		utils.LogDebug("NotificationExternalService: No terminals found for RoomID %s", req.RoomID)
		return nil, utils.NewAPIError(404, fmt.Sprintf("No terminals found for RoomID %s", req.RoomID))
	}

	payload := terminal_dtos.NotificationMQTTPayload{
		PublishAt:        publishAtStr,
		RemainingMinutes: 0,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MQTT payload: %w", err)
	}

	publishedTopics := make([]string, 0, len(terminals))
	for _, t := range terminals {
		topic := fmt.Sprintf("users/%s/%s/notification", t.MacAddress, utils.GetConfig().ApplicationEnvironment)

		utils.LogDebug("NotificationExternalService: MQTT connected=%v, publishing to topic: %s", s.mqttSvc.IsConnected(), topic)

		err := s.mqttSvc.Publish(topic, 1, false, payloadBytes)
		if err != nil {
			utils.LogError("NotificationExternalService: Failed to publish to %s: %v", topic, err)
			return nil, fmt.Errorf("failed to publish to topic %s: %w", topic, err)
		}

		publishedTopics = append(publishedTopics, topic)
	}

	var wanNotificationID string
	if len(req.PhoneNumbers) > 0 && s.scheduledRepo != nil && s.deviceInfoSvc != nil {
		wanID, err := s.scheduleWANotification(req, dateTimeEnd, terminals)
		if err != nil {
			utils.LogWarn("NotificationExternalService: Failed to schedule WA notification: %v", err)
		} else {
			wanNotificationID = wanID
			utils.LogInfo("NotificationExternalService: WA notification scheduled with ID %s", wanID)
		}
	}

	return &terminal_dtos.NotificationPublishResponse{
		RoomID:           req.RoomID,
		PublishAt:        dateTimeEnd,
		PublishedCount:   len(publishedTopics),
		PublishedTopics:  publishedTopics,
		WANotificationID: wanNotificationID,
	}, nil
}

func (s *NotificationExternalService) scheduleWANotification(req terminal_dtos.NotificationPublishRequest, bookingTimeEnd time.Time, terminals []terminal_entities.Terminal) (string, error) {
	if len(terminals) == 0 {
		return "", fmt.Errorf("no terminals found")
	}

	terminal := terminals[0]
	macAddress := terminal.MacAddress

	utils.LogDebug("scheduleWANotification: Terminal MAC=%s, RoomID=%s", macAddress, req.RoomID)

	bookingInfo, err := s.deviceInfoSvc.GetDeviceInfoByMac(macAddress)
	if err != nil {
		utils.LogWarn("NotificationExternalService: Failed to fetch booking info for MAC %s: %v", macAddress, err)
		bookingInfo = make(map[string]interface{})
	} else {
		utils.LogDebug("scheduleWANotification: BookingInfo=%+v", bookingInfo)
	}

	bookingInfoJSON, err := json.Marshal(bookingInfo)
	if err != nil {
		return "", fmt.Errorf("failed to marshal booking info: %w", err)
	}

	phoneNumbersJSON, err := json.Marshal(req.PhoneNumbers)
	if err != nil {
		return "", fmt.Errorf("failed to marshal phone numbers: %w", err)
	}

	scheduledAt := bookingTimeEnd

	notification := &entities.ScheduledNotification{
		ID:             uuid.New().String(),
		RoomID:         req.RoomID,
		MacAddress:     macAddress,
		PhoneNumbers:   string(phoneNumbersJSON),
		BookingInfo:    string(bookingInfoJSON),
		BookingTimeEnd: bookingTimeEnd.Format(time.RFC3339),
		ScheduledAt:    scheduledAt,
		Status:         entities.NotificationStatusPending,
		Template:       req.Template,
	}

	if err := s.scheduledRepo.Create(notification); err != nil {
		return "", fmt.Errorf("failed to save scheduled notification: %w", err)
	}

	return notification.ID, nil
}

func normalizeMacAddress(mac string) string {
	mac = strings.ReplaceAll(mac, ":", "-")
	mac = strings.ReplaceAll(mac, ":", "")
	return strings.ToLower(mac)
}
