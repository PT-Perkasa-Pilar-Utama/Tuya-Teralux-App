package dtos

// MailSendRequestDTO represents the request body for sending an email
type MailSendRequestDTO struct {
	To       []string `json:"to" binding:"required" swaggertype:"array,string" example:"user@example.com"`
	Subject  string   `json:"subject" binding:"required" example:"Notification"`
	Template string   `json:"template" binding:"omitempty" example:"test"`
}

// SendMailByMacRequestDTO represents the request body for sending an email by MAC address
type SendMailByMacRequestDTO struct {
	Subject  string `json:"subject" binding:"required" example:"Booking Confirmation"`
	Template string `json:"template" binding:"omitempty" example:"test"`
}
