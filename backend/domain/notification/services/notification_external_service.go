package notification

import (
	"encoding/json"
	"fmt"
	"sensio/domain/common/services"
	"sensio/domain/common/utils"
	"sensio/domain/infrastructure"
	notificationEntities "sensio/domain/notification/entities"
	notificationRepositories "sensio/domain/notification/repositories"
	terminal_dtos "sensio/domain/terminal/terminal/dtos"
	terminal_entities "sensio/domain/terminal/terminal/entities"
	terminal_repositories "sensio/domain/terminal/terminal/repositories"
	"strings"
	"time"

	"github.com/google/uuid"
)

type NotificationExternalService struct {
	terminalRepo  terminal_repositories.ITerminalRepository
	scheduledRepo notificationRepositories.IScheduledNotificationRepository
	deviceInfoSvc *services.DeviceInfoExternalService
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
	scheduledRepo notificationRepositories.IScheduledNotificationRepository,
	deviceInfoSvc *services.DeviceInfoExternalService,
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
	var dateTimeEnd time.Time

	if req.ScheduledAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			utils.LogError("NotificationExternalService: Failed to parse scheduled_at: %v", err)
			return nil, utils.NewAPIError(400, "Invalid scheduled_at format. Must be a valid ISO 8601 datetime with timezone offset (e.g. 2026-04-20T23:00:00+07:00).")
		}
		dateTimeEnd = parsed
	} else {
		if s.deviceInfoSvc == nil {
			return nil, utils.NewAPIError(400, "Cannot determine publish time: scheduled_at is required when no booking end time is available from device info.")
		}

		terminals, err := s.terminalRepo.GetByRoomID(req.RoomID)
		if err != nil {
			utils.LogError("NotificationExternalService: Failed to query terminals for RoomID %s: %v", req.RoomID, err)
			return nil, fmt.Errorf("failed to lookup terminals: %w", err)
		}
		if len(terminals) == 0 {
			return nil, utils.NewAPIError(404, fmt.Sprintf("No terminals found for RoomID %s", req.RoomID))
		}

		bookingInfo, err := s.deviceInfoSvc.GetDeviceInfoByMac(terminals[0].MacAddress)
		if err != nil {
			utils.LogWarn("NotificationExternalService: Failed to fetch booking info: %v", err)
			return nil, utils.NewAPIError(400, "Cannot determine publish time: scheduled_at is required when no booking end time is available from device info.")
		}

		dateStr := s.getStringValue(bookingInfo, "SDTGetRoomTeraluxtimeendDate")
		timeRange := s.getStringValue(bookingInfo, "SDTGetRoomTeraluxBookingtimeChar")
		if dateStr == "" || timeRange == "" {
			return nil, utils.NewAPIError(400, "Cannot determine publish time: scheduled_at is required when no booking end time is available from device info.")
		}

		loc, _ := time.LoadLocation("Asia/Jakarta")
		derived, err := s.deriveBookingEndFromRange(dateStr, timeRange, loc)
		if err != nil {
			utils.LogError("NotificationExternalService: Failed to derive booking end: %v", err)
			return nil, utils.NewAPIError(400, "Cannot determine publish time: scheduled_at is required when no booking end time is available from device info.")
		}
		dateTimeEnd = derived
	}

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

	template := req.Template
	if template == "" {
		template = "end_meeting"
	}

	eventType := "meeting_start"
	if template == "end_meeting" {
		eventType = "meeting_end"
	}

	publishedTopics := make([]string, 0, len(terminals))
	for _, t := range terminals {
		terminal := t
		macAddress := terminal.MacAddress

		var title, message string
		var meetingID, roomID string

		if s.deviceInfoSvc != nil {
			bookingInfo, err := s.deviceInfoSvc.GetDeviceInfoByMac(macAddress)
			if err != nil {
				utils.LogWarn("NotificationExternalService: Failed to fetch booking info for MAC %s: %v", macAddress, err)
				bookingInfo = make(map[string]interface{})
			}

			customerName := s.getStringValue(bookingInfo, "SDTGetRoomTeraluxCustomerName")
			roomName := s.getStringValue(bookingInfo, "SDTGetRoomTeraluxRoomName")
			meetingID = s.getStringValue(bookingInfo, "SDTGetRoomTeraluxBookingID")
			roomID = req.RoomID

			if customerName == "" {
				customerName = "Bapak/Ibu"
			}
			if roomName == "" {
				roomName = "the room"
			}

			if eventType == "meeting_start" {
				title = "Meeting Reminder"
				message = fmt.Sprintf("%s, your meeting at %s will start soon", customerName, roomName)
			} else {
				title = "Meeting Ending"
				message = fmt.Sprintf("%s, your meeting at %s is ending soon", customerName, roomName)
			}
		} else {
			if eventType == "meeting_start" {
				title = "Meeting Reminder"
				message = "Your meeting will start soon"
			} else {
				title = "Meeting Ending"
				message = "Your meeting is ending soon"
			}
		}

		payload := terminal_dtos.NotificationMQTTPayload{
			ID:        uuid.New().String(),
			PublishAt: publishAtStr,
			Title:     title,
			Message:   message,
			EventType: eventType,
			MeetingID: meetingID,
			RoomID:    roomID,
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal MQTT payload: %w", err)
		}

		topic := fmt.Sprintf("users/%s/%s/notification", t.MacAddress, utils.GetConfig().ApplicationEnvironment)

		utils.LogDebug("NotificationExternalService: MQTT connected=%v, publishing to topic: %s", s.mqttSvc.IsConnected(), topic)

		err = s.mqttSvc.Publish(topic, 1, false, payloadBytes)
		if err != nil {
			utils.LogError("NotificationExternalService: Failed to publish to %s: %v", topic, err)
			return nil, fmt.Errorf("failed to publish to topic %s: %w", topic, err)
		}

		publishedTopics = append(publishedTopics, topic)

		_ = terminal
	}

	var wanNotificationID string
	if len(req.PhoneNumbers) > 0 && s.scheduledRepo != nil && s.deviceInfoSvc != nil {
		wanID, err := s.scheduleWANotification(req, template, dateTimeEnd, terminals)
		if err != nil {
			utils.LogWarn("NotificationExternalService: Failed to schedule WA notification: %v", err)
		} else {
			wanNotificationID = wanID
			utils.LogInfo("NotificationExternalService: WA notification scheduled with ID %s", wanID)
		}
	}

	return &terminal_dtos.NotificationPublishResponse{
		RoomID:           req.RoomID,
		PublishAt:        dateTimeEnd.Format(time.RFC3339),
		PublishedCount:   len(publishedTopics),
		PublishedTopics:  publishedTopics,
		WANotificationID: wanNotificationID,
	}, nil
}

