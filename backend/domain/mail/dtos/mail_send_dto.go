package dtos

// MailSendRequestDTO represents the request body for sending an email
type MailSendRequestDTO struct {
	To             []string `json:"to" binding:"required" swaggertype:"array,string" example:"user@example.com"`
	Subject        string   `json:"subject" binding:"required" example:"Notification"`
	Template       string                 `json:"template" binding:"omitempty" example:"test"`
	Data           map[string]interface{} `json:"data,omitempty" example:"{\"customer_name\": \"John Doe\"}"`
	AttachmentPath *string                `json:"attachment_path,omitempty" example:"/uploads/reports/summary_123.pdf"`
}

// SendMailByMacRequestDTO represents the request body for sending an email by MAC address
type SendMailByMacRequestDTO struct {
	Subject        string                 `json:"subject" binding:"required" example:"Booking Confirmation"`
	Template       string                 `json:"template" binding:"omitempty" example:"test"`
	Data           map[string]interface{} `json:"data,omitempty" example:"{\"customer_name\": \"John Doe\"}"`
	AttachmentPath *string                `json:"attachment_path,omitempty" example:"/uploads/reports/summary_123.pdf"`
}
// MailTaskResponseDTO represents the immediate response for a mail task
type MailTaskResponseDTO struct {
	TaskID     string `json:"task_id" example:"mail-abc-123"`
	TaskStatus string `json:"task_status" example:"pending"`
}

// MailStatusDTO represents the detailed status and result of a mail task
type MailStatusDTO struct {
	Status          string      `json:"status" example:"completed"`
	Result          string      `json:"result,omitempty" example:"Email sent to user@example.com"`
	Error           string      `json:"error,omitempty" example:"smtp auth failed"`
	Trigger         string      `json:"trigger,omitempty"`
	HTTPStatusCode  int         `json:"-"`
	StartedAt       string      `json:"started_at,omitempty"`
	DurationSeconds float64     `json:"duration_seconds,omitempty"`
	ExpiresAt       string      `json:"expires_at,omitempty"`
	ExpiresInSecond int64       `json:"expires_in_seconds,omitempty"`
}

// SetExpiry implements tasks.StatusWithExpiry interface
func (s *MailStatusDTO) SetExpiry(expiresAt string, expiresInSeconds int64) {
	s.ExpiresAt = expiresAt
	s.ExpiresInSecond = expiresInSeconds
}

// SwaggerEmailTemplateData represents the expected map structure for the email template (used for Swagger Docs only)
type SwaggerEmailTemplateData struct {
	Email            string `json:"email,omitempty" example:"override@example.com"`
	CustomerName     string `json:"customer_name" example:"John Doe"`
	CustomerCompany  string `json:"customer_company" example:"PT Perkasa Pilar Utama"`
	BookingDate      string `json:"booking_date" example:"24 Februari 2026"`
	BookingTimeStart string `json:"booking_time_start" example:"10:00"`
	BookingTimeStop  string `json:"booking_time_stop" example:"12:00"`
	BookingPlace     string `json:"booking_place" example:"Lt. 3"`
	BookingRoom      string `json:"booking_room" example:"Ruang Cendrawasih"`
}

// SwaggerMailSendRequestDTO is used to generate proper Swagger documentation for the generic map
type SwaggerMailSendRequestDTO struct {
	To             []string                 `json:"to" binding:"required" example:"user@example.com"`
	Subject        string                   `json:"subject" binding:"required" example:"Notification"`
	Template       string                   `json:"template" example:"summary"`
	Data           SwaggerEmailTemplateData `json:"data,omitempty"`
	AttachmentPath *string                  `json:"attachment_path,omitempty" example:"/uploads/reports/summary_123.pdf"`
}

// SwaggerSendMailByMacRequestDTO is used to generate proper Swagger documentation for the generic map
type SwaggerSendMailByMacRequestDTO struct {
	Subject        string                   `json:"subject" binding:"required" example:"Booking Confirmation"`
	Template       string                   `json:"template" example:"summary"`
	Data           SwaggerEmailTemplateData `json:"data,omitempty"`
	AttachmentPath *string                  `json:"attachment_path,omitempty" example:"/uploads/reports/summary_123.pdf"`
}
