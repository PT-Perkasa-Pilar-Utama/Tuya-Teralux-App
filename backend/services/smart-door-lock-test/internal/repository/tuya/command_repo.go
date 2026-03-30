package tuya

import (
	"sensio/backend/services/smart-door-lock-test/internal/domain"
)

// CommandRepository handles command-related API operations
type CommandRepository struct {
	client *Client
}

// NewCommandRepository creates a new command repository
func NewCommandRepository(client *Client) *CommandRepository {
	return &CommandRepository{client: client}
}

// Send sends one or more commands to a device
func (r *CommandRepository) Send(deviceID string, commands []domain.Command) error {
	urlPath := "/v1.0/devices/" + deviceID + "/commands"

	body := BuildCommandRequest(commands)

	respBody, err := r.client.ExecuteRequest("POST", urlPath, body)
	if err != nil {
		return err
	}

	apiResp, err := ParseResponse(respBody)
	if err != nil {
		return err
	}

	return apiResp.CheckError()
}

// Lock locks the door
func (r *CommandRepository) Lock(deviceID string) error {
	return r.Send(deviceID, []domain.Command{domain.LockCommand(true)})
}

// Unlock unlocks the door
func (r *CommandRepository) Unlock(deviceID string) error {
	return r.Send(deviceID, []domain.Command{domain.LockCommand(false)})
}
