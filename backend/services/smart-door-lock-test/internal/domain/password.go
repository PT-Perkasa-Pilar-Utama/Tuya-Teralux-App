package domain

import "time"

// PasswordType represents the type of password
type PasswordType string

const (
	// PasswordTypeDynamic is a one-time password valid for 5 minutes
	PasswordTypeDynamic PasswordType = "dynamic"

	// PasswordTypeTemporary is a reusable password with custom duration
	PasswordTypeTemporary PasswordType = "temporary"

	// PasswordTypeUnlimited is a password with no expiration (if supported by device)
	PasswordTypeUnlimited PasswordType = "unlimited"
)

// SyncStatus represents the synchronization status of a password with the device/cloud
type SyncStatus string

const (
	// SyncStatusActive means the password is created and active on the device
	SyncStatusActive SyncStatus = "active"

	// SyncStatusPending means the password is stored locally, waiting for device to come online
	SyncStatusPending SyncStatus = "pending_sync"

	// SyncStatusFailed means the password creation failed permanently
	SyncStatusFailed SyncStatus = "failed"

	// SyncStatusExpired means the password expired before it could be synced
	SyncStatusExpired SyncStatus = "expired"
)

// Password represents a generated door lock password
type Password struct {
	Value        string
	Type         PasswordType
	ExpireAt     time.Time
	ValidMinutes int
	SyncStatus   SyncStatus
}

// IsValid returns true if the password is still valid
func (p *Password) IsValid() bool {
	return time.Now().Before(p.ExpireAt)
}

// TimeRemaining returns the time remaining until expiration
func (p *Password) TimeRemaining() time.Duration {
	return time.Until(p.ExpireAt)
}

// IsSynced returns true if the password is synced with the device
func (p *Password) IsSynced() bool {
	return p.SyncStatus == SyncStatusActive
}

// PasswordCreationResult represents the result of a password creation attempt
type PasswordCreationResult struct {
	Password *Password
	Status   SyncStatus
	Message  string
}

// IsPending returns true if the password creation is pending sync
func (r *PasswordCreationResult) IsPending() bool {
	return r.Status == SyncStatusPending
}

// IsActive returns true if the password is active and ready to use
func (r *PasswordCreationResult) IsActive() bool {
	return r.Status == SyncStatusActive
}

// PasswordRequest represents a request to generate a password
type PasswordRequest struct {
	Type       PasswordType
	DeviceID   string
	Duration   int    // minutes (for temporary)
	CustomValue string // optional custom password
}

// Validate validates the password request
func (r *PasswordRequest) Validate() error {
	if r.DeviceID == "" {
		return ErrDeviceIDRequired
	}

	if r.Type == PasswordTypeTemporary && r.Duration <= 0 {
		return ErrInvalidDuration
	}

	return nil
}

// Password errors
var (
	ErrDeviceIDRequired = &ValidationError{"device_id is required"}
	ErrInvalidDuration  = &ValidationError{"duration must be positive"}
	ErrPasswordExpired  = &ValidationError{"password has expired"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
