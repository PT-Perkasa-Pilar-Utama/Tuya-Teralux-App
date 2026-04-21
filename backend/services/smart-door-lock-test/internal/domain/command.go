package domain

// Command represents a device command
type Command struct {
	Code  string
	Value interface{}
}

// CommandRequest represents a request to send commands to a device
type CommandRequest struct {
	DeviceID string
	Commands []Command
}

// Validate validates the command request
func (r *CommandRequest) Validate() error {
	if r.DeviceID == "" {
		return ErrDeviceIDRequired
	}

	if len(r.Commands) == 0 {
		return &ValidationError{"at least one command is required"}
	}

	for i, cmd := range r.Commands {
		if cmd.Code == "" {
			return &ValidationError{"command code is required at index " + string(rune(i))}
		}
	}

	return nil
}

// LockCommand creates a command to lock/unlock the door
func LockCommand(locked bool) Command {
	return Command{
		Code:  "lock_motor_state",
		Value: !locked, // true = unlock, false = lock
	}
}

// Common command codes for smart door locks
const (
	CmdLockMotorState    = "lock_motor_state"
	CmdUnlockFingerprint = "unlock_fingerprint"
	CmdUnlockPassword    = "unlock_password"
	CmdUnlockTemporary   = "unlock_temporary"
	CmdUnlockCard        = "unlock_card"
	CmdUnlockFace        = "unlock_face"
	CmdUnlockApp         = "unlock_app"
	CmdAlarmLock         = "alarm_lock"
	CmdBatteryState      = "battery_state"
	CmdHijack            = "hijack"
	CmdDoorbell          = "doorbell"
)
