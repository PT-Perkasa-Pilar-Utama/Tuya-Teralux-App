package service

import (
	"sensio/backend/services/smart-door-lock-test/internal/domain"
	"sensio/backend/services/smart-door-lock-test/internal/repository/tuya"
)

// CommandService handles command-related business logic
type CommandService struct {
	commandRepo *tuya.CommandRepository
}

// NewCommandService creates a new command service
func NewCommandService(commandRepo *tuya.CommandRepository) *CommandService {
	return &CommandService{commandRepo: commandRepo}
}

// Lock locks the door
func (s *CommandService) Lock(deviceID string) error {
	return s.commandRepo.Lock(deviceID)
}

// Unlock unlocks the door
func (s *CommandService) Unlock(deviceID string) error {
	return s.commandRepo.Unlock(deviceID)
}

// SendCommand sends a custom command to the device
func (s *CommandService) SendCommand(deviceID string, code string, value interface{}) error {
	cmd := domain.Command{
		Code:  code,
		Value: value,
	}

	return s.commandRepo.Send(deviceID, []domain.Command{cmd})
}

// SendCommands sends multiple commands to the device
func (s *CommandService) SendCommands(deviceID string, commands []domain.Command) error {
	return s.commandRepo.Send(deviceID, commands)
}
