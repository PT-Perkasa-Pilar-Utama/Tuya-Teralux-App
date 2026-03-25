package dtos

import "time"

// NotificationPublishRequest represents the request to publish a notification to a room
type NotificationPublishRequest struct {
	RoomID       string `json:"room_id" validate:"required" example:"123"`
	DateTimeEnd  string `json:"datetime_end" validate:"required" example:"2026-03-17T14:00:00+07:00"`
	IntervalTime int    `json:"interval_time" validate:"min=0" example:"15"`
}

// NotificationPublishResponse represents the response after publishing notifications
type NotificationPublishResponse struct {
	RoomID          string    `json:"room_id" example:"123"`
	PublishAt       time.Time `json:"publish_at" example:"2026-03-17T13:45:00+07:00"`
	PublishedCount  int       `json:"published_count" example:"2"`
	PublishedTopics []string  `json:"published_topics" example:"[\"users/AA:BB:CC:DD:EE:FF/dev/notification\"]"`
}

// NotificationMQTTPayload represents the JSON payload sent via MQTT
type NotificationMQTTPayload struct {
	PublishAt        string `json:"publish_at" example:"2026-03-17T13:45:00+07:00"`
	RemainingMinutes int    `json:"remaining_minutes" example:"15"`
}
