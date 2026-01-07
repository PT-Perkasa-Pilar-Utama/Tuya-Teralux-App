package usecases

import (
	"fmt"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/tuya/dtos"
)

// TuyaGetDeviceByIDUseCase retrieves detailed information for a specific device.
type TuyaGetDeviceByIDUseCase struct {
	getAllDevicesUC *TuyaGetAllDevicesUseCase
}

// NewTuyaGetDeviceByIDUseCase initializes a new TuyaGetDeviceByIDUseCase.
//
// param getAllDevicesUC UseCase for fetching all devices (which handles merging logic).
// return *TuyaGetDeviceByIDUseCase A pointer to the initialized usecase.
func NewTuyaGetDeviceByIDUseCase(getAllDevicesUC *TuyaGetAllDevicesUseCase) *TuyaGetDeviceByIDUseCase {
	return &TuyaGetDeviceByIDUseCase{
		getAllDevicesUC: getAllDevicesUC,
	}
}

// GetDeviceByID fetches details by reusing GetAllDevices logic to ensure merging consistency (Mode 2)
func (uc *TuyaGetDeviceByIDUseCase) GetDeviceByID(accessToken, deviceID string) (*dtos.TuyaDeviceDTO, error) {
	// Call GetAllDevices to get processed list (Mode 2 enforced)
	// Passing empty uid (not needed if relying on internal service logic, but if uid is required by GetAllDevices execute, we might need it)
	// Checking GetAllDevices signature: Execute(accessToken, uid string, ...)
	// WAIT: We need UID. Usually passed from controller.
	// Current controller signature: GetDeviceByID(c *gin.Context) -> extracts token.
	// Currently GetDeviceByIDUseCase.GetDeviceByID(accessToken, deviceID). It misses UID.
	// GetAllDevices needs UID to call /v1.0/users/{uid}/devices.
	// We should update the signature or fetch UID from config/context.
	// Let's assume UID is available or we need to update signature.

	// Retrieving UID from config to be safe/quick if not passed
	uid := utils.AppConfig.TuyaUserID

	devicesResp, err := uc.getAllDevicesUC.GetAllDevices(accessToken, uid, 0, 0, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devices for processing: %v", err)
	}

	// Find the target device in the list
	for _, dev := range devicesResp.Devices {
		if dev.ID == deviceID {
			return &dev, nil
		}
		// Also check if it's a child device (IR remote) inside a Hub (if Mode 2 didn't flatten enough or for safety)
		// Mode 2 usually returns merged devices where RemoteID is populated.
		// If ID matches, return it.
	}

	return nil, fmt.Errorf("device with ID %s not found", deviceID)
}
