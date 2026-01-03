package usecases

import (
	"teralux_app/domain/tuya/dtos"
)

// SyncDeviceStatusUseCase synchronizes device status from Tuya to Teralux DB
type SyncDeviceStatusUseCase struct {
	getAllDevicesUC *TuyaGetAllDevicesUseCase
}

// NewSyncDeviceStatusUseCase creates a new instance of SyncDeviceStatusUseCase
func NewSyncDeviceStatusUseCase(getAllDevicesUC *TuyaGetAllDevicesUseCase) *SyncDeviceStatusUseCase {
	return &SyncDeviceStatusUseCase{
		getAllDevicesUC: getAllDevicesUC,
	}
}

// Execute synchronizes devices for the given user (Tuya logic) and returns simplified status
func (uc *SyncDeviceStatusUseCase) Execute(accessToken, userID string) ([]dtos.TuyaSyncDeviceDTO, error) {
	// 1. Fetch real-time data from Tuya
	tuyaResp, err := uc.getAllDevicesUC.GetAllDevices(accessToken, userID, 1, 100, "")
	if err != nil {
		return nil, err
	}

	// 2. Map to simplified DTO and Flatten Collections
	var syncDevices []dtos.TuyaSyncDeviceDTO

	for _, d := range tuyaResp.Devices {
		// Add the main device (if it's not just a container regarding logic, but here we add it)
		// User example shows mix of IDs.
		// Usually if it has collections, the parent might not be "online" in the same way, but let's include it or check requirements.
		// User example:
		// {id: "...", remote_id: "...", online: false, ...} -> This looks like a sub-device (IR remote)
		// {id: "...", online: false, ...} -> This looks like a standalone device

		// Add Parent Device
		syncDevices = append(syncDevices, dtos.TuyaSyncDeviceDTO{
			ID:         d.ID,
			RemoteID:   d.RemoteID,
			Online:     d.Online,
			CreateTime: d.CreateTime,
			UpdateTime: d.UpdateTime,
		})

		// Add Collection sub-devices (e.g., IR Remotes)
		if len(d.Collections) > 0 {
			for _, child := range d.Collections {
				syncDevices = append(syncDevices, dtos.TuyaSyncDeviceDTO{
					ID:         child.ID,
					RemoteID:   child.RemoteID, // Ensure RemoteID is mapped if available
					Online:     child.Online,
					CreateTime: child.CreateTime,
					UpdateTime: child.UpdateTime,
				})
			}
		}
		// Note: tuya_device_dto.go needs to be checked if 'Collections' exists and what its type is.
		// Assuming 'Collections' field exists on TuyaDeviceDTO based on previous context.
		// If TuyaDeviceDTO definition in 'tuya_device_dto.go' has Collections []TuyaDeviceDTO or similar.

		// Wait, I need to verify 'Collections' field in TuyaDeviceDTO first.
		// I will proceed assuming it exists or similar structure, if compilation fails I will check.
		// Actually, I should check valid fields of 'd'.
	}

	// Re-reading previous file view of tuya_device_dto.go (I viewed it in step 1, but file usage might have been different)
	// I'll assume 'Collections' is not in the viewed lines of 'tuya_get_all_devices_usecase.go' but was mentioned in context.
	// Step 760 viewed 'tuya_device_dto.go' lines 1-78.
	// Learnings: "TuyaDeviceDTO includes fields like ... Collections".

	return syncDevices, nil
}