func (s *NotificationExternalService) scheduleWANotification(req terminal_dtos.NotificationPublishRequest, template string, bookingTimeEnd time.Time, terminals []terminal_entities.Terminal) (string, error) {
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

	notification := &notificationEntities.ScheduledNotification{
		ID:             uuid.New().String(),
		RoomID:         req.RoomID,
		MacAddress:     macAddress,
		PhoneNumbers:   string(phoneNumbersJSON),
		BookingInfo:    string(bookingInfoJSON),
		BookingTimeEnd: bookingTimeEnd.Format(time.RFC3339),
		ScheduledAt:    scheduledAt,
		Status:         notificationEntities.NotificationStatusPending,
		Template:       template,
	}

	if err := s.scheduledRepo.Create(notification); err != nil {
		return "", fmt.Errorf("failed to save scheduled notification: %w", err)
	}

	return notification.ID, nil
}

func (s *NotificationExternalService) getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return strings.TrimSpace(str)
		}
	}
	return ""
}

func (s *NotificationExternalService) deriveBookingEndFromRange(dateStr, timeRange string, loc *time.Location) (time.Time, error) {
	dateOnly, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %w", err)
	}

	parts := strings.Split(timeRange, "-")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time range format: %s", timeRange)
	}

	endPart := strings.TrimSpace(parts[1])
	endPart = strings.Trim(endPart, " ")

	timeFormats := []string{"2006-01-02 03:04 PM", "2006-01-02 3:04 PM", "03:04 PM", "3:04 PM", "15:04"}
	var endTime time.Time
	for _, fmt := range timeFormats {
		if strings.Contains(endPart, "AM") || strings.Contains(endPart, "PM") {
			endTime, err = time.Parse(fmt, endPart)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse end time from range: %w", err)
	}

	return time.Date(dateOnly.Year(), dateOnly.Month(), dateOnly.Day(), endTime.Hour(), endTime.Minute(), 0, 0, loc), nil
}
