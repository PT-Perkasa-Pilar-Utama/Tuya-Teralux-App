package dtos

import "time"

// NotificationPublishRequest represents the request to publish a notification to a room
// At least one of DateTimeEnd or TimeEnd must be provided.
// If both are provided, DateTimeEnd takes priority.
// If PhoneNumbers is provided, WhatsApp notifications will be scheduled to be sent at publish_at.
type NotificationPublishRequest struct {
	RoomID       string   `json:"room_id" validate:"required" example:"123"`
	DateTimeEnd  string   `json:"datetime_end" example:"2026-03-17T14:00:00+07:00"`
	TimeEnd      string   `json:"time_end" example:"14:00:00"`
	IntervalTime int      `json:"interval_time" validate:"min=0" example:"15"`
	PhoneNumbers []string `json:"phone_numbers" example:"[\"6281234567890\", \"6289876543210\"]"`
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
