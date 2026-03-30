package domain

// Device represents a Tuya smart door lock device
type Device struct {
	ID         string
	Name       string
	Category   string
	Online     bool
	Statuses   []DeviceStatus
	Functions  []DeviceFunction
	LocalKey   string
	CreateTime int64
	UpdateTime int64
}

// DeviceStatus represents a single device status reading
type DeviceStatus struct {
	Code  string
	Value interface{}
}

// DeviceFunction represents a device function specification
type DeviceFunction struct {
	Code   string
	Type   string
	Values string
}

// IsOnline returns whether the device is currently online
func (d *Device) IsOnline() bool {
	return d.Online
}

// GetStatus returns the value of a specific status code
func (d *Device) GetStatus(code string) (interface{}, bool) {
	for _, status := range d.Statuses {
		if status.Code == code {
			return status.Value, true
		}
	}
	return nil, false
}

// GetLockState returns the current lock state
func (d *Device) GetLockState() LockState {
	value, exists := d.GetStatus("lock_motor_state")
	if !exists {
		return LockStateUnknown
	}

	if value == true {
		return LockStateUnlocked
	}
	return LockStateLocked
}

// LockState represents the physical state of the lock
type LockState int

const (
	LockStateUnknown LockState = iota
	LockStateLocked
	LockStateUnlocked
)

func (s LockState) String() string {
	switch s {
	case LockStateLocked:
		return "locked"
	case LockStateUnlocked:
		return "unlocked"
	default:
		return "unknown"
	}
}
