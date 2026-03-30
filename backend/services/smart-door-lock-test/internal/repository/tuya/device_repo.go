package tuya

import (
	"fmt"
	"sensio/backend/services/smart-door-lock-test/internal/domain"
)

// DeviceRepository handles device-related API operations
type DeviceRepository struct {
	client *Client
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(client *Client) *DeviceRepository {
	return &DeviceRepository{client: client}
}

// GetByID retrieves a device by its ID
func (r *DeviceRepository) GetByID(deviceID string) (*domain.Device, error) {
	urlPath := fmt.Sprintf("/v1.0/devices/%s", deviceID)

	respBody, err := r.client.ExecuteRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}

	apiResp, err := ParseResponse(respBody)
	if err != nil {
		return nil, err
	}

	if err := apiResp.CheckError(); err != nil {
		return nil, err
	}

	result, ok := apiResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid result format")
	}

	device := r.mapToDevice(result)
	return device, nil
}

// GetSpecifications retrieves device specifications (functions and status definitions)
func (r *DeviceRepository) GetSpecifications(deviceID string) (*DeviceSpecifications, error) {
	urlPath := fmt.Sprintf("/v1.0/devices/%s/specifications", deviceID)

	respBody, err := r.client.ExecuteRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}

	apiResp, err := ParseResponse(respBody)
	if err != nil {
		return nil, err
	}

	if err := apiResp.CheckError(); err != nil {
		return nil, err
	}

	result, ok := apiResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid result format")
	}

	specs := &DeviceSpecifications{}

	if functions, ok := result["functions"].([]interface{}); ok {
		specs.Functions = r.mapFunctions(functions)
	}

	if status, ok := result["status"].([]interface{}); ok {
		specs.Statuses = r.mapFunctions(status)
	}

	return specs, nil
}

// mapToDevice maps raw API response to domain Device entity
func (r *DeviceRepository) mapToDevice(data map[string]interface{}) *domain.Device {
	device := &domain.Device{
		ID:       getString(data, "id"),
		Name:     getString(data, "name"),
		Category: getString(data, "category"),
		Online:   getBool(data, "online"),
		LocalKey: getString(data, "local_key"),
	}

	if createTime, ok := data["create_time"].(float64); ok {
		device.CreateTime = int64(createTime)
	}

	if updateTime, ok := data["update_time"].(float64); ok {
		device.UpdateTime = int64(updateTime)
	}

	// Map statuses
	if statusArr, ok := data["status"].([]interface{}); ok {
		for _, s := range statusArr {
			sm := s.(map[string]interface{})
			device.Statuses = append(device.Statuses, domain.DeviceStatus{
				Code:  getString(sm, "code"),
				Value: sm["value"],
			})
		}
	}

	return device
}

// mapFunctions maps raw function data to domain entities
func (r *DeviceRepository) mapFunctions(data []interface{}) []domain.DeviceFunction {
	functions := make([]domain.DeviceFunction, 0, len(data))

	for _, f := range data {
		fm := f.(map[string]interface{})
		functions = append(functions, domain.DeviceFunction{
			Code:   getString(fm, "code"),
			Type:   getString(fm, "type"),
			Values: getString(fm, "values"),
		})
	}

	return functions
}

// DeviceSpecifications represents device function and status specifications
type DeviceSpecifications struct {
	Functions []domain.DeviceFunction
	Statuses  []domain.DeviceFunction
}

// Helper functions for type-safe map access

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

// JSON helpers for request building

// BuildCommandRequest builds a command request body
func BuildCommandRequest(commands []domain.Command) map[string]interface{} {
	cmds := make([]map[string]interface{}, 0, len(commands))

	for _, cmd := range commands {
		cmds = append(cmds, map[string]interface{}{
			"code":  cmd.Code,
			"value": cmd.Value,
		})
	}

	return map[string]interface{}{
		"commands": cmds,
	}
}
