package worker

import (
	"fmt"
	"log"
	"sensio/backend/services/smart-door-lock-test/internal/domain"
	"sensio/backend/services/smart-door-lock-test/internal/repository/sqlite"
	"sensio/backend/services/smart-door-lock-test/internal/repository/tuya"
	"time"
)

// SyncWorker handles synchronization of pending passwords with devices
type SyncWorker struct {
	passwordRepo        *tuya.PasswordRepository
	pendingPasswordRepo *sqlite.PendingPasswordRepository
	deviceService       DeviceService
	interval            time.Duration
	stopChan            chan struct{}
	logger              *log.Logger
}

// DeviceService defines the interface for device operations
type DeviceService interface {
	IsOnline(deviceID string) (bool, error)
}

// NewSyncWorker creates a new sync worker
func NewSyncWorker(
	passwordRepo *tuya.PasswordRepository,
	pendingPasswordRepo *sqlite.PendingPasswordRepository,
	deviceService DeviceService,
	interval time.Duration,
) *SyncWorker {
	return &SyncWorker{
		passwordRepo:        passwordRepo,
		pendingPasswordRepo: pendingPasswordRepo,
		deviceService:       deviceService,
		interval:            interval,
		stopChan:            make(chan struct{}),
		logger:              log.New(log.Writer(), "[SyncWorker] ", log.LstdFlags),
	}
}

// Start begins the sync worker loop
func (w *SyncWorker) Start() {
	w.logger.Println("Starting sync worker...")

	go w.run()
}

// Stop stops the sync worker
func (w *SyncWorker) Stop() {
	w.logger.Println("Stopping sync worker...")
	close(w.stopChan)
}

func (w *SyncWorker) run() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.syncPendingPasswords()
		case <-w.stopChan:
			return
		}
	}
}

// syncPendingPasswords attempts to sync all pending passwords
func (w *SyncWorker) syncPendingPasswords() {
	w.logger.Println("Checking for pending passwords...")

	// Get all pending passwords
	pending, err := w.pendingPasswordRepo.GetAllPending()
	if err != nil {
		w.logger.Printf("Error getting pending passwords: %v", err)
		return
	}

	if len(pending) == 0 {
		w.logger.Println("No pending passwords to sync")
		return
	}

	w.logger.Printf("Found %d pending password(s) to sync", len(pending))

	// Group by device ID
	devicePasswords := make(map[string][]*sqlite.PendingPassword)
	for _, pp := range pending {
		devicePasswords[pp.DeviceID] = append(devicePasswords[pp.DeviceID], pp)
	}

	// Process each device
	for deviceID, passwords := range devicePasswords {
		w.syncDevicePasswords(deviceID, passwords)
	}
}

// syncDevicePasswords syncs passwords for a specific device
func (w *SyncWorker) syncDevicePasswords(deviceID string, passwords []*sqlite.PendingPassword) {
	// Check if device is online
	online, err := w.deviceService.IsOnline(deviceID)
	if err != nil {
		w.logger.Printf("Error checking device %s status: %v", deviceID, err)
		// Increment retry count for all passwords
		for _, pp := range passwords {
			w.pendingPasswordRepo.IncrementRetryCount(pp.ID)
		}
		return
	}

	if !online {
		w.logger.Printf("Device %s is offline, skipping sync", deviceID)
		return
	}

	w.logger.Printf("Device %s is online, syncing %d password(s)", deviceID, len(passwords))

	// Try to sync each password
	for _, pp := range passwords {
		w.syncSinglePassword(deviceID, pp)
	}
}

// syncSinglePassword attempts to sync a single pending password
func (w *SyncWorker) syncSinglePassword(deviceID string, pp *sqlite.PendingPassword) {
	// Check if password has expired
	if time.Now().After(pp.ExpireAt) {
		w.logger.Printf("Password %d has expired, marking as failed", pp.ID)
		w.pendingPasswordRepo.UpdateStatus(pp.ID, domain.SyncStatusExpired)
		return
	}

	// Check max retries (optional, can be configured)
	maxRetries := 10
	if pp.RetryCount >= maxRetries {
		w.logger.Printf("Password %d exceeded max retries (%d), marking as failed", pp.ID, maxRetries)
		w.pendingPasswordRepo.UpdateStatus(pp.ID, domain.SyncStatusFailed)
		return
	}

	// Attempt to create password on device
	w.logger.Printf("Attempting to sync password %d for device %s", pp.ID, deviceID)

	req := &domain.PasswordRequest{
		Type:        pp.PasswordType,
		DeviceID:    deviceID,
		Duration:    pp.ValidMinutes,
		CustomValue: pp.PasswordValue,
	}

	_, err := w.passwordRepo.Generate(req)
	if err != nil {
		w.logger.Printf("Failed to sync password %d: %v", pp.ID, err)
		w.pendingPasswordRepo.IncrementRetryCount(pp.ID)
		return
	}

	// Success! Update status to active
	w.logger.Printf("Successfully synced password %d", pp.ID)
	w.pendingPasswordRepo.UpdateStatus(pp.ID, domain.SyncStatusActive)
}

// SyncNow triggers an immediate sync (useful for manual triggers or device-online events)
func (w *SyncWorker) SyncNow() {
	w.logger.Println("Manual sync triggered")
	w.syncPendingPasswords()
}

// SyncDevice triggers an immediate sync for a specific device
func (w *SyncWorker) SyncDevice(deviceID string) error {
	w.logger.Printf("Manual sync triggered for device %s", deviceID)

	// Check if device is online
	online, err := w.deviceService.IsOnline(deviceID)
	if err != nil {
		return fmt.Errorf("error checking device status: %w", err)
	}

	if !online {
		return fmt.Errorf("device %s is offline", deviceID)
	}

	// Get pending passwords for this device
	pending, err := w.pendingPasswordRepo.GetPendingByDeviceID(deviceID)
	if err != nil {
		return fmt.Errorf("error getting pending passwords: %w", err)
	}

	if len(pending) == 0 {
		w.logger.Printf("No pending passwords for device %s", deviceID)
		return nil
	}

	w.syncDevicePasswords(deviceID, pending)
	return nil
}

// GetStats returns statistics about pending passwords
func (w *SyncWorker) GetStats() (*SyncStats, error) {
	allPending, err := w.pendingPasswordRepo.GetAllPending()
	if err != nil {
		return nil, err
	}

	expired, err := w.pendingPasswordRepo.GetExpired()
	if err != nil {
		return nil, err
	}

	stats := &SyncStats{
		TotalPending: len(allPending),
		TotalExpired: len(expired),
	}

	// Count by device
	stats.ByDevice = make(map[string]int)
	for _, pp := range allPending {
		stats.ByDevice[pp.DeviceID]++
	}

	return stats, nil
}

// SyncStats contains statistics about pending passwords
type SyncStats struct {
	TotalPending int
	TotalExpired int
	ByDevice     map[string]int
}
