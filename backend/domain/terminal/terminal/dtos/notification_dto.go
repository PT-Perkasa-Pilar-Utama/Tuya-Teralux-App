package dtos

import "time"

// NotificationPublishRequest represents the request to publish a notification to a room
// Requires date, time, timezone, phone_numbers array, and template (start_meeting or end_meeting).
// If PhoneNumbers is provided, WhatsApp notifications will be scheduled to be sent at publish_at.
type NotificationPublishRequest struct {
	RoomID       string   `json:"room_id" validate:"required" example:"123"`
	Date         string   `json:"date" example:"2026-03-17"`
	Time         string   `json:"time" example:"14:00:00"`
	Timezone     string   `json:"timezone" example:"Asia/Jakarta"`
	PhoneNumbers []string `json:"phone_number" example:"[\"+6281234567890\", \"+6289876543210\"]`
	Template     string   `json:"template" example:"start_meeting"`
}

// NotificationPublishResponse represents the response after publishing notifications
type NotificationPublishResponse struct {
	RoomID           string    `json:"room_id" example:"123"`
	PublishAt        time.Time `json:"publish_at" example:"2026-03-17T13:45:00+07:00"`
	PublishedCount   int       `json:"published_count" example:"2"`
	PublishedTopics  []string  `json:"published_topics" example:"[\"users/AA:BB:CC:DD:EE:FF/dev/notification\"]"`
	WANotificationID string    `json:"wa_notification_id,omitempty" example:"uuid-here"`
}

// NotificationMQTTPayload represents the JSON payload sent via MQTT
type NotificationMQTTPayload struct {
	PublishAt        string `json:"publish_at" example:"2026-03-17T13:45:00+07:00"`
	RemainingMinutes int    `json:"remaining_minutes" example:"15"`
}
